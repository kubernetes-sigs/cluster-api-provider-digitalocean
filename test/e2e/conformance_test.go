// +build e2e

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
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi_e2e "sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"
	"sigs.k8s.io/cluster-api/test/framework/kubetest"
	"sigs.k8s.io/cluster-api/util"
)

var _ = Describe("Conformance Tests", func() {
	var (
		ctx                 = context.TODO()
		specName            = "conformance-tests"
		namespace           *corev1.Namespace
		cancelWatches       context.CancelFunc
		cluster             *clusterv1.Cluster
		clusterName         string
		clusterctlLogFolder string
	)

	BeforeEach(func() {
		Expect(ctx).NotTo(BeNil(), "ctx is required for %s spec", specName)
		Expect(e2eConfig).ToNot(BeNil(), "Invalid argument. e2eConfig can't be nil when calling %s spec", specName)
		Expect(clusterctlConfigPath).To(BeAnExistingFile(), "Invalid argument. clusterctlConfigPath must be an existing file when calling %s spec", specName)
		Expect(bootstrapClusterProxy).ToNot(BeNil(), "Invalid argument. bootstrapClusterProxy can't be nil when calling %s spec", specName)
		Expect(os.MkdirAll(artifactFolder, 0755)).To(Succeed(), "Invalid argument. artifactFolder can't be created for %s spec", specName)
		Expect(kubetestConfigFilePath).ToNot(BeNil(), "Invalid argument. kubetestConfigFilePath can't be nil")

		Expect(e2eConfig.Variables).To(HaveKey(KubernetesVersion))
		Expect(e2eConfig.Variables).To(HaveKey(capi_e2e.KubernetesVersion))
		Expect(e2eConfig.Variables).To(HaveKey(capi_e2e.CNIPath))

		clusterName = fmt.Sprintf("capdo-conf-%s", util.RandomString(6))

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

	Measure(specName, func(b Benchmarker) {
		var err error

		workerMachineCount, err := strconv.ParseInt(e2eConfig.GetVariable("CONFORMANCE_WORKER_MACHINE_COUNT"), 10, 64)
		Expect(err).NotTo(HaveOccurred())
		controlPlaneMachineCount, err := strconv.ParseInt(e2eConfig.GetVariable("CONFORMANCE_CONTROL_PLANE_MACHINE_COUNT"), 10, 64)
		Expect(err).NotTo(HaveOccurred())

		runtime := b.Time("cluster creation", func() {
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
					KubernetesVersion:        e2eConfig.GetVariable(capi_e2e.KubernetesVersion),
					ControlPlaneMachineCount: pointer.Int64Ptr(controlPlaneMachineCount),
					WorkerMachineCount:       pointer.Int64Ptr(workerMachineCount),
				},
				WaitForClusterIntervals:      e2eConfig.GetIntervals(specName, "wait-cluster"),
				WaitForControlPlaneIntervals: e2eConfig.GetIntervals(specName, "wait-control-plane"),
				WaitForMachineDeployments:    e2eConfig.GetIntervals(specName, "wait-worker-nodes"),
			})

			cluster = result.Cluster
		})

		b.RecordValue("cluster creation", runtime.Seconds())
		workloadProxy := bootstrapClusterProxy.GetWorkloadCluster(ctx, namespace.Name, clusterName)
		runtime = b.Time("conformance suite", func() {
			kubetest.Run(
				kubetest.RunInput{
					ClusterProxy:   workloadProxy,
					NumberOfNodes:  int(workerMachineCount),
					ConfigFilePath: kubetestConfigFilePath,
				},
			)
		})
		b.RecordValue("conformance suite run time", runtime.Seconds())
	}, 1)

	AfterEach(func() {
		/// Remove clusterctl apply log folder
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

})
