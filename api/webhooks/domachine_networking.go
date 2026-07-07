/*
Copyright 2026 The Kubernetes Authors.

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

package webhooks

import (
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const privateDropletWarning = "Private droplets require VPC connectivity; ensure DOCluster.spec.network.vpc.vpc_uuid is set and the VPC has a NAT gateway for outbound access."

func defaultDOMachineSpec(spec *v1beta1.DOMachineSpec) {
	if spec.PublicNetworking == nil {
		spec.PublicNetworking = ptr.To(true)
	}
}

func privateDropletWarnings(spec v1beta1.DOMachineSpec) admission.Warnings {
	if spec.PublicNetworking != nil && !*spec.PublicNetworking {
		return admission.Warnings{privateDropletWarning}
	}
	return nil
}
