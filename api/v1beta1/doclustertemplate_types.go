/*
Copyright 2021 The Kubernetes Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DOClusterTemplateSpec defines the desired state of DOClusterTemplate.
type DOClusterTemplateSpec struct {
	Template DOClusterTemplateResource `json:"template"`
}

// DOClusterTemplateResource contains spec for DOClusterSpec.
type DOClusterTemplateResource struct {
	Spec DOClusterSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=doclustertemplates,scope=Namespaced,categories=cluster-api,shortName=doct
// +kubebuilder:storageversion

// DOClusterTemplate is the Schema for the DOclustertemplates API.
type DOClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DOClusterTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// DOClusterTemplateList contains a list of DOClusterTemplate.
type DOClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DOClusterTemplate{}, &DOClusterTemplateList{})
}
