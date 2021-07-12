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

package kubernetes

import (
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"

	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
)

// GetCluster get a cluster instance.
func (s *Service) GetNodePool(clusterID string, nodePoolID string) (*godo.KubernetesNodePool, error) {
	if nodePoolID == "" {
		s.scope.Info("DOKSNodePool does not have an instance id")
		return nil, nil
	}

	s.scope.V(2).Info("Looking for instance by id", "instance-id", nodePoolID)

	nodePool, res, err := s.scope.Kubernetes.GetNodePool(s.ctx, clusterID, nodePoolID)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	return nodePool, nil
}

// CreateNodePool create a node pool instance.
func (s *Service) CreateNodePool(clusterID string, nodePoolScope *scope.DOKSNodePoolScope) (*godo.KubernetesNodePool, error) {
	s.scope.V(2).Info("Creating an instance for a cluster")

	nodePoolCreateRequest := nodePoolScope.DOAPICreateRequest()
	nodePool, _, err := s.scope.Kubernetes.CreateNodePool(s.ctx, clusterID, nodePoolCreateRequest)

	return nodePool, err
}

// DeleteNodePool delete a node pool instance.
// Returns nil on success, error in all other cases.
func (s *Service) DeleteNodePool(clusterID string, nodePoolID string) error {
	s.scope.V(2).Info("Attempting to delete instance", "instance-id", nodePoolID)
	if nodePoolID == "" {
		s.scope.Info("Instance does not have an instance id")
		return errors.New("cannot delete instance. instance does not have an instance id")
	}

	if _, err := s.scope.Kubernetes.DeleteNodePool(s.ctx, clusterID, nodePoolID); err != nil {
		return errors.Wrapf(err, "failed to delete instance with id %q", nodePoolID)
	}

	s.scope.V(2).Info("Deleted instance", "instance-id", nodePoolID)
	return nil
}
