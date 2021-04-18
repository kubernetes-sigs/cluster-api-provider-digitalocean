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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DOMachineTemplateSpec defines the desired state of DOMachineTemplate.
type DOMachineTemplateSpec struct {
	Template DOMachineTemplateResource `json:"template"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=domachinetemplates,scope=Namespaced,categories=cluster-api

// DOMachineTemplate is the Schema for the domachinetemplates API.
type DOMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DOMachineTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// DOMachineTemplateList contains a list of DOMachineTemplate.
type DOMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DOMachineTemplate{}, &DOMachineTemplateList{})
}
