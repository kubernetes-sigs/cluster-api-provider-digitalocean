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
	"fmt"
	"strings"
)

// DOSafeName returns DigitalOcean safe name with replacing '.' and '/' to '-'
// since DigitalOcean doesn't support naming with those character.
func DOSafeName(name string) string {
	r := strings.NewReplacer(".", "-", "/", "-")
	return r.Replace(name)
}

// DOControlPlaneDNS ...
type DOControlPlaneDNS struct {
	// Domain is the DO domain that this record should live in. It must be pre-existing in your DO account.
	// The format must be a string that conforms to the definition of a subdomain in DNS (RFC 1123)
	// +kubebuilder:validation:Pattern:=^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$
	Domain string `json:"domain"`
	// Name is the DNS short name of the record (non-FQDN)
	// The format must consist of alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character
	// +kubebuilder:validation:Pattern:=^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$
	Name string `json:"name"`
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

// DONetworkResource encapsulates DigitalOcean networking resources.
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

// DataDiskName is the volume name used for a data disk of a droplet.
// It's in the form of <dropletName>-<dataDiskNameSuffix>.
func DataDiskName(m *DOMachine, suffix string) string {
	return DOSafeName(fmt.Sprintf("%s-%s", m.Name, suffix))
}

// DataDisk specifies the parameters that are used to add a data disk to the machine.
type DataDisk struct {
	// NameSuffix is the suffix to be appended to the machine name to generate the disk name.
	// Each disk name will be in format <dropletName>-<nameSuffix>.
	NameSuffix string `json:"nameSuffix"`
	// DiskSizeGB is the size in GB to assign to the data disk.
	DiskSizeGB int64 `json:"diskSizeGB"`
	// FilesystemType to be used on the volume. When provided the volume will
	// be automatically formatted.
	FilesystemType string `json:"filesystemType,omitempty"`
	// FilesystemLabel is the label that is applied to the created filesystem.
	// Character limits apply: 16 for ext4; 12 for xfs.
	// May only be used in conjunction with filesystemType.
	FilesystemLabel string `json:"filesystemLabel,omitempty"`
}

// DONetwork encapsulates DigitalOcean networking configuration.
type DONetwork struct {
	// Configures an API Server loadbalancers
	// +optional
	APIServerLoadbalancers DOLoadBalancer `json:"apiServerLoadbalancers,omitempty"`
	// VPC defines the VPC configuration.
	// +optional
	VPC DOVPC `json:"vpc,omitempty"`
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
	// The DO load balancer UUID. If omitted, a new load balancer will be created.
	// +optional
	ResourceID string `json:"resourceId,omitempty"`
}

// DOVPC define the DigitalOcean VPC configuration.
type DOVPC struct {
	// VPCUUID defines the VPC UUID to use. An empty value implies using the
	// default VPC.
	// +optional
	VPCUUID string `json:"vpc_uuid,omitempty"`
}

// DOVolume defines a DO Block Storage Volume.
type DOVolume struct {
	ID string `json:"id"`
}

var (
	// DefaultLBPort default LoadBalancer port.
	DefaultLBPort = 6443
	// DefaultLBAlgorithm default LoadBalancer algorithm.
	DefaultLBAlgorithm = "round_robin"
	// DefaultLBHealthCheckInterval default LoadBalancer health check interval.
	DefaultLBHealthCheckInterval = 10
	// DefaultLBHealthCheckTimeout default LoadBalancer health check timeout.
	DefaultLBHealthCheckTimeout = 5
	// DefaultLBHealthCheckUnhealthyThreshold default LoadBalancer unhealthy threshold.
	DefaultLBHealthCheckUnhealthyThreshold = 3
	// DefaultLBHealthCheckHealthyThreshold default LoadBalancer healthy threshold.
	DefaultLBHealthCheckHealthyThreshold = 5
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
