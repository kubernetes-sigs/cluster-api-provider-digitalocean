//go:build e2e
// +build e2e

/*
Copyright 2021 The Kubernetes Authors.

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
	"os"

	. "github.com/onsi/ginkgo"

	capi_e2e "sigs.k8s.io/cluster-api/test/e2e"
)

var _ = Describe("Running the Cluster API E2E tests", func() {
	BeforeEach(func() {

	})

	AfterEach(func() {
		redactLogs(e2eConfig.GetVariable)
	})

	Context("Running the quick-start spec", func() {
		capi_e2e.QuickStartSpec(context.TODO(), func() capi_e2e.QuickStartSpecInput {
			return capi_e2e.QuickStartSpecInput{
				E2EConfig:             e2eConfig,
				ClusterctlConfigPath:  clusterctlConfigPath,
				BootstrapClusterProxy: bootstrapClusterProxy,
				ArtifactFolder:        artifactFolder,
				SkipCleanup:           skipCleanup,
			}
		})
	})

	if os.Getenv("LOCAL_ONLY") != "true" {
		Context("API Version Upgrade", func() {
			Context("upgrade from v1alpha3 to v1beta1, and scale workload clusters created in v1alpha3 ", func() {
				BeforeEach(func() {
				})
				capi_e2e.ClusterctlUpgradeSpec(context.TODO(), func() capi_e2e.ClusterctlUpgradeSpecInput {
					return capi_e2e.ClusterctlUpgradeSpecInput{
						E2EConfig:             e2eConfig,
						ClusterctlConfigPath:  clusterctlConfigPath,
						BootstrapClusterProxy: bootstrapClusterProxy,
						ArtifactFolder:        artifactFolder,
						SkipCleanup:           skipCleanup,
					}
				})
			})

			Context("upgrade from v1alpha4 to v1beta1, and scale workload clusters created in v1alpha4", func() {
				BeforeEach(func() {
				})
				capi_e2e.ClusterctlUpgradeSpec(context.TODO(), func() capi_e2e.ClusterctlUpgradeSpecInput {
					return capi_e2e.ClusterctlUpgradeSpecInput{
						E2EConfig:                 e2eConfig,
						ClusterctlConfigPath:      clusterctlConfigPath,
						BootstrapClusterProxy:     bootstrapClusterProxy,
						ArtifactFolder:            artifactFolder,
						SkipCleanup:               skipCleanup,
						InitWithProvidersContract: "v1alpha4",
						InitWithBinary:            "https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.4.4/clusterctl-{OS}-{ARCH}",
					}
				})
			})
		})
	}
})
