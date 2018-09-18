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
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"

	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/kubeadm"
	apiutil "sigs.k8s.io/cluster-api/pkg/util"

	"github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/actuators/machine/machinesetup"
	doconfigv1 "github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/providerconfig/v1alpha1"
	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/ssh"
	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/sshutil"
	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/util"

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

type DOClientMachineSetupConfigGetter interface {
	GetMachineSetupConfig() (machinesetup.MachineSetupConfig, error)
}

// DOClientSSHCreds has path to the private key and user associated with it.
type DOClientSSHCreds struct {
	privateKeyPath string
	publicKeyPath  string
	user           string
}

// DOClient is responsible for performing machine reconciliation
type DOClient struct {
	godoClient               *godo.Client
	scheme                   *runtime.Scheme
	doProviderConfigCodec    *doconfigv1.DigitalOceanProviderConfigCodec
	kubeadm                  DOClientKubeadm
	ctx                      context.Context
	SSHCreds                 DOClientSSHCreds
	v1Alpha1Client           client.ClusterV1alpha1Interface
	eventRecorder            record.EventRecorder
	machineSetupConfigGetter DOClientMachineSetupConfigGetter
}

// ActuatorParams holds parameter information for DOClient
type ActuatorParams struct {
	Kubeadm                  DOClientKubeadm
	V1Alpha1Client           client.ClusterV1alpha1Interface
	EventRecorder            record.EventRecorder
	MachineSetupConfigGetter DOClientMachineSetupConfigGetter
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

	var user, privateKeyPath, publicKeyPath string
	if _, err := os.Stat("/etc/sshkeys/private"); err == nil {
		privateKeyPath = "/etc/sshkeys/private"

		// TODO: A PR is coming for this. We will match images to OSes. This will be also needed for userdata.
		user = "root"
	}
	if _, err := os.Stat("/etc/sshkeys/public"); err == nil {
		publicKeyPath = "/etc/sshkeys/public"
	}

	return &DOClient{
		godoClient:            getGodoClient(),
		scheme:                scheme,
		doProviderConfigCodec: codec,
		kubeadm:               getKubeadm(params),
		ctx:                   context.Background(),
		SSHCreds: DOClientSSHCreds{
			privateKeyPath: privateKeyPath,
			publicKeyPath:  publicKeyPath,
			user:           user,
		},
		v1Alpha1Client:           params.V1Alpha1Client,
		eventRecorder:            params.EventRecorder,
		machineSetupConfigGetter: params.MachineSetupConfigGetter,
	}, nil
}

// Create creates a machine and is invoked by the Machine Controller
func (do *DOClient) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	if do.machineSetupConfigGetter == nil {
		return fmt.Errorf("machine setup config is required")
	}

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

	token, err := do.getKubeadmToken()
	if err != nil {
		return err
	}

	var parsedMetadata string
	configParams := &machinesetup.ConfigParams{
		Image:    machineConfig.Image,
		Versions: machine.Spec.Versions,
	}
	machineSetupConfig, err := do.machineSetupConfigGetter.GetMachineSetupConfig()
	if err != nil {
		return err
	}
	metadata, err := machineSetupConfig.GetUserdata(configParams)
	if err != nil {
		return err
	}
	if util.IsMachineMaster(machine) {
		parsedMetadata, err = masterUserdata(cluster, machine, machineConfig.Image, token, metadata)
		if err != nil {
			return err
		}
	} else {
		parsedMetadata, err = nodeUserdata(cluster, machine, machineConfig.Image, token, metadata)
		if err != nil {
			return err
		}
	}

	dropletSSHKeys := []godo.DropletCreateSSHKey{}
	// Add machineSpec provided keys.
	for _, k := range machineConfig.SSHPublicKeys {
		sshkey, err := sshutil.NewKeyFromString(k)
		if err != nil {
			return err
		}
		if err := sshkey.Create(do.ctx, do.godoClient.Keys); err != nil {
			return err
		}
		dropletSSHKeys = append(dropletSSHKeys, godo.DropletCreateSSHKey{Fingerprint: sshkey.FingerprintMD5})
	}
	// Add machineActuator public key.
	if do.SSHCreds.publicKeyPath != "" {
		sshkey, err := sshutil.NewKeyFromFile(do.SSHCreds.publicKeyPath)
		if err != nil {
			return err
		}
		if err := sshkey.Create(do.ctx, do.godoClient.Keys); err != nil {
			return err
		}
		dropletSSHKeys = append(dropletSSHKeys, godo.DropletCreateSSHKey{Fingerprint: sshkey.FingerprintMD5})
	}

	dropletCreateReq := &godo.DropletCreateRequest{
		Name:   machine.Name,
		Region: machineConfig.Region,
		Size:   machineConfig.Size,
		Image: godo.DropletCreateImage{
			Slug: machineConfig.Image,
		},
		Backups:           machineConfig.Backups,
		IPv6:              machineConfig.IPv6,
		PrivateNetworking: machineConfig.PrivateNetworking,
		Monitoring:        machineConfig.Monitoring,
		Tags: append([]string{
			string(machine.UID),
		}, machineConfig.Tags...),
		SSHKeys:  dropletSSHKeys,
		UserData: parsedMetadata,
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
		if sets.NewString(droplet.Tags...).Has(string(machine.UID)) && len(droplet.Networks.V4) > 0 {
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
	err = do.updateInstanceStatus(machine)
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
func (do *DOClient) Update(cluster *clusterv1.Cluster, goalMachine *clusterv1.Machine) error {
	goalMachineConfig, err := do.decodeMachineProviderConfig(goalMachine.Spec.ProviderConfig)
	if err != nil {
		return fmt.Errorf("error decoding provided machineConfig: %v", err)
	}

	if err := do.validateMachine(goalMachineConfig); err != nil {
		return fmt.Errorf("error validating provided machineConfig: %v", err)
	}

	droplet, err := do.instanceExists(goalMachine)
	if err != nil {
		return err
	}
	if droplet == nil {
		return fmt.Errorf("machine %s doesn't exist", goalMachine.Name)
	}

	status, err := do.instanceStatus(goalMachine)
	if err != nil {
		return err
	}

	currentMachine := (*clusterv1.Machine)(status)
	if currentMachine == nil {
		return fmt.Errorf("status annotation not set")
	}

	if !do.requiresUpdate(currentMachine, goalMachine) {
		return nil
	}

	if util.IsMachineMaster(currentMachine) {
		if currentMachine.Spec.Versions.ControlPlane != goalMachine.Spec.Versions.ControlPlane {
			cmd, err := do.upgradeCommandMasterControlPlane(goalMachine)
			if err != nil {
				return err
			}

			sshClient, err := ssh.NewClient(cluster.Status.APIEndpoints[0].Host, "22", do.SSHCreds.user, do.SSHCreds.privateKeyPath)
			if err != nil {
				return err
			}
			err = sshClient.Connect()
			if err != nil {
				return err
			}

			for _, c := range cmd {
				_, err = sshClient.Execute(c)
				if err != nil {
					return err
				}
			}
			err = do.updateInstanceStatus(goalMachine)
			if err != nil {
				return err
			}
			err = sshClient.Close()
			if err != nil {
				return err
			}
		}

		if currentMachine.Spec.Versions.Kubelet != goalMachine.Spec.Versions.Kubelet {
			cmd, err := do.upgradeCommandMasterKubelet(goalMachine)
			if err != nil {
				return err
			}

			sshClient, err := ssh.NewClient(cluster.Status.APIEndpoints[0].Host, "22", do.SSHCreds.user, do.SSHCreds.privateKeyPath)
			if err != nil {
				return err
			}
			err = sshClient.Connect()
			if err != nil {
				return err
			}

			for _, c := range cmd {
				_, err = sshClient.Execute(c)
				if err != nil {
					return err
				}
			}

			err = do.updateInstanceStatus(goalMachine)
			if err != nil {
				return err
			}
			err = sshClient.Close()
			if err != nil {
				return err
			}
		}
	} else {
		glog.Infof("re-creating node %s for update", currentMachine.Name)
		err = do.Delete(cluster, currentMachine)
		if err != nil {
			return err
		}

		goalMachine.Annotations[IDAnnotationKey] = ""
		err = do.Create(cluster, goalMachine)
		if err != nil {
			return err
		}
	}

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
	if len(droplet.Networks.V4) == 0 {
		return "", fmt.Errorf("instance %s doesn't have IP address assigned", droplet.Name)
	}
	return droplet.Networks.V4[0].IPAddress, nil
}

// GetKubeConfig returns kubeconfig from the master.
func (do *DOClient) GetKubeConfig(cluster *clusterv1.Cluster, master *clusterv1.Machine) (string, error) {
	droplet, err := do.instanceExists(master)
	if err != nil {
		return "", err
	}
	if droplet == nil {
		return "", fmt.Errorf("instance does not exists")
	}
	if len(cluster.Status.APIEndpoints) == 0 {
		return "", fmt.Errorf("unable to find cluster api endpoint address")
	}

	// We're using system SSH to download kubeconfig file from master.
	// The reasons for that are:
	//   * GetKubeConfig is executed by clusterctl on local machine. We don't know location of SSH key used for
	//     authentication and we can't add clusterctl flag for SSH key path, as it may not be possible to pass it to
	//     machine-controller and GetKubeConfig function.
	//   * Because we don't have SSH key, we can't use Go SSH implementation here.
	//   * Using Go SSH implementation and SSH agent together misbehaves, so is not an option.
	// Therefore, we're using system SSH here, so it correctly handles keys and authentication.
	result := strings.TrimSpace(apiutil.ExecCommand(
		"ssh", "-q",
		"-o", "StrictHostKeyChecking no",
		"-o", "UserKnownHostsFile /dev/null",
		fmt.Sprintf("%s@%s", "root", cluster.Status.APIEndpoints[0].Host),
		"echo STARTFILE; sudo cat /etc/kubernetes/admin.conf"))
	kubeconfig := strings.Split(result, "STARTFILE")

	if len(kubeconfig) < 2 {
		return "", fmt.Errorf("kubeconfig not available")
	}

	return strings.TrimSpace(kubeconfig[1]), nil
}

func getKubeadm(params ActuatorParams) DOClientKubeadm {
	if params.Kubeadm == nil {
		return kubeadm.New()
	}
	return params.Kubeadm
}

func (do *DOClient) getKubeadmToken() (string, error) {
	tokenParams := kubeadm.TokenCreateParams{
		Ttl: time.Duration(30) * time.Minute,
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
		if strID == "" {
			return nil, nil
		}
		id, err := strconv.Atoi(strID)
		if err != nil {
			return nil, err
		}
		droplet, resp, err := do.godoClient.Droplets.Get(do.ctx, id)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				// Machine exists as an object, but Droplet is already deleted.
				return nil, nil
			}
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

// requiresUpdate compares ObjectMeta, ProviderConfig and Versions object of two machines.
func (do *DOClient) requiresUpdate(a *clusterv1.Machine, b *clusterv1.Machine) bool {
	// Do not want status changes. Do want changes that impact machine provisioning
	return !equality.Semantic.DeepEqual(a.Spec.ObjectMeta, b.Spec.ObjectMeta) ||
		!equality.Semantic.DeepEqual(a.Spec.ProviderConfig, b.Spec.ProviderConfig) ||
		!equality.Semantic.DeepEqual(a.Spec.Versions, b.Spec.Versions)
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
