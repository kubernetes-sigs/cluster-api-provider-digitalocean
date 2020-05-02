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

import "strings"

// APIEndpoint represents a reachable Kubernetes API endpoint.
type APIEndpoint struct {
	// The hostname on which the API server is serving.
	Host string `json:"host"`
	// The port on which the API server is serving.
	Port int `json:"port"`
}

// DOSafeName returns DigitalOcean safe name with replacing '.' and '/' to '-'
// since DigitalOcean doesn't support naming with those character.
func DOSafeName(name string) string {
	r := strings.NewReplacer(".", "-", "/", "-")
	return r.Replace(name)
}

// DOResourceStatus describes the status of a DigitalOcean resource.
type DOResourceStatus string

var (
	// DOResourceStatusNew is the string representing a DigitalOcean resource just created and in a provisioning state.
	DOResourceStatusNew = DOResourceStatus("new")
	// DOResourceStatusRunning is the string representing a DigitalOcean resource already provisioned and in a active state.
	DOResourceStatusRunning = DOResourceStatus("active")
	// DOResourceStatusErrored is the string representing a DigitalOcean resource in a errored state.
	DOResourceStatusErrored = DOResourceStatus("errored")
	// DOResourceStatusOff is the string representing a DigitalOcean resource in off state.
	DOResourceStatusOff = DOResourceStatus("off")
	// DOResourceStatusArchive is the string representing a DigitalOcean resource in archive state.
	DOResourceStatusArchive = DOResourceStatus("archive")
)

// DOResourceReference is a reference to a DigitalOcean resource.
type DOResourceReference struct {
	// ID of DigitalOcean resource
	// +optional
	ResourceID string `json:"resourceId,omitempty"`
	// Status of DigitalOcean resource
	// +optional
	ResourceStatus DOResourceStatus `json:"resourceStatus,omitempty"`
}

// Network encapsulates DigitalOcean networking resources.
type DONetworkResource struct {
	// APIServerLoadbalancersRef is the id of apiserver loadbalancers.
	// +optional
	APIServerLoadbalancersRef DOResourceReference `json:"apiServerLoadbalancersRef,omitempty"`
}

// DOMachineTemplateResource describes the data needed to create am DOMachine from a template.
type DOMachineTemplateResource struct {
	// Spec is the specification of the desired behavior of the machine.
	Spec DOMachineSpec `json:"spec"`
}

// DONetwork encapsulates DigitalOcean networking configuration.
type DONetwork struct {
	// Configures an API Server loadbalancers
	// +optional
	APIServerLoadbalancers DOLoadBalancer `json:"apiServerLoadbalancers,omitempty"`
}

// DOLoadBalancer define the DigitalOcean loadbalancers configurations.
type DOLoadBalancer struct {
	// API Server port. It must be valid ports range (1-65535). If omitted, default value is 6443.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`
	// The API Server load balancing algorithm used to determine which backend Droplet will be selected by a client.
	// It must be either "round_robin" or "least_connections". The default value is "round_robin".
	// +optional
	// +kubebuilder:validation:Enum=round_robin;least_connections
	Algorithm string `json:"algorithm,omitempty"`
	// An object specifying health check settings for the Load Balancer. If omitted, default values will be provided.
	// +optional
	HealthCheck DOLoadBalancerHealthCheck `json:"healthCheck,omitempty"`
}

var (
	DefaultLBPort                          = 6443
	DefaultLBAlgorithm                     = "round_robin"
	DefaultLBHealthCheckInterval           = 10
	DefaultLBHealthCheckTimeout            = 5
	DefaultLBHealthCheckUnhealthyThreshold = 3
	DefaultLBHealthCheckHealthyThreshold   = 5
)

// ApplyDefault give APIServerLoadbalancers default values.
func (in *DOLoadBalancer) ApplyDefault() {
	if in.Port == 0 {
		in.Port = DefaultLBPort
	}
	if in.Algorithm == "" {
		in.Algorithm = DefaultLBAlgorithm
	}
	if in.HealthCheck.Interval == 0 {
		in.HealthCheck.Interval = DefaultLBHealthCheckInterval
	}
	if in.HealthCheck.Timeout == 0 {
		in.HealthCheck.Timeout = DefaultLBHealthCheckTimeout
	}
	if in.HealthCheck.UnhealthyThreshold == 0 {
		in.HealthCheck.UnhealthyThreshold = DefaultLBHealthCheckUnhealthyThreshold
	}
	if in.HealthCheck.HealthyThreshold == 0 {
		in.HealthCheck.HealthyThreshold = DefaultLBHealthCheckHealthyThreshold
	}
}

// DOLoadBalancerHealthCheck define the DigitalOcean loadbalancers health check configurations.
type DOLoadBalancerHealthCheck struct {
	// The number of seconds between between two consecutive health checks. The value must be between 3 and 300.
	// If not specified, the default value is 10.
	// +optional
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Maximum=300
	Interval int `json:"interval,omitempty"`
	// The number of seconds the Load Balancer instance will wait for a response until marking a health check as failed.
	// The value must be between 3 and 300. If not specified, the default value is 5.
	// +optional
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Maximum=300
	Timeout int `json:"timeout,omitempty"`
	// The number of times a health check must fail for a backend Droplet to be marked "unhealthy" and be removed from the pool.
	// The vaule must be between 2 and 10. If not specified, the default value is 3.
	// +optional
	// +kubebuilder:validation:Minimum=2
	// +kubebuilder:validation:Maximum=10
	UnhealthyThreshold int `json:"unhealthyThreshold,omitempty"`
	// The number of times a health check must pass for a backend Droplet to be marked "healthy" and be re-added to the pool.
	// The vaule must be between 2 and 10. If not specified, the default value is 5.
	// +optional
	// +kubebuilder:validation:Minimum=2
	// +kubebuilder:validation:Maximum=10
	HealthyThreshold int `json:"healthyThreshold,omitempty"`
}
