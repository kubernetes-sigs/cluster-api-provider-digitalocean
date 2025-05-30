//go:build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/cluster-api/errors"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BuildTagParams) DeepCopyInto(out *BuildTagParams) {
	*out = *in
	if in.Additional != nil {
		in, out := &in.Additional, &out.Additional
		*out = make(Tags, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BuildTagParams.
func (in *BuildTagParams) DeepCopy() *BuildTagParams {
	if in == nil {
		return nil
	}
	out := new(BuildTagParams)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOCluster) DeepCopyInto(out *DOCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOCluster.
func (in *DOCluster) DeepCopy() *DOCluster {
	if in == nil {
		return nil
	}
	out := new(DOCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOClusterList) DeepCopyInto(out *DOClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DOCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOClusterList.
func (in *DOClusterList) DeepCopy() *DOClusterList {
	if in == nil {
		return nil
	}
	out := new(DOClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOClusterSpec) DeepCopyInto(out *DOClusterSpec) {
	*out = *in
	out.Network = in.Network
	out.ControlPlaneEndpoint = in.ControlPlaneEndpoint
	if in.ControlPlaneDNS != nil {
		in, out := &in.ControlPlaneDNS, &out.ControlPlaneDNS
		*out = new(DOControlPlaneDNS)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOClusterSpec.
func (in *DOClusterSpec) DeepCopy() *DOClusterSpec {
	if in == nil {
		return nil
	}
	out := new(DOClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOClusterStatus) DeepCopyInto(out *DOClusterStatus) {
	*out = *in
	out.Network = in.Network
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOClusterStatus.
func (in *DOClusterStatus) DeepCopy() *DOClusterStatus {
	if in == nil {
		return nil
	}
	out := new(DOClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOClusterTemplate) DeepCopyInto(out *DOClusterTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOClusterTemplate.
func (in *DOClusterTemplate) DeepCopy() *DOClusterTemplate {
	if in == nil {
		return nil
	}
	out := new(DOClusterTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOClusterTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOClusterTemplateList) DeepCopyInto(out *DOClusterTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DOClusterTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOClusterTemplateList.
func (in *DOClusterTemplateList) DeepCopy() *DOClusterTemplateList {
	if in == nil {
		return nil
	}
	out := new(DOClusterTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOClusterTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOClusterTemplateResource) DeepCopyInto(out *DOClusterTemplateResource) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOClusterTemplateResource.
func (in *DOClusterTemplateResource) DeepCopy() *DOClusterTemplateResource {
	if in == nil {
		return nil
	}
	out := new(DOClusterTemplateResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOClusterTemplateSpec) DeepCopyInto(out *DOClusterTemplateSpec) {
	*out = *in
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOClusterTemplateSpec.
func (in *DOClusterTemplateSpec) DeepCopy() *DOClusterTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(DOClusterTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOControlPlaneDNS) DeepCopyInto(out *DOControlPlaneDNS) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOControlPlaneDNS.
func (in *DOControlPlaneDNS) DeepCopy() *DOControlPlaneDNS {
	if in == nil {
		return nil
	}
	out := new(DOControlPlaneDNS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOLoadBalancer) DeepCopyInto(out *DOLoadBalancer) {
	*out = *in
	out.HealthCheck = in.HealthCheck
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOLoadBalancer.
func (in *DOLoadBalancer) DeepCopy() *DOLoadBalancer {
	if in == nil {
		return nil
	}
	out := new(DOLoadBalancer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOLoadBalancerHealthCheck) DeepCopyInto(out *DOLoadBalancerHealthCheck) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOLoadBalancerHealthCheck.
func (in *DOLoadBalancerHealthCheck) DeepCopy() *DOLoadBalancerHealthCheck {
	if in == nil {
		return nil
	}
	out := new(DOLoadBalancerHealthCheck)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOMachine) DeepCopyInto(out *DOMachine) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOMachine.
func (in *DOMachine) DeepCopy() *DOMachine {
	if in == nil {
		return nil
	}
	out := new(DOMachine)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOMachine) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOMachineList) DeepCopyInto(out *DOMachineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DOMachine, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOMachineList.
func (in *DOMachineList) DeepCopy() *DOMachineList {
	if in == nil {
		return nil
	}
	out := new(DOMachineList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOMachineList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOMachineSpec) DeepCopyInto(out *DOMachineSpec) {
	*out = *in
	if in.ProviderID != nil {
		in, out := &in.ProviderID, &out.ProviderID
		*out = new(string)
		**out = **in
	}
	out.Image = in.Image
	if in.DataDisks != nil {
		in, out := &in.DataDisks, &out.DataDisks
		*out = make([]DataDisk, len(*in))
		copy(*out, *in)
	}
	if in.SSHKeys != nil {
		in, out := &in.SSHKeys, &out.SSHKeys
		*out = make([]intstr.IntOrString, len(*in))
		copy(*out, *in)
	}
	if in.AdditionalTags != nil {
		in, out := &in.AdditionalTags, &out.AdditionalTags
		*out = make(Tags, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOMachineSpec.
func (in *DOMachineSpec) DeepCopy() *DOMachineSpec {
	if in == nil {
		return nil
	}
	out := new(DOMachineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOMachineStatus) DeepCopyInto(out *DOMachineStatus) {
	*out = *in
	if in.Addresses != nil {
		in, out := &in.Addresses, &out.Addresses
		*out = make([]v1.NodeAddress, len(*in))
		copy(*out, *in)
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]DOVolume, len(*in))
		copy(*out, *in)
	}
	if in.InstanceStatus != nil {
		in, out := &in.InstanceStatus, &out.InstanceStatus
		*out = new(DOResourceStatus)
		**out = **in
	}
	if in.FailureReason != nil {
		in, out := &in.FailureReason, &out.FailureReason
		*out = new(errors.MachineStatusError)
		**out = **in
	}
	if in.FailureMessage != nil {
		in, out := &in.FailureMessage, &out.FailureMessage
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOMachineStatus.
func (in *DOMachineStatus) DeepCopy() *DOMachineStatus {
	if in == nil {
		return nil
	}
	out := new(DOMachineStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOMachineTemplate) DeepCopyInto(out *DOMachineTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOMachineTemplate.
func (in *DOMachineTemplate) DeepCopy() *DOMachineTemplate {
	if in == nil {
		return nil
	}
	out := new(DOMachineTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOMachineTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOMachineTemplateList) DeepCopyInto(out *DOMachineTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DOMachineTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOMachineTemplateList.
func (in *DOMachineTemplateList) DeepCopy() *DOMachineTemplateList {
	if in == nil {
		return nil
	}
	out := new(DOMachineTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOMachineTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOMachineTemplateResource) DeepCopyInto(out *DOMachineTemplateResource) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOMachineTemplateResource.
func (in *DOMachineTemplateResource) DeepCopy() *DOMachineTemplateResource {
	if in == nil {
		return nil
	}
	out := new(DOMachineTemplateResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOMachineTemplateSpec) DeepCopyInto(out *DOMachineTemplateSpec) {
	*out = *in
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOMachineTemplateSpec.
func (in *DOMachineTemplateSpec) DeepCopy() *DOMachineTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(DOMachineTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DONetwork) DeepCopyInto(out *DONetwork) {
	*out = *in
	out.APIServerLoadbalancers = in.APIServerLoadbalancers
	out.VPC = in.VPC
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DONetwork.
func (in *DONetwork) DeepCopy() *DONetwork {
	if in == nil {
		return nil
	}
	out := new(DONetwork)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DONetworkResource) DeepCopyInto(out *DONetworkResource) {
	*out = *in
	out.APIServerLoadbalancersRef = in.APIServerLoadbalancersRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DONetworkResource.
func (in *DONetworkResource) DeepCopy() *DONetworkResource {
	if in == nil {
		return nil
	}
	out := new(DONetworkResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOResourceReference) DeepCopyInto(out *DOResourceReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOResourceReference.
func (in *DOResourceReference) DeepCopy() *DOResourceReference {
	if in == nil {
		return nil
	}
	out := new(DOResourceReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOVPC) DeepCopyInto(out *DOVPC) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOVPC.
func (in *DOVPC) DeepCopy() *DOVPC {
	if in == nil {
		return nil
	}
	out := new(DOVPC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOVolume) DeepCopyInto(out *DOVolume) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOVolume.
func (in *DOVolume) DeepCopy() *DOVolume {
	if in == nil {
		return nil
	}
	out := new(DOVolume)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataDisk) DeepCopyInto(out *DataDisk) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataDisk.
func (in *DataDisk) DeepCopy() *DataDisk {
	if in == nil {
		return nil
	}
	out := new(DataDisk)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Tags) DeepCopyInto(out *Tags) {
	{
		in := &in
		*out = make(Tags, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tags.
func (in Tags) DeepCopy() Tags {
	if in == nil {
		return nil
	}
	out := new(Tags)
	in.DeepCopyInto(out)
	return *out
}
