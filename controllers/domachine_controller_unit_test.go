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

package controllers

import (
	"testing"

	. "github.com/onsi/gomega"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	namespace = "default"
)

func setupScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := infrav1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := clusterv1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func newCluster(name string) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func newMachine(clusterName, machineName string) *clusterv1.Machine {
	return &clusterv1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				clusterv1.ClusterLabelName: clusterName,
			},
			Name:      machineName,
			Namespace: namespace,
		},
	}
}

func newMachineWithInfrastructureRef(clusterName, machineName string) *clusterv1.Machine {
	m := newMachine(clusterName, machineName)
	m.Spec.InfrastructureRef = corev1.ObjectReference{
		Kind:       "DOMachine",
		Namespace:  "",
		Name:       machineName,
		APIVersion: infrav1.GroupVersion.String(),
	}
	return m
}

func TestDOMachineReconciler_DOClusterToDOMachines(t *testing.T) {
	g := NewWithT(t)
	scheme, err := setupScheme()
	if err != nil {
		t.Fatal(err)
	}
	clusterName := "test-cluster"
	initObjects := []client.Object{
		newCluster(clusterName),
		newMachine(clusterName, "my-machine-0"),
		newMachineWithInfrastructureRef(clusterName, "my-machine-1"),
		newMachineWithInfrastructureRef(clusterName, "my-machine-2"),
	}

	fakec := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjects...).
		Build()

	type fields struct {
		Client   client.Client
		Recorder record.EventRecorder
	}

	tests := []struct {
		name     string
		fields   fields
		request  client.Object
		expected []reconcile.Request
	}{
		{
			name: "two-machine-with-infra-ref",
			fields: fields{
				Client: fakec,
			},
			request: &infrav1.DOCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							Name:       clusterName,
							Kind:       "Cluster",
							APIVersion: clusterv1.GroupVersion.String(),
						},
					},
				},
			},
			expected: []reconcile.Request{
				{
					NamespacedName: client.ObjectKey{
						Namespace: "default",
						Name:      "my-machine-1",
					},
				},
				{
					NamespacedName: client.ObjectKey{
						Namespace: "default",
						Name:      "my-machine-2",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DOMachineReconciler{
				Client:   tt.fields.Client,
				Recorder: tt.fields.Recorder,
			}

			fn := r.DOClusterToDOMachines(ctrl.SetupSignalHandler())
			out := fn(tt.request)
			g.Expect(out).To(Equal(tt.expected))
		})
	}
}
