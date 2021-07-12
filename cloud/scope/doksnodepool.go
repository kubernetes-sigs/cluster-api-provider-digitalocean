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

package scope

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"

	"k8s.io/klog/v2/klogr"
	"k8s.io/utils/pointer"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/cluster-api/controllers/noderefutil"
	expclusterv1 "sigs.k8s.io/cluster-api/exp/api/v1alpha4"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DOKSClusterScopeParams defines the input parameters used to create a new Scope.
type DOKSNodePoolScopeParams struct {
	DOClients
	Client client.Client
	Logger logr.Logger

	Cluster      *clusterv1.Cluster
	DOKSCluster  *infrav1.DOKSCluster
	MachinePool  *expclusterv1.MachinePool
	DOKSNodePool *infrav1.DOKSNodePool
}

// NewDOKSNodePoolScope creates a new DOKSNodePoolScope from the supplied parameters.
// This is meant to be called for each reconcile iteration only on DOKSNodePoolReconciler.
func NewDOKSNodePoolScope(params DOKSNodePoolScopeParams) (*DOKSNodePoolScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("Cluster is required when creating a ClusterScope")
	}
	if params.DOKSCluster == nil {
		return nil, errors.New("DOKSCluster is required when creating a DOKSNodePoolScope")
	}
	if params.MachinePool == nil {
		return nil, errors.New("MachinePool is required when creating a DOKSNodePoolScope")
	}
	if params.DOKSNodePool == nil {
		return nil, errors.New("DOKSNodePool is required when creating a DOKSNodePoolScope")
	}
	if params.Logger == nil {
		params.Logger = klogr.New()
	}

	session, err := params.DOClients.Session()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DO session")
	}

	if params.DOClients.Kubernetes == nil {
		params.DOClients.Kubernetes = session.Kubernetes
	}

	helper, err := patch.NewHelper(params.DOKSNodePool, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &DOKSNodePoolScope{
		Logger:       params.Logger,
		client:       params.Client,
		DOClients:    params.DOClients,
		Cluster:      params.Cluster,
		DOKSCluster:  params.DOKSCluster,
		MachinePool:  params.MachinePool,
		DOKSNodePool: params.DOKSNodePool,
		patchHelper:  helper,
	}, nil
}

// DOKSNodePoolScope defines the basic context for an actuator to operate upon.
type DOKSNodePoolScope struct {
	logr.Logger
	client      client.Client
	patchHelper *patch.Helper

	DOClients
	Cluster      *clusterv1.Cluster
	DOKSCluster  *infrav1.DOKSCluster
	MachinePool  *expclusterv1.MachinePool
	DOKSNodePool *infrav1.DOKSNodePool
}

func (s *DOKSNodePoolScope) ToDOKSClusterScope() (*DOKSClusterScope, error) {
	return NewDOKSClusterScope(DOKSClusterScopeParams{
		DOClients:   s.DOClients,
		Client:      s.client,
		Logger:      s.Logger,
		Cluster:     s.Cluster,
		DOKSCluster: s.DOKSCluster,
	})
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *DOKSNodePoolScope) Close() error {
	return s.patchHelper.Patch(context.TODO(), s.DOKSNodePool)
}

// Name returns the machine pool name.
func (s *DOKSNodePoolScope) Name() string {
	return s.MachinePool.GetName()
}

// Replicas returns the machine pool replicas compatible to godo format.
func (s *DOKSNodePoolScope) Replicas() *int {
	replicas := *s.MachinePool.Spec.Replicas

	var conv int
	conv = int(replicas)

	return &conv
}

// Namespace returns the cluster namespace.
func (s *DOKSNodePoolScope) Namespace() string {
	return s.Cluster.GetNamespace()
}

// SetReady sets the DOKSNodePool Ready Status.
func (s *DOKSNodePoolScope) SetReady() {
	s.DOKSNodePool.Status.Ready = true
}

// GetProviderID returns the DOKSNodePool providerID from the spec.
func (s *DOKSNodePoolScope) GetProviderID() string {
	if s.DOKSNodePool.Spec.ProviderID != nil {
		return *s.DOKSNodePool.Spec.ProviderID
	}
	return ""
}

// SetProviderID sets the DOKSNodePool providerID in spec from cluster id.
func (s *DOKSNodePoolScope) SetProviderID(nodePoolID string) {
	pid := fmt.Sprintf("digitalocean://%s", nodePoolID)
	s.DOKSNodePool.Spec.ProviderID = pointer.StringPtr(pid)
}

// SetProviderIDList sets the DOKSNodePool providerIDList in spec from nodes.
func (s *DOKSNodePoolScope) SetProviderIDList(nodes []*godo.KubernetesNode) {
	providerIDList := make([]string, 0, len(nodes))
	for _, node := range nodes {
		pid := fmt.Sprintf("digitalocean://%s", node.ID)
		providerIDList = append(providerIDList, pid)
	}
	s.DOKSNodePool.Spec.ProviderIDList = providerIDList
}

// GetInstanceID returns the DOKSNodePool intsance id by parsing Spec.ProviderID.
func (s *DOKSNodePoolScope) GetInstanceID() string {
	parsed, err := noderefutil.NewProviderID(s.GetProviderID())
	if err != nil {
		return ""
	}
	return parsed.ID()
}

func (s *DOKSNodePoolScope) DOAPICreateRequest() *godo.KubernetesNodePoolCreateRequest {
	return &godo.KubernetesNodePoolCreateRequest{
		Name:  s.Name(),
		Size:  s.DOKSNodePool.Spec.Size,
		Count: int(*s.MachinePool.Spec.Replicas),
	}
}
