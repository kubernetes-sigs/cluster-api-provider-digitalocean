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

package e2e

import (
	"context"
	"os/exec"
	"reflect"
	"time"

	. "github.com/onsi/gomega"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha2"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// TypeToKind returns the Kind without the package prefix. Pass in a pointer to a struct
// This will panic if used incorrectly.
func TypeToKind(i interface{}) string {
	return reflect.ValueOf(i).Elem().Type().Name()
}

var (
	pollTimeout  = 5 * time.Minute
	pollInterval = 10 * time.Second
)

func WaitForClusterInfrastructureReady(client crclient.Client, namespace, name string) {
	Eventually(func() (bool, error) {
		cluster := &clusterv1.Cluster{}
		err := client.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: name}, cluster)
		if err != nil {
			return false, err
		}

		return cluster.Status.InfrastructureReady, nil
	}, pollTimeout, pollInterval).Should(BeTrue())
}

func WaitForClusterControlplaneInitialized(client crclient.Client, namespace, name string) {
	Eventually(func() (bool, error) {
		cluster := &clusterv1.Cluster{}
		err := client.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: name}, cluster)
		if err != nil {
			return false, err
		}

		return cluster.Status.ControlPlaneInitialized, nil
	}, pollTimeout, pollInterval).Should(BeTrue())
}

func WaitForMachineBootstrapReady(client crclient.Client, namespace, name string) {
	Eventually(func() (bool, error) {
		machine := &clusterv1.Machine{}
		err := client.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: name}, machine)
		if err != nil {
			return false, err
		}

		return machine.Status.BootstrapReady, nil
	}, pollTimeout, pollInterval).Should(BeTrue())
}

func WaitForDOMachineRunning(client crclient.Client, namespace, name string) {
	Eventually(func() (bool, error) {
		domachine := &infrav1.DOMachine{}
		err := client.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: name}, domachine)
		if err != nil {
			return false, err
		}

		if *domachine.Status.InstanceStatus != infrav1.DOResourceStatusRunning {
			return false, nil
		}

		return true, nil
	}, pollTimeout, pollInterval).Should(BeTrue())
}

func WaitForDOMachineReady(client crclient.Client, namespace, name string) {
	Eventually(func() (bool, error) {
		domachine := &infrav1.DOMachine{}
		err := client.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: name}, domachine)
		if err != nil {
			return false, err
		}

		return domachine.Status.Ready, nil
	}, pollTimeout, pollInterval).Should(BeTrue())
}

func WaitForMachineNodeRef(client crclient.Client, namespace, name string) {
	Eventually(func() *corev1.ObjectReference {
		machine := &clusterv1.Machine{}
		err := client.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: name}, machine)
		if err != nil {
			return nil
		}

		return machine.Status.NodeRef
	}, pollTimeout, pollInterval).ShouldNot(BeNil())
}

func WaitForMachineDeploymentRunning(client crclient.Client, namespace, name string, expectedReplicas int32) {
	Eventually(func() (bool, error) {
		machineDeployment := &clusterv1.MachineDeployment{}
		err := client.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: name}, machineDeployment)
		if err != nil {
			return false, err
		}

		if machineDeployment.Status.ReadyReplicas == expectedReplicas {
			return true, nil
		}

		return false, nil
	}, pollTimeout, pollInterval).Should(BeTrue())
}

func WaitForDeletion(client crclient.Client, obj runtime.Object, namespace, name string) {
	Eventually(func() bool {
		err := client.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: name}, obj)
		if err != nil {
			return apierrors.IsNotFound(err)
		}

		return false
	}, pollTimeout, pollInterval).Should(BeTrue())
}

func ApplyYaml(kubeconfig, manifest string) {
	Eventually(func() error {
		return exec.Command("kubectl", "create", "--kubeconfig", kubeconfig, "-f", manifest).Run()
	}, pollTimeout, pollInterval).Should(Succeed())
}
