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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
)

func TestDOMachineWebhook_Default(t *testing.T) {
	webhook := &DOMachineWebhook{}
	doMachine := &v1beta1.DOMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: v1beta1.DOMachineSpec{
			Size:    "s-2vcpu-2gb",
			Image:   intstr.FromString("ubuntu-24-04-x64"),
			SSHKeys: []intstr.IntOrString{intstr.FromString("aa:bb:cc")},
		},
	}

	if err := webhook.Default(context.Background(), doMachine); err != nil {
		t.Fatalf("Default() error = %v", err)
	}
	if doMachine.Spec.PublicNetworking == nil || !*doMachine.Spec.PublicNetworking {
		t.Fatalf("expected publicNetworking to default to true, got %v", doMachine.Spec.PublicNetworking)
	}

	doMachine.Spec.PublicNetworking = ptr.To(false)
	if err := webhook.Default(context.Background(), doMachine); err != nil {
		t.Fatalf("Default() error = %v", err)
	}
	if doMachine.Spec.PublicNetworking == nil || *doMachine.Spec.PublicNetworking {
		t.Fatalf("expected explicit publicNetworking false to be preserved, got %v", doMachine.Spec.PublicNetworking)
	}
}

func TestDOMachineWebhook_ValidateCreate(t *testing.T) {
	webhook := &DOMachineWebhook{}

	t.Run("public networking enabled", func(t *testing.T) {
		doMachine := &v1beta1.DOMachine{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: v1beta1.DOMachineSpec{
				Size:             "s-2vcpu-2gb",
				Image:            intstr.FromString("ubuntu-24-04-x64"),
				SSHKeys:          []intstr.IntOrString{intstr.FromString("aa:bb:cc")},
				PublicNetworking: ptr.To(true),
			},
		}
		warnings, err := webhook.ValidateCreate(context.Background(), doMachine)
		if err != nil {
			t.Fatalf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got %v", warnings)
		}
	})

	t.Run("private droplet", func(t *testing.T) {
		doMachine := &v1beta1.DOMachine{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: v1beta1.DOMachineSpec{
				Size:             "s-2vcpu-2gb",
				Image:            intstr.FromString("ubuntu-24-04-x64"),
				SSHKeys:          []intstr.IntOrString{intstr.FromString("aa:bb:cc")},
				PublicNetworking: ptr.To(false),
			},
		}
		warnings, err := webhook.ValidateCreate(context.Background(), doMachine)
		if err != nil {
			t.Fatalf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 1 || warnings[0] != privateDropletWarning {
			t.Fatalf("expected private droplet warning, got %v", warnings)
		}
	})
}

func TestDOMachineTemplateWebhook_Default(t *testing.T) {
	webhook := &DOMachineTemplateWebhook{}
	doMachineTemplate := &v1beta1.DOMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: v1beta1.DOMachineTemplateSpec{
			Template: v1beta1.DOMachineTemplateResource{
				Spec: v1beta1.DOMachineSpec{
					Size:    "s-2vcpu-2gb",
					Image:   intstr.FromString("ubuntu-24-04-x64"),
					SSHKeys: []intstr.IntOrString{intstr.FromString("aa:bb:cc")},
				},
			},
		},
	}

	if err := webhook.Default(context.Background(), doMachineTemplate); err != nil {
		t.Fatalf("Default() error = %v", err)
	}
	if doMachineTemplate.Spec.Template.Spec.PublicNetworking == nil || !*doMachineTemplate.Spec.Template.Spec.PublicNetworking {
		t.Fatalf("expected publicNetworking to default to true, got %v", doMachineTemplate.Spec.Template.Spec.PublicNetworking)
	}
}

func TestDOMachineTemplateWebhook_ValidateCreate(t *testing.T) {
	webhook := &DOMachineTemplateWebhook{}

	t.Run("private droplet", func(t *testing.T) {
		doMachineTemplate := &v1beta1.DOMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: v1beta1.DOMachineTemplateSpec{
				Template: v1beta1.DOMachineTemplateResource{
					Spec: v1beta1.DOMachineSpec{
						Size:             "s-2vcpu-2gb",
						Image:            intstr.FromString("ubuntu-24-04-x64"),
						SSHKeys:          []intstr.IntOrString{intstr.FromString("aa:bb:cc")},
						PublicNetworking: ptr.To(false),
					},
				},
			},
		}
		warnings, err := webhook.ValidateCreate(context.Background(), doMachineTemplate)
		if err != nil {
			t.Fatalf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 1 || warnings[0] != privateDropletWarning {
			t.Fatalf("expected private droplet warning, got %v", warnings)
		}
	})
}
