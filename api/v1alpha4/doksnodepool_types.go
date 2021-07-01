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

package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/errors"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// DOKSNodePoolFinalizer allows ReconcileDOKSNodePool to clean up DigitalOcean resources associated
	// with DOKSNodePool before removing it from the apiserver.
	DOKSNodePoolFinalizer = "doksnodepool.infrastructure.cluster.x-k8s.io"
)

// DOKSNodePoolSpec defines the desired state of DOKSNodePool
type DOKSNodePoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The slug identifier for the type of Droplet to be used as workers in the node pool.
	Size string `json:"size"`

	// A boolean value indicating whether auto-scaling is enabled for this node pool.
	// This requires DOKS versions at least 1.13.10-do.3, 1.14.6-do.3, or 1.15.3-do.3.
	AutoScale bool `json:"auto_scale,omitempty"`
	// The minimum number of nodes that this node pool can be auto-scaled to.
	// This will fail validation if the additional nodes will exceed your account droplet limit.
	MinNodes *int `json:"min_nodes,omitempty"`
	// The maximum number of nodes that this node pool can be auto-scaled to.
	// This can be 0, but your cluster must contain at least 1 node across all node pools.
	MaxNodes *int `json:"max_nodes,omitempty"`

	// ProviderID is the unique identifier as specified by the cloud provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`
}

// DOKSNodePoolStatus defines the observed state of DOKSNodePool
type DOKSNodePoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ready denotes that the cluster (infrastructure) is ready.
	// +optional
	Ready bool `json:"ready"`

	// ProviderStatus is the status of the DigitalOcean droplet instance for this machine.
	// +optional
	ProviderStatus *DOResourceStatus `json:"providerStatus,omitempty"`

	// FailureReason will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a succinct value suitable
	// for machine interpretation.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *errors.MachineStatusError `json:"failureReason,omitempty"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DOKSNodePool is the Schema for the doksnodepools API
type DOKSNodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DOKSNodePoolSpec   `json:"spec,omitempty"`
	Status DOKSNodePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DOKSNodePoolList contains a list of DOKSNodePool
type DOKSNodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOKSNodePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DOKSNodePool{}, &DOKSNodePoolList{})
}
