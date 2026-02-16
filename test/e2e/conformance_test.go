//go:build e2e
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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	capi_e2e "sigs.k8s.io/cluster-api/test/e2e"
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
		result              *clusterctl.ApplyClusterTemplateAndWaitResult
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

		clusterName = os.Getenv("CLUSTER_NAME")
		if clusterName == "" {
			clusterName = fmt.Sprintf("capdo-conf-%s", util.RandomString(6))
		}
		fmt.Fprintf(GinkgoWriter, "INFO: Cluster name is %s\n", clusterName)

		// Setup a Namespace where to host objects for this spec and create a watcher for the namespace events.
		namespace, cancelWatches = setupSpecNamespace(ctx, specName, bootstrapClusterProxy, artifactFolder)

		clusterctlLogFolder = filepath.Join(artifactFolder, "clusters", bootstrapClusterProxy.GetName())

		result = new(clusterctl.ApplyClusterTemplateAndWaitResult)
	})

	AfterEach(func() {
		dumpSpecResourcesAndCleanup(ctx, specName, bootstrapClusterProxy, artifactFolder, namespace, cancelWatches, result.Cluster, e2eConfig.GetIntervals, skipCleanup, clusterctlConfigPath)
		redactLogs(e2eConfig.MustGetVariable)
	})

	It("Should run conformance tests", func() {
		var err error

		flavor := clusterctl.DefaultFlavor
		if useCIArtifacts {
			flavor = "conformance-ci-artifacts"
		}

		workerMachineCount, err := strconv.ParseInt(e2eConfig.MustGetVariable("CONFORMANCE_WORKER_MACHINE_COUNT"), 10, 64)
		Expect(err).NotTo(HaveOccurred())
		controlPlaneMachineCount, err := strconv.ParseInt(e2eConfig.MustGetVariable("CONFORMANCE_CONTROL_PLANE_MACHINE_COUNT"), 10, 64)
		Expect(err).NotTo(HaveOccurred())

		By("Initializes the work cluster")
		clusterctl.ApplyClusterTemplateAndWait(ctx, clusterctl.ApplyClusterTemplateAndWaitInput{
			ClusterProxy: bootstrapClusterProxy,
			ConfigCluster: clusterctl.ConfigClusterInput{
				LogFolder:                clusterctlLogFolder,
				ClusterctlConfigPath:     clusterctlConfigPath,
				KubeconfigPath:           bootstrapClusterProxy.GetKubeconfigPath(),
				InfrastructureProvider:   clusterctl.DefaultInfrastructureProvider,
				Flavor:                   flavor,
				Namespace:                namespace.Name,
				ClusterName:              clusterName,
				KubernetesVersion:        e2eConfig.MustGetVariable(capi_e2e.KubernetesVersion),
				ControlPlaneMachineCount: ptr.To[int64](controlPlaneMachineCount),
				WorkerMachineCount:       ptr.To[int64](workerMachineCount),
			},
			WaitForClusterIntervals:      e2eConfig.GetIntervals(specName, "wait-cluster"),
			WaitForControlPlaneIntervals: e2eConfig.GetIntervals(specName, "wait-control-plane"),
			WaitForMachineDeployments:    e2eConfig.GetIntervals(specName, "wait-worker-nodes"),
		}, result)

		workloadProxy := bootstrapClusterProxy.GetWorkloadCluster(ctx, namespace.Name, clusterName)
		kubetest.Run(context.Background(),
			kubetest.RunInput{
				ClusterProxy:   workloadProxy,
				NumberOfNodes:  int(workerMachineCount),
				ConfigFilePath: kubetestConfigFilePath,
			},
		)
	})
})
