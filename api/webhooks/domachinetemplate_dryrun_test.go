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
	"context"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
)

func TestDOMachineTemplateWebhook_ValidateUpdate_TopologyDryRun(t *testing.T) {
	w := &DOMachineTemplateWebhook{}

	newTemplate := func(size string, annotations map[string]string) *v1beta1.DOMachineTemplate {
		return &v1beta1.DOMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Annotations: annotations},
			Spec: v1beta1.DOMachineTemplateSpec{
				Template: v1beta1.DOMachineTemplateResource{
					Spec: v1beta1.DOMachineSpec{
						Size:    size,
						Image:   intstr.FromString("ubuntu-24-04-x64"),
						SSHKeys: []intstr.IntOrString{intstr.FromString("aa:bb:cc")},
					},
				},
			},
		}
	}
	dryRunAnnotation := map[string]string{clusterv1.TopologyDryRunAnnotation: ""}

	tests := []struct {
		name        string
		dryRun      *bool
		annotations map[string]string
		newSize     string
		wantErr     bool
	}{
		{name: "unchanged spec", dryRun: ptr.To(false), newSize: "s-2vcpu-2gb"},
		{name: "spec change", dryRun: ptr.To(false), newSize: "s-4vcpu-8gb", wantErr: true},
		{name: "spec change in a topology dry run", dryRun: ptr.To(true), annotations: dryRunAnnotation, newSize: "s-4vcpu-8gb"},
		{name: "spec change in a dry run without the topology annotation", dryRun: ptr.To(true), newSize: "s-4vcpu-8gb", wantErr: true},
		{name: "spec change with the topology annotation but not a dry run", dryRun: ptr.To(false), annotations: dryRunAnnotation, newSize: "s-4vcpu-8gb", wantErr: true},
		{name: "spec change with dry run unset", annotations: dryRunAnnotation, newSize: "s-4vcpu-8gb", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := admission.NewContextWithRequest(context.Background(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{DryRun: tt.dryRun},
			})
			_, err := w.ValidateUpdate(ctx, newTemplate("s-2vcpu-2gb", nil), newTemplate(tt.newSize, tt.annotations))
			if tt.wantErr {
				if !apierrors.IsInvalid(err) {
					t.Fatalf("expected an Invalid error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}

	t.Run("no admission request in the context is a normal update", func(t *testing.T) {
		if _, err := w.ValidateUpdate(context.Background(), newTemplate("s-2vcpu-2gb", nil), newTemplate("s-2vcpu-2gb", nil)); err != nil {
			t.Fatalf("expected no error for an unchanged spec, got %v", err)
		}
		_, err := w.ValidateUpdate(context.Background(), newTemplate("s-2vcpu-2gb", nil), newTemplate("s-4vcpu-8gb", dryRunAnnotation))
		if !apierrors.IsInvalid(err) {
			t.Fatalf("expected an Invalid error for a spec change, got %v", err)
		}
	})
}
