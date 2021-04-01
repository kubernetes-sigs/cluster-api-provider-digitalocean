/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package computes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"

	corev1 "k8s.io/api/core/v1"
)

// GetDroplet get a droplet instance.
func (s *Service) GetDroplet(id string) (*godo.Droplet, error) {
	if id == "" {
		s.scope.Info("DOMachine does not have an instance id")
		return nil, nil
	}

	s.scope.V(2).Info("Looking for instance by id", "instance-id", id)
	dropletID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse instance id with id %q", id)
	}

	droplet, res, err := s.scope.Droplets.Get(s.ctx, dropletID)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	return droplet, nil
}

// CreateDroplet create a droplet instance.
func (s *Service) CreateDroplet(scope *scope.MachineScope) (*godo.Droplet, error) {
	s.scope.V(2).Info("Creating an instance for a machine")

	bootstrapData, err := scope.GetBootstrapData()
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode bootstrap data")
	}

	clusterName := infrav1.DOSafeName(s.scope.Name())
	instanceName := infrav1.DOSafeName(scope.Name())

	imageID, err := s.GetImageID(scope.DOMachine.Spec.Image)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting image")
	}

	sshkeys := []godo.DropletCreateSSHKey{}
	for _, v := range scope.DOMachine.Spec.SSHKeys {
		keys, err := s.GetSSHKey(v)
		if err != nil {
			return nil, err
		}
		sshkeys = append(sshkeys, godo.DropletCreateSSHKey{
			ID:          keys.ID,
			Fingerprint: keys.Fingerprint,
		})
	}

	volumes := []godo.DropletCreateVolume{}
	for _, disk := range scope.DOMachine.Spec.DataDisks {
		volName := infrav1.DataDiskName(scope.DOMachine, disk.NameSuffix)
		vol, err := s.GetVolumeByName(volName)
		if err != nil {
			return nil, fmt.Errorf("could not get volume to attach to droplet: %w", err)
		}
		if vol == nil {
			return nil, fmt.Errorf("volume %q does not exist", volName)
		}
		volumes = append(volumes, godo.DropletCreateVolume{ID: vol.ID})
	}

	request := &godo.DropletCreateRequest{
		Name:    instanceName,
		Region:  s.scope.Region(),
		Size:    scope.DOMachine.Spec.Size,
		SSHKeys: sshkeys,
		Image: godo.DropletCreateImage{
			ID: imageID,
		},
		UserData:          bootstrapData,
		PrivateNetworking: true,
		Volumes:           volumes,
		VPCUUID:           s.scope.VPC().VPCUUID,
	}

	request.Tags = infrav1.BuildTags(infrav1.BuildTagParams{
		ClusterName: clusterName,
		ClusterUID:  s.scope.UID(),
		Name:        instanceName,
		Role:        scope.Role(),
		Additional:  scope.AdditionalTags(),
	})

	droplet, _, err := s.scope.Droplets.Create(s.ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new droplet")
	}

	return droplet, nil
}

// DeleteDroplet delete a droplet instance.
// Returns nil on success, error in all other cases.
func (s *Service) DeleteDroplet(id string) error {
	s.scope.V(2).Info("Attempting to delete instance", "instance-id", id)
	if id == "" {
		s.scope.Info("Instance does not have an instance id")
		return errors.New("cannot delete instance. instance does not have an instance id")
	}

	dropletID, err := strconv.Atoi(id)
	if err != nil {
		return errors.Wrapf(err, "failed to parse instance id with id %q", id)
	}

	if _, err := s.scope.Droplets.Delete(s.ctx, dropletID); err != nil {
		return errors.Wrapf(err, "failed to delete instance with id %q", id)
	}

	s.scope.V(2).Info("Deleted instance", "instance-id", id)
	return nil
}

// GetDropletAddress convert droplet IPs to corev1.NodeAddresses.
func (s *Service) GetDropletAddress(droplet *godo.Droplet) ([]corev1.NodeAddress, error) {
	addresses := []corev1.NodeAddress{}
	privatev4, err := droplet.PrivateIPv4()
	if err != nil {
		return addresses, err
	}

	addresses = append(addresses, corev1.NodeAddress{
		Type:    corev1.NodeInternalIP,
		Address: privatev4,
	})

	publicv4, err := droplet.PublicIPv4()
	if err != nil {
		return addresses, err
	}

	addresses = append(addresses, corev1.NodeAddress{
		Type:    corev1.NodeExternalIP,
		Address: publicv4,
	})

	return addresses, nil
}
