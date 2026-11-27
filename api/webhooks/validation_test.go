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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
)

func checkResult(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if !apierrors.IsInvalid(err) && !apierrors.IsBadRequest(err) {
			t.Fatalf("expected an Invalid or BadRequest error, got %v", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func newDOCluster(region string) *v1beta1.DOCluster {
	return &v1beta1.DOCluster{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec:       v1beta1.DOClusterSpec{Region: region},
	}
}

func newDOMachineSpec() v1beta1.DOMachineSpec {
	return v1beta1.DOMachineSpec{
		Size:    "s-2vcpu-2gb",
		Image:   intstr.FromString("ubuntu-24-04-x64"),
		SSHKeys: []intstr.IntOrString{intstr.FromString("aa:bb:cc")},
	}
}

func TestDOClusterWebhook_Validate(t *testing.T) {
	w := &DOClusterWebhook{}
	ctx := context.Background()

	if _, err := w.ValidateCreate(ctx, newDOCluster("nyc1")); err != nil {
		t.Fatalf("ValidateCreate() error = %v", err)
	}
	if _, err := w.ValidateDelete(ctx, newDOCluster("nyc1")); err != nil {
		t.Fatalf("ValidateDelete() error = %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*v1beta1.DOCluster)
		wantErr bool
	}{
		{name: "no change", mutate: func(*v1beta1.DOCluster) {}},
		{name: "region change", mutate: func(c *v1beta1.DOCluster) { c.Spec.Region = "sfo3" }, wantErr: true},
		{name: "control plane endpoint set", mutate: func(c *v1beta1.DOCluster) {
			c.Spec.ControlPlaneEndpoint.Host = "203.0.113.10"
			c.Spec.ControlPlaneEndpoint.Port = 6443
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldObj := newDOCluster("nyc1")
			newObj := oldObj.DeepCopy()
			tt.mutate(newObj)
			_, err := w.ValidateUpdate(ctx, oldObj, newObj)
			checkResult(t, err, tt.wantErr)
		})
	}
}

func TestDOClusterTemplateWebhook_Validate(t *testing.T) {
	w := &DOClusterTemplateWebhook{}
	ctx := context.Background()
	newTemplate := func(region string) *v1beta1.DOClusterTemplate {
		return &v1beta1.DOClusterTemplate{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: v1beta1.DOClusterTemplateSpec{
				Template: v1beta1.DOClusterTemplateResource{Spec: v1beta1.DOClusterSpec{Region: region}},
			},
		}
	}

	if _, err := w.ValidateCreate(ctx, newTemplate("nyc1")); err != nil {
		t.Fatalf("ValidateCreate() error = %v", err)
	}
	if _, err := w.ValidateDelete(ctx, newTemplate("nyc1")); err != nil {
		t.Fatalf("ValidateDelete() error = %v", err)
	}

	t.Run("no change", func(t *testing.T) {
		_, err := w.ValidateUpdate(ctx, newTemplate("nyc1"), newTemplate("nyc1"))
		checkResult(t, err, false)
	})
	t.Run("spec change", func(t *testing.T) {
		_, err := w.ValidateUpdate(ctx, newTemplate("nyc1"), newTemplate("sfo3"))
		checkResult(t, err, true)
	})
}

func TestDOMachineWebhook_ValidateUpdate(t *testing.T) {
	w := &DOMachineWebhook{}
	ctx := context.Background()

	if _, err := w.ValidateDelete(ctx, &v1beta1.DOMachine{}); err != nil {
		t.Fatalf("ValidateDelete() error = %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*v1beta1.DOMachine)
		wantErr bool
	}{
		{name: "no change", mutate: func(*v1beta1.DOMachine) {}},
		{name: "providerID set", mutate: func(m *v1beta1.DOMachine) { m.Spec.ProviderID = ptr.To("digitalocean://123") }},
		{name: "additionalTags changed", mutate: func(m *v1beta1.DOMachine) { m.Spec.AdditionalTags = v1beta1.Tags{"team-a"} }},
		{name: "size changed", mutate: func(m *v1beta1.DOMachine) { m.Spec.Size = "s-4vcpu-8gb" }, wantErr: true},
		{name: "image changed", mutate: func(m *v1beta1.DOMachine) { m.Spec.Image = intstr.FromInt32(123) }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldObj := &v1beta1.DOMachine{ObjectMeta: metav1.ObjectMeta{Name: "test"}, Spec: newDOMachineSpec()}
			newObj := oldObj.DeepCopy()
			tt.mutate(newObj)
			_, err := w.ValidateUpdate(ctx, oldObj, newObj)
			checkResult(t, err, tt.wantErr)
		})
	}
}

func TestDOMachineTemplateWebhook_Validate(t *testing.T) {
	w := &DOMachineTemplateWebhook{}
	ctx := context.Background()
	newTemplate := func() *v1beta1.DOMachineTemplate {
		return &v1beta1.DOMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: v1beta1.DOMachineTemplateSpec{
				Template: v1beta1.DOMachineTemplateResource{Spec: newDOMachineSpec()},
			},
		}
	}

	t.Run("create", func(t *testing.T) {
		_, err := w.ValidateCreate(ctx, newTemplate())
		checkResult(t, err, false)
	})
	t.Run("create with providerID", func(t *testing.T) {
		tmpl := newTemplate()
		tmpl.Spec.Template.Spec.ProviderID = ptr.To("digitalocean://123")
		_, err := w.ValidateCreate(ctx, tmpl)
		checkResult(t, err, true)
	})
	t.Run("update without change", func(t *testing.T) {
		_, err := w.ValidateUpdate(ctx, newTemplate(), newTemplate())
		checkResult(t, err, false)
	})
	t.Run("update with spec change", func(t *testing.T) {
		newObj := newTemplate()
		newObj.Spec.Template.Spec.Size = "s-4vcpu-8gb"
		_, err := w.ValidateUpdate(ctx, newTemplate(), newObj)
		checkResult(t, err, true)
	})
	t.Run("delete", func(t *testing.T) {
		_, err := w.ValidateDelete(ctx, newTemplate())
		checkResult(t, err, false)
	})
}
