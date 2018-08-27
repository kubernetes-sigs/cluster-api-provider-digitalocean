// Copyright © 2018 The Kubernetes Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package machine

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"

	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/kubeadm"

	"github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/actuators/machine/userdata"
	doconfigv1 "github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/providerconfig/v1alpha1"

	"github.com/digitalocean/godo"
	"github.com/golang/glog"
)

const (
	ProviderName = "digitalocean"

	createCheckPeriod  = 10 * time.Second
	createCheckTimeout = 5 * time.Minute

	eventReasonCreate = "Create"
	eventReasonUpdate = "Update"
	eventReasonDelete = "Delete"

	NameAnnotationKey   = "droplet-name"
	IDAnnotationKey     = "droplet-id"
	RegionAnnotationKey = "droplet-region"
)

func init() {
	actuator, err := NewMachineActuator(ActuatorParams{})
	if err != nil {
		glog.Fatalf("Error creating cluster provisioner for %v : %v", ProviderName, err)
	}
	clustercommon.RegisterClusterProvisioner(ProviderName, actuator)
}

type DOClientKubeadm interface {
	TokenCreate(params kubeadm.TokenCreateParams) (string, error)
}

// DOClient is responsible for performing machine reconciliation
type DOClient struct {
	godoClient            *godo.Client
	scheme                *runtime.Scheme
	doProviderConfigCodec *doconfigv1.DigitalOceanProviderConfigCodec
	kubeadm               DOClientKubeadm
	ctx                   context.Context
	v1Alpha1Client        client.ClusterV1alpha1Interface
	eventRecorder         record.EventRecorder
}

// ActuatorParams holds parameter information for DOClient
type ActuatorParams struct {
	Kubeadm        DOClientKubeadm
	V1Alpha1Client client.ClusterV1alpha1Interface
	EventRecorder  record.EventRecorder
}

// NewMachineActuator creates a new DOClient
func NewMachineActuator(params ActuatorParams) (*DOClient, error) {
	scheme, err := doconfigv1.NewScheme()
	if err != nil {
		return nil, err
	}

	codec, err := doconfigv1.NewCodec()
	if err != nil {
		return nil, err
	}

	return &DOClient{
		godoClient:            getGodoClient(),
		scheme:                scheme,
		doProviderConfigCodec: codec,
		kubeadm:               getKubeadm(params),
		ctx:                   context.Background(),
		v1Alpha1Client:        params.V1Alpha1Client,
		eventRecorder:         params.EventRecorder,
	}, nil
}

// Create creates a machine and is invoked by the Machine Controller
func (do *DOClient) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	machineConfig, err := do.decodeMachineProviderConfig(machine.Spec.ProviderConfig)
	if err != nil {
		return fmt.Errorf("error decoding provided machineConfig: %v", err)
	}

	if err := do.validateMachine(machineConfig); err != nil {
		return fmt.Errorf("error validating provided machineConfig: %v", err)
	}

	droplet, err := do.instanceExists(machine)
	if err != nil {
		return err
	}
	if droplet != nil {
		glog.Info("Skipping the machine that already exists.")
		return nil
	}

	metadataProvider, err := userdata.ForOS(machineConfig.Image)
	if err != nil {
		return err
	}
	kubeadmToken, err := do.getKubeadmToken()
	if err != nil {
		return err
	}
	metadata, err := metadataProvider.UserData(cluster, machine, machineConfig, kubeadmToken)
	if err != nil {
		return err
	}

	// TODO: Handle SSH keys.
	// Metadata handlers can be ported from the machine-controller.
	dropletCreateReq := &godo.DropletCreateRequest{
		Name:   machine.Name,
		Region: machineConfig.Region,
		Size:   machineConfig.Size,
		Image: godo.DropletCreateImage{
			Slug: machineConfig.Image,
		},
		Backups:           machineConfig.Backups,
		IPv6:              machineConfig.IPv6,
		PrivateNetworking: machineConfig.IPv6,
		Monitoring:        machineConfig.Monitoring,
		Tags: append([]string{
			string(machine.UID),
		}, machineConfig.Tags...),
		UserData: metadata,
	}

	droplet, _, err = do.godoClient.Droplets.Create(do.ctx, dropletCreateReq)
	if err != nil {
		return err
	}

	//We need to wait until the droplet really got created as tags will be only applied when the droplet is running.
	err = wait.Poll(createCheckPeriod, createCheckTimeout, func() (done bool, err error) {
		droplet, _, err := do.godoClient.Droplets.Get(do.ctx, droplet.ID)
		if err != nil {
			return false, err
		}
		if sets.NewString(droplet.Tags...).Has(string(machine.UID)) {
			return true, nil
		}
		glog.Infof("waiting until machine %s gets fully created", machine.Name)
		return false, nil
	})

	if machine.ObjectMeta.Annotations == nil {
		machine.ObjectMeta.Annotations = map[string]string{}
	}
	machine.ObjectMeta.Annotations[NameAnnotationKey] = droplet.Name
	machine.ObjectMeta.Annotations[IDAnnotationKey] = strconv.Itoa(droplet.ID)
	machine.ObjectMeta.Annotations[RegionAnnotationKey] = droplet.Region.Name

	_, err = do.v1Alpha1Client.Machines(machine.Namespace).Update(machine)
	if err != nil {
		return err
	}

	do.eventRecorder.Eventf(machine, corev1.EventTypeNormal, eventReasonCreate, "machine %s successfully created", machine.ObjectMeta.Name)
	return nil
}

// Delete deletes a machine and is invoked by the Machine Controller
func (do *DOClient) Delete(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	droplet, err := do.instanceExists(machine)
	if err != nil {
		return err
	}
	if droplet == nil {
		glog.Info("Skipping the machine that doesn't exist.")
		return nil
	}

	_, err = do.godoClient.Droplets.Delete(do.ctx, droplet.ID)
	if err != nil {
		return err
	}

	do.eventRecorder.Eventf(machine, corev1.EventTypeNormal, eventReasonDelete, "machine %s successfully deleted", machine.ObjectMeta.Name)
	return nil
}

// Update updates a machine and is invoked by the Machine Controller
func (do *DOClient) Update(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	machineConfig, err := do.decodeMachineProviderConfig(machine.Spec.ProviderConfig)
	if err != nil {
		return fmt.Errorf("error decoding provided machineConfig: %v", err)
	}

	if err := do.validateMachine(machineConfig); err != nil {
		return fmt.Errorf("error validating provided machineConfig: %v", err)
	}

	droplet, err := do.instanceExists(machine)
	if err != nil {
		return err
	}
	if droplet != nil {
		return fmt.Errorf("machine %s doesn't exist", machine.Name)
	}

	// Check is machine a Master
	if machine.Spec.Versions.ControlPlane != "" {
		// If machine is a master, we need to do in-place upgrade.
		// TODO: In-place upgrade implementation.
		return fmt.Errorf("TODO: Not yet implemented")
	}
	// If machine is a node, we can remove it and add a new one.
	// TODO: This need to work better: drain node, create a new one, delete old one.
	// TODO: A node must wait successful creation before proceeding. Wait implementation TODO.
	// TODO: Problem: how we can create machine before we delete it? They are going to have same name.
	err = do.Delete(cluster, machine)
	if err != nil {
		return err
	}
	err = do.Create(cluster, machine)
	if err != nil {
		return err
	}

	do.eventRecorder.Eventf(machine, corev1.EventTypeNormal, eventReasonUpdate, "machine %s successfully updated", machine.ObjectMeta.Name)
	return nil
}

// Exists test for the existance of a machine and is invoked by the Machine Controller
func (do *DOClient) Exists(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	droplet, err := do.instanceExists(machine)
	if err != nil {
		return false, err
	}
	if droplet != nil {
		return true, nil
	}
	return false, nil
}

// GetIP returns public IP address of the node in the cluster.
func (do *DOClient) GetIP(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (string, error) {
	droplet, err := do.instanceExists(machine)
	if err != nil {
		return "", err
	}
	if droplet == nil {
		return "", fmt.Errorf("instance %s doesn't exist", droplet.Name)
	}
	return droplet.Networks.V4[0].IPAddress, nil
}

// GetKubeConfig returns kubeconfig from the master.
func (do *DOClient) GetKubeConfig(cluster *clusterv1.Cluster, master *clusterv1.Machine) (string, error) {
	// TODO: Implement getKubeConfig. Possibly using SSH/SCP.
	return "", fmt.Errorf("TODO: Not yet implemented")
}

func getKubeadm(params ActuatorParams) DOClientKubeadm {
	if params.Kubeadm == nil {
		return kubeadm.New()
	}
	return params.Kubeadm
}

func (do *DOClient) getKubeadmToken() (string, error) {
	tokenParams := kubeadm.TokenCreateParams{
		Ttl: time.Duration(10) * time.Minute,
	}

	token, err := do.kubeadm.TokenCreate(tokenParams)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(token), nil
}

// instanceExists returns instance with provided name if it already exists in the cloud.
func (do *DOClient) instanceExists(machine *clusterv1.Machine) (*godo.Droplet, error) {
	if strID, ok := machine.ObjectMeta.Annotations[IDAnnotationKey]; ok {
		id, err := strconv.Atoi(strID)
		if err != nil {
			return nil, err
		}
		droplet, _, err := do.godoClient.Droplets.Get(do.ctx, id)
		if err != nil {
			return nil, err
		}
		if droplet != nil {
			return droplet, nil
		}
		// Fallback to searching by name.
	}

	droplets, _, err := do.godoClient.Droplets.List(do.ctx, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, d := range droplets {
		if d.Name == machine.Name && sets.NewString(d.Tags...).Has(string(machine.UID)) {
			glog.Infof("found a machine %s by name", d.Name)
			return &d, nil
		}
	}
	return nil, nil
}

func (do *DOClient) validateMachine(providerConfig *doconfigv1.DigitalOceanMachineProviderConfig) error {
	if len(providerConfig.Image) == 0 {
		return fmt.Errorf("image slug must be provided")
	}
	if len(providerConfig.Region) == 0 {
		return fmt.Errorf("region must be provided")
	}
	if len(providerConfig.Size) == 0 {
		return fmt.Errorf("size must be provided")
	}

	return nil
}

// decodeMachineProviderConfig returns DigitalOcean MachineProviderConfig from upstream Spec.
func (do *DOClient) decodeMachineProviderConfig(providerConfig clusterv1.ProviderConfig) (*doconfigv1.DigitalOceanMachineProviderConfig, error) {
	var config doconfigv1.DigitalOceanMachineProviderConfig
	err := do.doProviderConfigCodec.DecodeFromProviderConfig(providerConfig, &config)
	if err != nil {
		return nil, err
	}

	return &config, err
}
