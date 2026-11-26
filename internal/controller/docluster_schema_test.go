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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
)

// The control plane endpoint is only known once the load balancer exists. Cluster API v1.12
// made host and port optional in the embedded v1beta1 APIEndpoint. The typed Go client always
// sends both fields (they have no omitempty), so these specs create raw objects, like
// `kubectl apply` would, to check the generated CRD schema against a real API server.
var _ = Describe("DOCluster CRD schema", func() {
	ctx := context.Background()

	newDOCluster := func(name string, spec map[string]interface{}) *unstructured.Unstructured {
		u := &unstructured.Unstructured{Object: map[string]interface{}{"spec": spec}}
		u.SetGroupVersionKind(infrav1.GroupVersion.WithKind("DOCluster"))
		u.SetNamespace("default")
		u.SetName(name)
		return u
	}

	It("accepts a control plane endpoint with only a host", func() {
		c := newDOCluster("endpoint-host-only", map[string]interface{}{
			"region":               "sfo2",
			"controlPlaneEndpoint": map[string]interface{}{"host": "203.0.113.10"},
		})
		Expect(k8sClient.Create(ctx, c)).To(Succeed())
		DeferCleanup(func() { Expect(k8sClient.Delete(ctx, c)).To(Succeed()) })
	})

	It("accepts an empty control plane endpoint and setting it later", func() {
		c := newDOCluster("endpoint-empty", map[string]interface{}{
			"region":               "sfo2",
			"controlPlaneEndpoint": map[string]interface{}{},
		})
		Expect(k8sClient.Create(ctx, c)).To(Succeed())
		DeferCleanup(func() { Expect(k8sClient.Delete(ctx, c)).To(Succeed()) })

		By("setting the endpoint, as the controller does once the load balancer exists")
		Expect(unstructured.SetNestedMap(c.Object, map[string]interface{}{"host": "203.0.113.11", "port": int64(6443)}, "spec", "controlPlaneEndpoint")).To(Succeed())
		Expect(k8sClient.Update(ctx, c)).To(Succeed())
	})
})
