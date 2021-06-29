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
type DOKSClusterScopeParams struct {
	DOClients
	Client        client.Client
	Logger        logr.Logger
	Cluster       *clusterv1.Cluster
	DOKSCluster   *infrav1.DOKSCluster
	MachinePools  []*expclusterv1.MachinePool
	DOKSNodePools []*infrav1.DOKSNodePool
}

// NewDOKSClusterScope creates a new DOKSClusterScope from the supplied parameters.
// This is meant to be called for each reconcile iteration only on DOKSClusterReconciler.
func NewDOKSClusterScope(params DOKSClusterScopeParams) (*DOKSClusterScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("Cluster is required when creating a ClusterScope")
	}
	if params.DOKSCluster == nil {
		return nil, errors.New("DOKSCluster is required when creating a DOKSClusterScope")
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

	helper, err := patch.NewHelper(params.DOKSCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &DOKSClusterScope{
		Logger:        params.Logger,
		client:        params.Client,
		DOClients:     params.DOClients,
		Cluster:       params.Cluster,
		DOKSCluster:   params.DOKSCluster,
		MachinePools:  params.MachinePools,
		DOKSNodePools: params.DOKSNodePools,
		patchHelper:   helper,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type DOKSClusterScope struct {
	logr.Logger
	client      client.Client
	patchHelper *patch.Helper

	DOClients
	Cluster       *clusterv1.Cluster
	DOKSCluster   *infrav1.DOKSCluster
	MachinePools  []*expclusterv1.MachinePool
	DOKSNodePools []*infrav1.DOKSNodePool
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *DOKSClusterScope) Close() error {
	return s.patchHelper.Patch(context.TODO(), s.DOKSCluster)
}

// Name returns the cluster name.
func (s *DOKSClusterScope) Name() string {
	return s.Cluster.GetName()
}

// Namespace returns the cluster namespace.
func (s *DOKSClusterScope) Namespace() string {
	return s.Cluster.GetNamespace()
}

// Region returns the cluster region.
func (s *DOKSClusterScope) Region() string {
	return s.DOKSCluster.Spec.Region
}

// SetReady sets the DOKSCluster Ready Status.
func (s *DOKSClusterScope) SetReady() {
	s.DOKSCluster.Status.Ready = true
}

// GetProviderID returns the DOKSCluster providerID from the spec.
func (s *DOKSClusterScope) GetProviderID() string {
	if s.DOKSCluster.Spec.ProviderID != nil {
		return *s.DOKSCluster.Spec.ProviderID
	}
	return ""
}

// SetProviderID sets the DOKSCluster providerID in spec from cluster id.
func (s *DOKSClusterScope) SetProviderID(clusterID string) {
	pid := fmt.Sprintf("digitalocean://%s", clusterID)
	s.DOKSCluster.Spec.ProviderID = pointer.StringPtr(pid)
}

// GetInstanceID returns the DOKSCluster cluster intsance id by parsing Spec.ProviderID.
func (s *DOKSClusterScope) GetInstanceID() string {
	parsed, err := noderefutil.NewProviderID(s.GetProviderID())
	if err != nil {
		return ""
	}
	return parsed.ID()
}
