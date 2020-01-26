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

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	capierrors "sigs.k8s.io/cluster-api/errors"
)

const (
	// MachineFinalizer allows ReconcileDOMachine to clean up DigitalOcean resources associated with DOMachine before
	// removing it from the apiserver.
	MachineFinalizer = "domachine.infrastructure.cluster.x-k8s.io"
)

// DOMachineSpec defines the desired state of DOMachine
type DOMachineSpec struct {
	// ProviderID is the unique identifier as specified by the cloud provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`
	// Droplet size. It must be known DigitalOcean droplet size. See https://developers.digitalocean.com/documentation/v2/#list-all-sizes
	Size string `json:"size"`
	// Droplet image can be image id or slug. See https://developers.digitalocean.com/documentation/v2/#list-all-images
	Image intstr.IntOrString `json:"image"`
	// SSHKeys is the ssh key id or fingerprint to attach in DigitalOcean droplet.
	// It must be available on DigitalOcean account. See https://developers.digitalocean.com/documentation/v2/#list-all-keys
	SSHKeys []intstr.IntOrString `json:"sshKeys"`
	// AdditionalTags is an optional set of tags to add to DigitalOcean resources managed by the DigitalOcean provider.
	// +optional
	AdditionalTags Tags `json:"additionalTags,omitempty"`
}

// DOMachineStatus defines the observed state of DOMachine
type DOMachineStatus struct {
	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`
	// Addresses contains the DigitalOcean droplet associated addresses.
	Addresses []corev1.NodeAddress `json:"addresses,omitempty"`
	// InstanceStatus is the status of the DigitalOcean droplet instance for this machine.
	// +optional
	InstanceStatus *DOResourceStatus `json:"instanceStatus,omitempty"`
	// ErrorReason will be set in the event that there is a terminal problem
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
	ErrorReason *capierrors.MachineStatusError `json:"errorReason,omitempty"`
	// ErrorMessage will be set in the event that there is a terminal problem
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
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=domachines,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status

// DOMachine is the Schema for the domachines API
type DOMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DOMachineSpec   `json:"spec,omitempty"`
	Status DOMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DOMachineList contains a list of DOMachine
type DOMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DOMachine{}, &DOMachineList{})
}
