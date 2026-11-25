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
	"testing"

	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
)

func TestDOMachineReconciler_Paused(t *testing.T) {
	// The scopes build a godo client from these; nothing in the tested paths
	// calls the DigitalOcean API, and the URL makes sure nothing could.
	t.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "fake-token")
	t.Setenv("DIGITALOCEAN_API_URL", "http://127.0.0.1:1")

	const (
		clusterName = "paused-cluster"
		machineName = "paused-machine"
	)

	tests := []struct {
		name            string
		clusterPaused   bool
		doMachinePaused bool
		doClusterPaused bool
		wantFinalizer   bool
	}{
		{
			name:          "nothing paused: DOMachine is reconciled",
			wantFinalizer: true,
		},
		{
			name:            "DOMachine has the paused annotation: not reconciled",
			doMachinePaused: true,
			wantFinalizer:   false,
		},
		{
			name:          "Cluster is paused: not reconciled",
			clusterPaused: true,
			wantFinalizer: false,
		},
		{
			name:            "only the DOCluster has the paused annotation: DOMachine is still reconciled",
			doClusterPaused: true,
			wantFinalizer:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			scheme, err := setupScheme()
			g.Expect(err).NotTo(HaveOccurred())

			cluster := newCluster(clusterName)
			cluster.Spec.InfrastructureRef = clusterv1beta2.ContractVersionedObjectReference{
				APIGroup: infrav1.GroupVersion.Group,
				Kind:     "DOCluster",
				Name:     clusterName,
			}
			if tt.clusterPaused {
				cluster.Spec.Paused = ptr.To(true)
			}

			doCluster := &infrav1.DOCluster{
				ObjectMeta: metav1.ObjectMeta{Name: clusterName, Namespace: namespace},
			}
			if tt.doClusterPaused {
				doCluster.Annotations = map[string]string{clusterv1beta2.PausedAnnotation: ""}
			}

			machine := newMachineWithInfrastructureRef(clusterName, machineName)

			doMachine := &infrav1.DOMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      machineName,
					Namespace: namespace,
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: clusterv1beta2.GroupVersion.String(),
						Kind:       "Machine",
						Name:       machineName,
					}},
				},
			}
			if tt.doMachinePaused {
				doMachine.Annotations = map[string]string{clusterv1beta2.PausedAnnotation: ""}
			}

			c := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(cluster, doCluster, machine, doMachine).
				Build()

			r := &DOMachineReconciler{
				Client:   c,
				Recorder: record.NewFakeRecorder(10),
			}

			key := client.ObjectKeyFromObject(doMachine)
			_, err = r.Reconcile(context.Background(), reconcile.Request{NamespacedName: key})
			g.Expect(err).NotTo(HaveOccurred())

			got := &infrav1.DOMachine{}
			g.Expect(c.Get(context.Background(), key, got)).To(Succeed())
			g.Expect(controllerutil.ContainsFinalizer(got, infrav1.MachineFinalizer)).To(Equal(tt.wantFinalizer))
		})
	}
}
