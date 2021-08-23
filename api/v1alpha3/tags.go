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
	"fmt"
)

// Tags defines a slice of tags.
type Tags []string

const (
	// NameDigitalOceanProviderPrefix is the tag prefix for
	// cluster-api-provider-digitalocean owned components.
	NameDigitalOceanProviderPrefix = "sigs-k8s-io:capdo"
	// APIServerRoleTagValue describes the value for the apiserver role.
	APIServerRoleTagValue = "apiserver"
	// NodeRoleTagValue describes the value for the node role.
	NodeRoleTagValue = "node"
)

// ClusterNameTag generates the tag with prefix `NameDigitalOceanProviderPrefix`
// for resources associated with a cluster. It will generated tag like `sigs-k8s-io:capdo:{clusterName}`.
func ClusterNameTag(clusterName string) string {
	return fmt.Sprintf("%s:%s", NameDigitalOceanProviderPrefix, clusterName)
}

// ClusterNameRoleTag generates the tag with prefix `NameDigitalOceanProviderPrefix` and `RoleValue` as suffix
// It will generated tag like `sigs-k8s-io:capdo:{clusterName}:{role}`.
func ClusterNameRoleTag(clusterName, role string) string {
	return fmt.Sprintf("%s:%s:%s", NameDigitalOceanProviderPrefix, clusterName, role)
}

// ClusterNameUIDRoleTag generates the tag with prefix `NameDigitalOceanProviderPrefix` and `RoleValue` as suffix
// It will generated tag like `sigs-k8s-io:capdo:{clusterName}:{UID}:{role}`.
func ClusterNameUIDRoleTag(clusterName, clusterUID, role string) string {
	return fmt.Sprintf("%s:%s:%s:%s", NameDigitalOceanProviderPrefix, clusterName, clusterUID, role)
}

// NameTagFromName returns DigitalOcean safe name tag from name.
func NameTagFromName(name string) string {
	return fmt.Sprintf("name:%s", DOSafeName(name))
}

// BuildTagParams is used to build tags around an DigitalOcean resource.
type BuildTagParams struct {
	// ClusterName is the cluster associated with the resource.
	ClusterName string
	// ClusterUID is the cluster uid from clusters.cluster.x-k8s.io uid
	ClusterUID string
	// Name is the name of the resource, it's applied as the tag "name" on DigitalOcean.
	Name string
	// Role is the role associated to the resource.
	Role string
	// Any additional tags to be added to the resource.
	// +optional
	Additional Tags
}

// BuildTags builds tags including the cluster tag and returns them in map form.
func BuildTags(params BuildTagParams) Tags {
	var tags Tags
	tags = append(tags, ClusterNameTag(params.ClusterName))
	tags = append(tags, ClusterNameRoleTag(params.ClusterName, params.Role))
	tags = append(tags, ClusterNameUIDRoleTag(params.ClusterName, params.ClusterUID, params.Role))
	tags = append(tags, NameTagFromName(params.Name))

	tags = append(tags, params.Additional...)
	return tags
}
