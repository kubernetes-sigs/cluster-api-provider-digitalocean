//go:build e2e
// +build e2e

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

package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/digitalocean/godo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"
	"sigs.k8s.io/cluster-api/util"
)

var _ = Describe("Load balancer creation", func() {
	var (
		ctx                 = context.TODO()
		specName            = "create-lb"
		namespace           *corev1.Namespace
		cancelWatches       context.CancelFunc
		result              *clusterctl.ApplyClusterTemplateAndWaitResult
		clusterName         string
		clusterctlLogFolder string
		workloadClient      client.Client
		lbIP                string
	)

	BeforeEach(func() {
		Expect(e2eConfig).ToNot(BeNil(), "Invalid argument. e2eConfig can't be nil when calling %s spec", specName)
		Expect(clusterctlConfigPath).To(BeAnExistingFile(), "Invalid argument. clusterctlConfigPath must be an existing file when calling %s spec", specName)
		Expect(bootstrapClusterProxy).ToNot(BeNil(), "Invalid argument. bootstrapClusterProxy can't be nil when calling %s spec", specName)
		Expect(os.MkdirAll(artifactFolder, 0755)).To(Succeed(), "Invalid argument. artifactFolder can't be created for %s spec", specName)
		Expect(e2eConfig.Variables).To(HaveKey(KubernetesVersion))

		clusterName = fmt.Sprintf("capdo-e2e-lb-%s", util.RandomString(6))
		namespace, cancelWatches = setupSpecNamespace(ctx, specName, bootstrapClusterProxy, artifactFolder)
		clusterctlLogFolder = filepath.Join(artifactFolder, "clusters", bootstrapClusterProxy.GetName())
		result = new(clusterctl.ApplyClusterTemplateAndWaitResult)
		workloadClient = nil
		lbIP = ""
	})

	AfterEach(func() {
		// Delete the LoadBalancer Service before the cluster teardown so the CCM (running in the
		// workload cluster) removes the DigitalOcean LB. Deleting droplets first would kill the CCM
		// and orphan the LB, and CleanDOResources can't reap it since the CCM names LBs by service UID.
		if !skipCleanup && workloadClient != nil {
			By("Deleting the LoadBalancer Service so the CCM removes the DigitalOcean LB")
			// Log delete failures instead of asserting: a failed assertion here would abort the
			// AfterEach and skip the cluster teardown below, leaking droplets. NotFound is fine.
			if err := client.IgnoreNotFound(workloadClient.Delete(ctx, lbTestService())); err != nil {
				fmt.Fprintf(GinkgoWriter, "WARNING: failed to delete lb-test Service: %v\n", err)
			}
			if err := client.IgnoreNotFound(workloadClient.Delete(ctx, lbTestDeployment())); err != nil {
				fmt.Fprintf(GinkgoWriter, "WARNING: failed to delete lb-test Deployment: %v\n", err)
			}

			// CCM deletes the DO LB asynchronously; wait for it to be gone before droplets are
			// deleted below, otherwise the CCM dies mid-deletion and the LB leaks.
			if lbIP != "" {
				Eventually(func() (bool, error) {
					return doLoadBalancerExists(ctx, lbIP)
				}, e2eConfig.GetIntervals(specName, "wait-service")...).Should(BeFalse(), "DigitalOcean LB %s was not removed", lbIP)
			}
		}

		dumpSpecResourcesAndCleanup(ctx, specName, bootstrapClusterProxy, artifactFolder, namespace, cancelWatches, result.Cluster, e2eConfig.GetIntervals, skipCleanup, clusterctlConfigPath)
		redactLogs(e2eConfig.MustGetVariable)
	})

	It("Should provision a DigitalOcean load balancer for a Service of type LoadBalancer", func() {
		By("Creating a workload cluster")
		clusterctl.ApplyClusterTemplateAndWait(ctx, clusterctl.ApplyClusterTemplateAndWaitInput{
			ClusterProxy: bootstrapClusterProxy,
			ConfigCluster: clusterctl.ConfigClusterInput{
				LogFolder:                clusterctlLogFolder,
				ClusterctlConfigPath:     clusterctlConfigPath,
				KubeconfigPath:           bootstrapClusterProxy.GetKubeconfigPath(),
				InfrastructureProvider:   clusterctl.DefaultInfrastructureProvider,
				Flavor:                   clusterctl.DefaultFlavor,
				Namespace:                namespace.Name,
				ClusterName:              clusterName,
				KubernetesVersion:        e2eConfig.MustGetVariable(KubernetesVersion),
				ControlPlaneMachineCount: ptr.To[int64](1),
				WorkerMachineCount:       ptr.To[int64](1),
			},
			WaitForClusterIntervals:      e2eConfig.GetIntervals(specName, "wait-cluster"),
			WaitForControlPlaneIntervals: e2eConfig.GetIntervals(specName, "wait-control-plane"),
			WaitForMachineDeployments:    e2eConfig.GetIntervals(specName, "wait-worker-nodes"),
		}, result)

		workloadProxy := bootstrapClusterProxy.GetWorkloadCluster(ctx, namespace.Name, clusterName)
		workloadClient = workloadProxy.GetClient()

		By("Deploying an nginx Deployment backing the Service")
		deployment := lbTestDeployment()
		Expect(workloadClient.Create(ctx, deployment)).To(Succeed())
		framework.WaitForDeploymentsAvailable(ctx, framework.WaitForDeploymentsAvailableInput{
			Getter:     workloadClient,
			Deployment: deployment,
		}, e2eConfig.GetIntervals(specName, "wait-deployment")...)

		By("Creating a Service of type LoadBalancer")
		service := lbTestService()
		Expect(workloadClient.Create(ctx, service)).To(Succeed())

		By("Waiting for the cloud controller manager to assign an external IP")
		Eventually(func() (string, error) {
			svc := &corev1.Service{}
			if err := workloadClient.Get(ctx, types.NamespacedName{Namespace: service.Namespace, Name: service.Name}, svc); err != nil {
				return "", err
			}
			for _, ing := range svc.Status.LoadBalancer.Ingress {
				if ing.IP != "" {
					lbIP = ing.IP
					return ing.IP, nil
				}
			}
			return "", nil
		}, e2eConfig.GetIntervals(specName, "wait-service")...).ShouldNot(BeEmpty(), "Service did not get an external IP assigned")

		By("Verifying the load balancer exists in DigitalOcean")
		Expect(doLoadBalancerExists(ctx, lbIP)).To(BeTrue(), "no DigitalOcean load balancer found with IP %s", lbIP)
	})
})

func lbTestDeployment() *appsv1.Deployment {
	labels := map[string]string{"app": "lb-test"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "lb-test", Namespace: metav1.NamespaceDefault},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To[int32](1),
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "nginx",
						Image: "nginx:1.27",
						Ports: []corev1.ContainerPort{{ContainerPort: 80}},
					}},
				},
			},
		},
	}
}

func lbTestService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "lb-test", Namespace: metav1.NamespaceDefault},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{"app": "lb-test"},
			Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt32(80)}},
		},
	}
}

// doLoadBalancerExists lists DigitalOcean load balancers and reports whether one matches the given IP.
func doLoadBalancerExists(ctx context.Context, ip string) (bool, error) {
	token := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if token == "" {
		return false, fmt.Errorf("missing DO token")
	}
	lbs, err := lbList(ctx, godo.NewFromToken(token))
	if err != nil {
		return false, err
	}
	for _, lb := range lbs {
		if lb.IP == ip {
			return true, nil
		}
	}
	return false, nil
}
