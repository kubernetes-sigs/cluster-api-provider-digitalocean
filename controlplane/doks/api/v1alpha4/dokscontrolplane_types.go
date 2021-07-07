/*
Copyright 2021.

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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// DOKSControlPlaneFinalizer allows ReconcileDOKSControlPlane to clean up DigitalOcean resources associated
	// with DOKSControlPlane before removing it from the apiserver.
	DOKSControlPlaneFinalizer = "dokscontrolplane.infrastructure.cluster.x-k8s.io"
)

// DOKSControlPlaneSpec defines the desired state of DOKSControlPlane
type DOKSControlPlaneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the Cluster resource located in the same namespace.
	ClusterName string `json:"clusterName,omitempty"`
}

// DOKSControlPlaneStatus defines the observed state of DOKSControlPlane
type DOKSControlPlaneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ExternalManagedControlPlane indicates to cluster-api that the control plane
	// is managed by an external service such as AKS, EKS, GKE, etc.
	// +kubebuilder:default=true
	ExternalManagedControlPlane *bool `json:"externalManagedControlPlane,omitempty"`
	// Initialized denotes that the control plane kubeconfig secret is available.
	// +optional
	Initialized bool `json:"initialized"`
	// Ready denotes that the control plane (infrastructure) is ready.
	// +optional
	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DOKSControlPlane is the Schema for the dokscontrolplanes API
type DOKSControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DOKSControlPlaneSpec   `json:"spec,omitempty"`
	Status DOKSControlPlaneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DOKSControlPlaneList contains a list of DOKSControlPlane
type DOKSControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOKSControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DOKSControlPlane{}, &DOKSControlPlaneList{})
}
