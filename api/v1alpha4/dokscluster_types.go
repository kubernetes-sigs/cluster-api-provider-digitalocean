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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DOKSClusterSpec defines the desired state of DOKSCluster
type DOKSClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The slug identifier for the region where the Kubernetes cluster will be created.
	Region string `json:"region"`

	// The slug identifier for the version of Kubernetes used for the cluster.
	// See the /v2/kubernetes/options endpoint for available versions.
	Version string `json:"version"`
}

// DOKSClusterStatus defines the observed state of DOKSCluster
type DOKSClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DOKSCluster is the Schema for the doksclusters API
type DOKSCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DOKSClusterSpec   `json:"spec,omitempty"`
	Status DOKSClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DOKSClusterList contains a list of DOKSCluster
type DOKSClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOKSCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DOKSCluster{}, &DOKSClusterList{})
}
