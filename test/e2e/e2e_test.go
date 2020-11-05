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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"
	"sigs.k8s.io/cluster-api/util"
)

var _ = Describe("Workload cluster creation", func() {
	var (
		ctx                 = context.TODO()
		specName            = "create-workload-cluster"
		namespace           *corev1.Namespace
		cancelWatches       context.CancelFunc
		cluster             *clusterv1.Cluster
		clusterName         string
		clusterctlLogFolder string
	)

	BeforeEach(func() {
		Expect(e2eConfig).ToNot(BeNil(), "Invalid argument. e2eConfig can't be nil when calling %s spec", specName)
		Expect(clusterctlConfigPath).To(BeAnExistingFile(), "Invalid argument. clusterctlConfigPath must be an existing file when calling %s spec", specName)
		Expect(bootstrapClusterProxy).ToNot(BeNil(), "Invalid argument. bootstrapClusterProxy can't be nil when calling %s spec", specName)
		Expect(os.MkdirAll(artifactFolder, 0755)).To(Succeed(), "Invalid argument. artifactFolder can't be created for %s spec", specName)

		Expect(e2eConfig.Variables).To(HaveKey(KubernetesVersion))

		clusterName = fmt.Sprintf("capdo-e2e-%s", util.RandomString(6))

		// Setup a Namespace where to host objects for this spec and create a watcher for the namespace events.
		namespace, cancelWatches = framework.CreateNamespaceAndWatchEvents(ctx, framework.CreateNamespaceAndWatchEventsInput{
			Creator:   bootstrapClusterProxy.GetClient(),
			ClientSet: bootstrapClusterProxy.GetClientSet(),
			Name:      fmt.Sprintf("%s-%s", specName, util.RandomString(6)),
			LogFolder: filepath.Join(artifactFolder, "clusters", bootstrapClusterProxy.GetName()),
		})

		// We need to override clusterctl apply log folder to avoid getting our credentials exposed.
		clusterctlLogFolder = filepath.Join(os.TempDir(), "clusters", bootstrapClusterProxy.GetName())
	})

	AfterEach(func() {
		// Remove clusterctl apply log folder
		Expect(os.RemoveAll(clusterctlLogFolder)).ShouldNot(HaveOccurred())

		// Dumps all the resources in the spec namespace, then cleanups the cluster object and the spec namespace itself.
		By(fmt.Sprintf("Dumping all the Cluster API resources in the %q namespace", namespace.Name))
		// Dump all Cluster API related resources to artifacts before deleting them.
		framework.DumpAllResources(ctx, framework.DumpAllResourcesInput{
			Lister:    bootstrapClusterProxy.GetClient(),
			Namespace: namespace.Name,
			LogPath:   filepath.Join(artifactFolder, "clusters", bootstrapClusterProxy.GetName(), "resources"),
		})

		if !skipCleanup {
			By(fmt.Sprintf("Deleting cluster %s/%s", cluster.Namespace, cluster.Name))
			// While https://github.com/kubernetes-sigs/cluster-api/issues/2955 is addressed in future iterations, there is a chance
			// that cluster variable is not set even if the cluster exists, so we are calling DeleteAllClustersAndWait
			// instead of DeleteClusterAndWait
			framework.DeleteAllClustersAndWait(ctx, framework.DeleteAllClustersAndWaitInput{
				Client:    bootstrapClusterProxy.GetClient(),
				Namespace: namespace.Name,
			}, e2eConfig.GetIntervals(specName, "wait-delete-cluster")...)

			By(fmt.Sprintf("Deleting namespace used for hosting the %q test spec", specName))
			framework.DeleteNamespace(ctx, framework.DeleteNamespaceInput{
				Deleter: bootstrapClusterProxy.GetClient(),
				Name:    namespace.Name,
			})

			// Will call the clean resources just to make sure we clean everything
			By(fmt.Sprintf("Making sure there is no leftover running for %s", cluster.Name))
			Expect(CleanDOResources(clusterName)).ShouldNot(HaveOccurred())
		}
		cancelWatches()
	})

	Context("Creating a single control-plane cluster", func() {
		It("Should create a cluster with 1 worker node and can be scaled", func() {
			By("Initializes with 1 worker node")
			result := clusterctl.ApplyClusterTemplateAndWait(ctx, clusterctl.ApplyClusterTemplateAndWaitInput{
				ClusterProxy: bootstrapClusterProxy,
				ConfigCluster: clusterctl.ConfigClusterInput{
					LogFolder:                clusterctlLogFolder,
					ClusterctlConfigPath:     clusterctlConfigPath,
					KubeconfigPath:           bootstrapClusterProxy.GetKubeconfigPath(),
					InfrastructureProvider:   clusterctl.DefaultInfrastructureProvider,
					Flavor:                   clusterctl.DefaultFlavor,
					Namespace:                namespace.Name,
					ClusterName:              clusterName,
					KubernetesVersion:        e2eConfig.GetVariable(KubernetesVersion),
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
					WorkerMachineCount:       pointer.Int64Ptr(1),
				},
				WaitForClusterIntervals:      e2eConfig.GetIntervals(specName, "wait-cluster"),
				WaitForControlPlaneIntervals: e2eConfig.GetIntervals(specName, "wait-control-plane"),
				WaitForMachineDeployments:    e2eConfig.GetIntervals(specName, "wait-worker-nodes"),
			})

			cluster = result.Cluster

			By("Scaling worker node to 3")
			result = clusterctl.ApplyClusterTemplateAndWait(ctx, clusterctl.ApplyClusterTemplateAndWaitInput{
				ClusterProxy: bootstrapClusterProxy,
				ConfigCluster: clusterctl.ConfigClusterInput{
					LogFolder:                clusterctlLogFolder,
					ClusterctlConfigPath:     clusterctlConfigPath,
					KubeconfigPath:           bootstrapClusterProxy.GetKubeconfigPath(),
					InfrastructureProvider:   clusterctl.DefaultInfrastructureProvider,
					Flavor:                   clusterctl.DefaultFlavor,
					Namespace:                namespace.Name,
					ClusterName:              clusterName,
					KubernetesVersion:        e2eConfig.GetVariable(KubernetesVersion),
					ControlPlaneMachineCount: pointer.Int64Ptr(1),
					WorkerMachineCount:       pointer.Int64Ptr(3),
				},
				WaitForClusterIntervals:      e2eConfig.GetIntervals(specName, "wait-cluster"),
				WaitForControlPlaneIntervals: e2eConfig.GetIntervals(specName, "wait-control-plane"),
				WaitForMachineDeployments:    e2eConfig.GetIntervals(specName, "wait-worker-nodes"),
			})

			cluster = result.Cluster
		})
	})

	Context("Creating a highly available control-plane cluster", func() {
		It("Should create a cluster with 3 control-plane and 2 worker nodes", func() {
			By("Creating a high available cluster")
			result := clusterctl.ApplyClusterTemplateAndWait(ctx, clusterctl.ApplyClusterTemplateAndWaitInput{
				ClusterProxy: bootstrapClusterProxy,
				ConfigCluster: clusterctl.ConfigClusterInput{
					LogFolder:                clusterctlLogFolder,
					ClusterctlConfigPath:     clusterctlConfigPath,
					KubeconfigPath:           bootstrapClusterProxy.GetKubeconfigPath(),
					InfrastructureProvider:   clusterctl.DefaultInfrastructureProvider,
					Flavor:                   clusterctl.DefaultFlavor,
					Namespace:                namespace.Name,
					ClusterName:              clusterName,
					KubernetesVersion:        e2eConfig.GetVariable(KubernetesVersion),
					ControlPlaneMachineCount: pointer.Int64Ptr(3),
					WorkerMachineCount:       pointer.Int64Ptr(2),
				},
				WaitForClusterIntervals:      e2eConfig.GetIntervals(specName, "wait-cluster"),
				WaitForControlPlaneIntervals: e2eConfig.GetIntervals(specName, "wait-control-plane"),
				WaitForMachineDeployments:    e2eConfig.GetIntervals(specName, "wait-worker-nodes"),
			})

			cluster = result.Cluster
		})
	})
})
