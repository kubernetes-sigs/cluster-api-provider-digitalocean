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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha3"

	"k8s.io/klog/klogr"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	DOClients
	Client    client.Client
	Logger    logr.Logger
	Cluster   *clusterv1.Cluster
	DOCluster *infrav1.DOCluster
}

// NewClusterScope creates a new ClusterScope from the supplied parameters.
// This is meant to be called for each reconcile iteration only on DOClusterReconciler.
func NewClusterScope(params ClusterScopeParams) (*ClusterScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("Cluster is required when creating a ClusterScope")
	}
	if params.DOCluster == nil {
		return nil, errors.New("DOCluster is required when creating a ClusterScope")
	}
	if params.Logger == nil {
		params.Logger = klogr.New()
	}

	session, err := params.DOClients.Session()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DO session")
	}

	if params.DOClients.Actions == nil {
		params.DOClients.Actions = session.Actions
	}

	if params.DOClients.Droplets == nil {
		params.DOClients.Droplets = session.Droplets
	}

	if params.DOClients.Storage == nil {
		params.DOClients.Storage = session.Storage
	}

	if params.DOClients.Images == nil {
		params.DOClients.Images = session.Images
	}

	if params.DOClients.Keys == nil {
		params.DOClients.Keys = session.Keys
	}

	if params.DOClients.LoadBalancers == nil {
		params.DOClients.LoadBalancers = session.LoadBalancers
	}

	helper, err := patch.NewHelper(params.DOCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &ClusterScope{
		Logger:      params.Logger,
		client:      params.Client,
		DOClients:   params.DOClients,
		Cluster:     params.Cluster,
		DOCluster:   params.DOCluster,
		patchHelper: helper,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	logr.Logger
	client      client.Client
	patchHelper *patch.Helper

	DOClients
	Cluster   *clusterv1.Cluster
	DOCluster *infrav1.DOCluster
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *ClusterScope) Close() error {
	return s.patchHelper.Patch(context.TODO(), s.DOCluster)
}

// Name returns the cluster name.
func (s *ClusterScope) Name() string {
	return s.Cluster.GetName()
}

// Namespace returns the cluster namespace.
func (s *ClusterScope) Namespace() string {
	return s.Cluster.GetNamespace()
}

func (s *ClusterScope) UID() string {
	return string(s.Cluster.UID)
}

// Region returns the cluster region.
func (s *ClusterScope) Region() string {
	return s.DOCluster.Spec.Region
}

// Network returns the cluster network object.
func (s *ClusterScope) Network() *infrav1.DONetworkResource {
	return &s.DOCluster.Status.Network
}

// SetReady sets the DOCluster Ready Status.
func (s *ClusterScope) SetReady() {
	s.DOCluster.Status.Ready = true
}

// SetControlPlaneEndpoint sets the DOCluster status APIEndpoints.
func (s *ClusterScope) SetControlPlaneEndpoint(apiEndpoint clusterv1.APIEndpoint) {
	s.DOCluster.Spec.ControlPlaneEndpoint = apiEndpoint
}

// APIServerLoadbalancers get the DOCluster Spec Network APIServerLoadbalancers.
func (s *ClusterScope) APIServerLoadbalancers() *infrav1.DOLoadBalancer {
	return &s.DOCluster.Spec.Network.APIServerLoadbalancers
}

// APIServerLoadbalancersRef get the DOCluster status Network APIServerLoadbalancersRef.
func (s *ClusterScope) APIServerLoadbalancersRef() *infrav1.DOResourceReference {
	return &s.DOCluster.Status.Network.APIServerLoadbalancersRef
}

// VPC gets the DOCluster Spec Network VPC.
func (s *ClusterScope) VPC() *infrav1.DOVPC {
	return &s.DOCluster.Spec.Network.VPC
}
