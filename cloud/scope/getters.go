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

package scope

var (
	// DefaultClusterScopeGetter ...
	DefaultClusterScopeGetter ClusterScopeGetter = ClusterScopeGetterFunc(NewClusterScope)
	// DefaultMachineScopeGetter ...
	DefaultMachineScopeGetter MachineScopeGetter = MachineScopeGetterFunc(NewMachineScope)
)

// ClusterScopeGetter ...
type ClusterScopeGetter interface {
	ClusterScope(params ClusterScopeParams) (*ClusterScope, error)
}

// ClusterScopeGetterFunc ...
type ClusterScopeGetterFunc func(params ClusterScopeParams) (*ClusterScope, error)

// ClusterScope ...
func (f ClusterScopeGetterFunc) ClusterScope(params ClusterScopeParams) (*ClusterScope, error) {
	return f(params)
}

// MachineScopeGetter ...
type MachineScopeGetter interface {
	MachineScope(params MachineScopeParams) (*MachineScope, error)
}

// MachineScopeGetterFunc ...
type MachineScopeGetterFunc func(params MachineScopeParams) (*MachineScope, error)

// MachineScope ...
func (f MachineScopeGetterFunc) MachineScope(params MachineScopeParams) (*MachineScope, error) {
	return f(params)
}
