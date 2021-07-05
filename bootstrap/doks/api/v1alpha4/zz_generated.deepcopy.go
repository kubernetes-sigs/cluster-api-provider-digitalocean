// +build !ignore_autogenerated

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

package v1alpha4

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOKSBootstrapConfig) DeepCopyInto(out *DOKSBootstrapConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOKSBootstrapConfig.
func (in *DOKSBootstrapConfig) DeepCopy() *DOKSBootstrapConfig {
	if in == nil {
		return nil
	}
	out := new(DOKSBootstrapConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOKSBootstrapConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOKSBootstrapConfigList) DeepCopyInto(out *DOKSBootstrapConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DOKSBootstrapConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOKSBootstrapConfigList.
func (in *DOKSBootstrapConfigList) DeepCopy() *DOKSBootstrapConfigList {
	if in == nil {
		return nil
	}
	out := new(DOKSBootstrapConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DOKSBootstrapConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOKSBootstrapConfigSpec) DeepCopyInto(out *DOKSBootstrapConfigSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOKSBootstrapConfigSpec.
func (in *DOKSBootstrapConfigSpec) DeepCopy() *DOKSBootstrapConfigSpec {
	if in == nil {
		return nil
	}
	out := new(DOKSBootstrapConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DOKSBootstrapConfigStatus) DeepCopyInto(out *DOKSBootstrapConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DOKSBootstrapConfigStatus.
func (in *DOKSBootstrapConfigStatus) DeepCopy() *DOKSBootstrapConfigStatus {
	if in == nil {
		return nil
	}
	out := new(DOKSBootstrapConfigStatus)
	in.DeepCopyInto(out)
	return out
}
