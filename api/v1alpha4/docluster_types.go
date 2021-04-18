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

package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
)

const (
	// ClusterFinalizer allows ReconcileDOCluster to clean up DigitalOcean resources associated with DOCluster before
	// removing it from the apiserver.
	ClusterFinalizer = "docluster.infrastructure.cluster.x-k8s.io"
)

// DOClusterSpec defines the desired state of DOCluster.
type DOClusterSpec struct {
	// The DigitalOcean Region the cluster lives in. It must be one of available
	// region on DigitalOcean. See
	// https://developers.digitalocean.com/documentation/v2/#list-all-regions
	Region string `json:"region"`
	// Network configurations
	// +optional
	Network DONetwork `json:"network,omitempty"`
	// ControlPlaneEndpoint represents the endpoint used to communicate with the
	// control plane. If ControlPlaneDNS is unset, the DO load-balancer IP
	// of the Kubernetes API Server is used.
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`
	// ControlPlaneDNS is a managed DNS name that points to the load-balancer
	// IP used for the ControlPlaneEndpoint.
	// +optional
	ControlPlaneDNS *DOControlPlaneDNS `json:"controlPlaneDNS,omitempty"`
}

// DOClusterStatus defines the observed state of DOCluster.
type DOClusterStatus struct {
	// Ready denotes that the cluster (infrastructure) is ready.
	// +optional
	Ready bool `json:"ready"`
	// ControlPlaneDNSRecordReady denotes that the DNS record is ready and
	// propagated to the DO DNS servers.
	// +optional
	ControlPlaneDNSRecordReady bool `json:"controlPlaneDNSRecordReady,omitempty"`
	// Network encapsulates all things related to DigitalOcean network.
	// +optional
	Network DONetworkResource `json:"network,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=doclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this DOCluster belongs"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Cluster infrastructure is ready for DigitalOcean droplet instances"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.ControlPlaneEndpoint",description="API Endpoint",priority=1

// DOCluster is the Schema for the DOClusters API.
type DOCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DOClusterSpec   `json:"spec,omitempty"`
	Status DOClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DOClusterList contains a list of DOCluster.
type DOClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DOCluster{}, &DOClusterList{})
}
