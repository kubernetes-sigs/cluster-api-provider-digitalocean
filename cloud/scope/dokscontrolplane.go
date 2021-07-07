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

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"

	"k8s.io/klog/v2/klogr"

	controlplanev1 "sigs.k8s.io/cluster-api-provider-digitalocean/controlplane/doks/api/v1alpha4"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DOKSControlPlaneScopeParams defines the input parameters used to create a new Scope.
type DOKSControlPlaneScopeParams struct {
	DOClients
	Client           client.Client
	Logger           logr.Logger
	Cluster          *clusterv1.Cluster
	DOKSCluster      *infrav1.DOKSCluster
	DOKSControlPlane *controlplanev1.DOKSControlPlane
}

// NewDOKSControlPlaneScope creates a new DOKSControlPlaneScope from the supplied parameters.
// This is meant to be called for each reconcile iteration only on DOKSClusterReconciler.
func NewDOKSControlPlaneScope(params DOKSControlPlaneScopeParams) (*DOKSControlPlaneScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("Cluster is required when creating a ControlPlaneScope")
	}
	if params.DOKSCluster == nil {
		return nil, errors.New("DOKSCluster is required when creating a DOKSControlPlaneScope")
	}
	if params.DOKSControlPlane == nil {
		return nil, errors.New("DOKSControlPlane is required when creating a DOKSControlPlaneScope")
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

	return &DOKSControlPlaneScope{
		Logger:           params.Logger,
		client:           params.Client,
		DOClients:        params.DOClients,
		Cluster:          params.Cluster,
		DOKSCluster:      params.DOKSCluster,
		DOKSControlPlane: params.DOKSControlPlane,
		patchHelper:      helper,
	}, nil
}

// DOKSControlPlaneScope defines the basic context for an actuator to operate upon.
type DOKSControlPlaneScope struct {
	logr.Logger
	client      client.Client
	patchHelper *patch.Helper

	DOClients
	Cluster          *clusterv1.Cluster
	DOKSCluster      *infrav1.DOKSCluster
	DOKSControlPlane *controlplanev1.DOKSControlPlane
}

func (s *DOKSControlPlaneScope) ToDOKSClusterScope() (*DOKSClusterScope, error) {
	return NewDOKSClusterScope(DOKSClusterScopeParams{
		DOClients:   s.DOClients,
		Client:      s.client,
		Logger:      s.Logger,
		Cluster:     s.Cluster,
		DOKSCluster: s.DOKSCluster,
	})
}

// PatchObject persists the control plane configuration and status.
func (s *DOKSControlPlaneScope) PatchObject() error {
	return s.patchHelper.Patch(context.TODO(), s.DOKSControlPlane)
}

// Close closes the current scope persisting the control plane configuration and status.
func (s *DOKSControlPlaneScope) Close() error {
	return s.PatchObject()
}

// Name returns the control plane name.
func (s *DOKSControlPlaneScope) Name() string {
	return s.Cluster.GetName()
}

// Namespace returns the control plane namespace.
func (s *DOKSControlPlaneScope) Namespace() string {
	return s.Cluster.GetNamespace()
}

// SetInitialized sets the DOKSControlPlane Initialized Status.
func (s *DOKSControlPlaneScope) SetInitialized() {
	s.DOKSControlPlane.Status.Initialized = true
}

// SetReady sets the DOKSControlPlane Ready Status.
func (s *DOKSControlPlaneScope) SetReady() {
	s.DOKSControlPlane.Status.Ready = true
}

// SetNotReady sets the DOKSControlPlane Ready Status.
func (s *DOKSControlPlaneScope) SetNotReady() {
	s.DOKSControlPlane.Status.Ready = false
}
