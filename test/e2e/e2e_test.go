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
	"fmt"
	"io/ioutil"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha3"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	bootstrapkubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
)

var _ = Describe("functional tests", func() {
	Describe("cluster lifecycle", func() {
		var (
			clusterName                string
			clusterNamespace           string
			clusterGenenerator         ClusterGenerator
			machineGenerator           MachineGenerator
			machineDeploymentGenerator MachineDeploymentGenerator
		)

		BeforeEach(func() {
			var err error
			clusterName = "capdo-test-" + util.RandomString(6)
			clusterNamespace = "default"

			testTmpDir, err = ioutil.TempDir(suiteTmpDir, "e2e-test")
			Expect(err).NotTo(HaveOccurred())
		})

		It("It should be creatable, scalable and deletable", func() {
			By("Create a cluster")
			cluster, docluster := clusterGenenerator.Generate(clusterNamespace, clusterName)
			createCluster(cluster, docluster)

			By("Create a single controlplane")
			controlPlaneMachine, controlPlaneKubeadmconfig, controlPlaneDomachine := machineGenerator.Generate(clusterNamespace, clusterName, true)
			createMachine(controlPlaneMachine, controlPlaneKubeadmconfig, controlPlaneDomachine)

			By("Ensuring Cluster Controlplane Initialized")
			WaitForClusterControlplaneInitialized(kindclient, cluster.Namespace, cluster.Name)

			By("Exporting Cluster kubeconfig")
			kubeConfigData, err := kubeconfig.FromSecret(context.Background(), kindclient, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace})
			Expect(err).NotTo(HaveOccurred())
			kubeConfigPath := path.Join(testTmpDir, clusterName+".kubeconfig")
			Expect(ioutil.WriteFile(kubeConfigPath, kubeConfigData, 0600)).To(Succeed())

			By("Deploying CNI")
			ApplyYaml(kubeConfigPath, "https://docs.projectcalico.org/manifests/calico.yaml")

			By("Deploying Cloud Controller Manager")
			var ccmManifest string
			buildCloudControllerManager(&ccmManifest)
			ApplyYaml(kubeConfigPath, ccmManifest)

			By("Create worker machine deployment")
			machindeDeployment, kubeadmConfigTemplate, domachineTemplate := machineDeploymentGenerator.Generate(clusterNamespace, clusterName)
			createMachineDeployment(machindeDeployment, kubeadmConfigTemplate, domachineTemplate)

			By("Scale worker machine deployment")
			scaleMachineDeployment(machindeDeployment, 3)

			By("Delete worker machine deployment")
			deleteMachineDeployment(machindeDeployment, kubeadmConfigTemplate, domachineTemplate)

			By("Delete controlplane")
			deleteMachine(controlPlaneMachine, controlPlaneKubeadmconfig, controlPlaneDomachine)

			By("Delete cluster")
			deleteCluster(cluster, docluster)
		})
	})
})

func createCluster(cluster *clusterv1.Cluster, docluster *infrav1.DOCluster) {
	By("Creating a Cluster")
	Expect(kindclient.Create(context.TODO(), cluster)).To(Succeed())

	By("Creating a DOCluster")
	Expect(kindclient.Create(context.TODO(), docluster)).To(Succeed())

	By("Ensuring Cluster Infrastructure is Ready")
	WaitForClusterInfrastructureReady(kindclient, cluster.Namespace, cluster.Name)
}

func deleteCluster(cluster *clusterv1.Cluster, docluster *infrav1.DOCluster) {
	By("Deleting a DOCluster")
	Expect(kindclient.Delete(context.TODO(), docluster)).To(Succeed())
	WaitForDeletion(kindclient, docluster, docluster.Namespace, docluster.Name)

	By("Deleting a Cluster")
	Expect(kindclient.Delete(context.TODO(), cluster)).To(Succeed())
	WaitForDeletion(kindclient, cluster, cluster.Namespace, cluster.Name)
}

func createMachine(machine *clusterv1.Machine, kubeadmConfig *bootstrapkubeadmv1.KubeadmConfig, domachine *infrav1.DOMachine) {
	role := "worker"
	if util.IsControlPlaneMachine(machine) {
		role = "controlplane"
	}

	By(fmt.Sprintf("Creating %s KubeadmConfig", role))
	Expect(kindclient.Create(context.TODO(), kubeadmConfig)).To(Succeed())

	By(fmt.Sprintf("Creating %s DOMachine", role))
	Expect(kindclient.Create(context.TODO(), domachine)).To(Succeed())

	By(fmt.Sprintf("Creating %s Machine", role))
	Expect(kindclient.Create(context.TODO(), machine)).To(Succeed())

	By(fmt.Sprintf("Ensuring %s Machine Bootstrap is Ready", role))
	WaitForMachineBootstrapReady(kindclient, machine.Namespace, machine.Name)

	By(fmt.Sprintf("Ensuring %s DOMachine is Running", role))
	WaitForDOMachineRunning(kindclient, domachine.Namespace, domachine.Name)

	By(fmt.Sprintf("Ensuring %s DOMachine is Ready", role))
	WaitForDOMachineReady(kindclient, domachine.Namespace, domachine.Name)

	By(fmt.Sprintf("Ensuring %s MachineNodeRef is Already Set", role))
	WaitForMachineNodeRef(kindclient, machine.Namespace, machine.Name)
}

func deleteMachine(machine *clusterv1.Machine, kubeadmConfig *bootstrapkubeadmv1.KubeadmConfig, domachine *infrav1.DOMachine) {
	role := "worker"
	if util.IsControlPlaneMachine(machine) {
		role = "controlplane"
	}

	By(fmt.Sprintf("Deleting %s KubeadmConfig", role))
	Expect(kindclient.Delete(context.TODO(), kubeadmConfig)).To(Succeed())
	WaitForDeletion(kindclient, kubeadmConfig, kubeadmConfig.Namespace, kubeadmConfig.Name)

	By(fmt.Sprintf("Deleting %s DOMachine", role))
	Expect(kindclient.Delete(context.TODO(), domachine)).To(Succeed())
	WaitForDeletion(kindclient, domachine, domachine.Namespace, domachine.Name)

	By(fmt.Sprintf("Deleting %s Machine", role))
	Expect(kindclient.Delete(context.TODO(), machine)).To(Succeed())
	WaitForDeletion(kindclient, machine, machine.Namespace, machine.Name)
}

func createMachineDeployment(machineDeployment *clusterv1.MachineDeployment, kubeadmConfigTemplate *bootstrapkubeadmv1.KubeadmConfigTemplate, domachineTemplate *infrav1.DOMachineTemplate) {
	By("Creating kubeadmConfigTemplate")
	Expect(kindclient.Create(context.TODO(), kubeadmConfigTemplate)).To(Succeed())

	By("Creating domachineTemplate")
	Expect(kindclient.Create(context.TODO(), domachineTemplate)).To(Succeed())

	By("Creating machineDeployment")
	Expect(kindclient.Create(context.TODO(), machineDeployment)).To(Succeed())

	By("Ensuring machineDeployment running with expected replicas")
	WaitForMachineDeploymentRunning(kindclient, machineDeployment.Namespace, machineDeployment.Name, 1)
}

func scaleMachineDeployment(machineDeployment *clusterv1.MachineDeployment, replicas int32) {
	By("Scaling machineDeployment")
	md := &clusterv1.MachineDeployment{}
	Expect(kindclient.Get(context.TODO(), crclient.ObjectKey{Namespace: machineDeployment.Namespace, Name: machineDeployment.Name}, md))
	md.Spec.Replicas = &replicas
	Expect(kindclient.Update(context.TODO(), md)).NotTo(HaveOccurred())

	By("Ensuring machineDeployment running with expected replicas")
	WaitForMachineDeploymentRunning(kindclient, machineDeployment.Namespace, machineDeployment.Name, replicas)
}

func deleteMachineDeployment(machineDeployment *clusterv1.MachineDeployment, kubeadmConfigTemplate *bootstrapkubeadmv1.KubeadmConfigTemplate, domachineTemplate *infrav1.DOMachineTemplate) {
	By("Scaling down machineDeployment")
	scaleMachineDeployment(machineDeployment, 0)

	By("Deleting kubeadmConfigTemplate")
	Expect(kindclient.Delete(context.TODO(), kubeadmConfigTemplate)).To(Succeed())
	WaitForDeletion(kindclient, kubeadmConfigTemplate, kubeadmConfigTemplate.Namespace, kubeadmConfigTemplate.Name)

	By("Deleting domachineTemplate")
	Expect(kindclient.Delete(context.TODO(), domachineTemplate)).To(Succeed())
	WaitForDeletion(kindclient, domachineTemplate, domachineTemplate.Namespace, domachineTemplate.Name)

	By("Deleting machineDeployment")
	Expect(kindclient.Delete(context.TODO(), machineDeployment)).To(Succeed())
	WaitForDeletion(kindclient, machineDeployment, machineDeployment.Namespace, machineDeployment.Name)
}
