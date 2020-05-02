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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha2"

	bootstrapkubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha2"
	"sigs.k8s.io/cluster-api/test/helpers/components"
	capiFlag "sigs.k8s.io/cluster-api/test/helpers/flag"
	"sigs.k8s.io/cluster-api/test/helpers/kind"
	"sigs.k8s.io/cluster-api/test/helpers/scheme"
	"sigs.k8s.io/cluster-api/util"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CAPIVERSION  = "v0.2.9"
	CABPKVERSION = "v0.1.6"

	capiNamespace       = "capi-system"
	capiDeploymentName  = "capi-controller-manager"
	cabpkNamespace      = "cabpk-system"
	cabpkDeploymentName = "cabpk-controller-manager"
	capdoNamespace      = "capdo-system"
	capdoDeploymentName = "capdo-controller-manager"
	setupTimeout        = 10 * 60
)

var (
	managerImage      = capiFlag.DefineOrLookupStringFlag("managerImage", "", "Docker image to load into the kind cluster for testing")
	capiComponents    = capiFlag.DefineOrLookupStringFlag("capiComponents", "https://github.com/kubernetes-sigs/cluster-api/releases/download/"+CAPIVERSION+"/cluster-api-components.yaml", "URL to CAPI components to load")
	cabpkComponents   = capiFlag.DefineOrLookupStringFlag("cabpkComponents", "https://github.com/kubernetes-sigs/cluster-api-bootstrap-provider-kubeadm/releases/download/"+CABPKVERSION+"/bootstrap-components.yaml", "URL to CAPI components to load")
	capdoComponents   = capiFlag.DefineOrLookupStringFlag("capdoComponents", "", "capdo components to load")
	kustomizeBinary   = capiFlag.DefineOrLookupStringFlag("kustomizeBinary", "kustomize", "path to the kustomize binary")
	kubernetesVersion = capiFlag.DefineOrLookupStringFlag("kubernetesVersion", "v1.16.2", "kubernetes version to test on")

	kindcluster kind.Cluster
	kindclient  crclient.Client
	suiteTmpDir string
	testTmpDir  string

	credentials   string
	machineSize   string
	machineImage  string
	machineSSHKey string
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)

	// If running in prow, output the junit files to the artifacts path
	junitPath := fmt.Sprintf("junit.e2e_suite.%d.xml", config.GinkgoConfig.ParallelNode)
	artifactPath, exists := os.LookupEnv("ARTIFACTS")
	if exists {
		junitPath = path.Join(artifactPath, junitPath)
	}
	junitReporter := reporters.NewJUnitReporter(junitPath)
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	fmt.Fprintf(GinkgoWriter, "Verify required env variable\n")
	credentials = os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	Expect(credentials).NotTo(BeEmpty())
	machineSize = os.Getenv("MACHINE_TYPE")
	Expect(machineSize).NotTo(BeEmpty())
	machineImage = os.Getenv("MACHINE_IMAGE")
	Expect(machineImage).NotTo(BeEmpty())
	machineSSHKey = os.Getenv("MACHINE_SSHKEY")
	Expect(machineSSHKey).NotTo(BeEmpty())

	fmt.Fprintf(GinkgoWriter, "Setting up kind cluster\n")
	var err error
	suiteTmpDir, err = ioutil.TempDir("", "capdo-e2e-suite")
	Expect(err).NotTo(HaveOccurred())

	s := scheme.SetupScheme()
	Expect(bootstrapkubeadmv1.AddToScheme(s)).To(Succeed())
	Expect(infrav1.AddToScheme(s)).To(Succeed())

	kindcluster = kind.Cluster{
		Name: "capdo-e2e-test-" + util.RandomString(6),
	}

	kindcluster.Setup()
	kindcluster.LoadImage(*managerImage)

	kindclient, err = crclient.New(kindcluster.RestConfig(), crclient.Options{Scheme: s})
	Expect(err).NotTo(HaveOccurred())

	kindcluster.ApplyYAML(*capiComponents)
	kindcluster.ApplyYAML(*cabpkComponents)

	if capdoComponents == nil || *capdoComponents == "" {
		buildCAPDOComponents(capdoComponents)
	}
	kindcluster.ApplyYAML(*capdoComponents)

	components.WaitDeployment(kindclient, capiNamespace, capiDeploymentName)
	components.WaitDeployment(kindclient, cabpkNamespace, cabpkDeploymentName)
	components.WaitDeployment(kindclient, capdoNamespace, capdoDeploymentName)

	// Recreate kindclient so that it knows about the cluster api types
	kindclient, err = crclient.New(kindcluster.RestConfig(), crclient.Options{Scheme: s})
	Expect(err).NotTo(HaveOccurred())
}, setupTimeout)

var _ = AfterSuite(func() {
	fmt.Fprintf(GinkgoWriter, "Tearing down kind cluster\n")
	capiLogs := retrieveLogs(capiNamespace, capiDeploymentName)
	cabpkLogs := retrieveLogs(cabpkNamespace, cabpkDeploymentName)
	capdoLogs := retrieveLogs(capdoNamespace, capdoDeploymentName)

	// If running in prow, output the logs to the artifacts path
	artifactPath, exists := os.LookupEnv("ARTIFACTS")
	if exists {
		ioutil.WriteFile(path.Join(artifactPath, "capi.log"), []byte(capiLogs), 0644)   // nolint
		ioutil.WriteFile(path.Join(artifactPath, "cabpk.log"), []byte(cabpkLogs), 0644) // nolint
		ioutil.WriteFile(path.Join(artifactPath, "capdo.log"), []byte(capdoLogs), 0644) // nolint
		return
	}

	fmt.Fprintf(GinkgoWriter, "CAPI Logs:\n%s\n", capiLogs)
	fmt.Fprintf(GinkgoWriter, "CABPK Logs:\n%s\n", cabpkLogs)
	fmt.Fprintf(GinkgoWriter, "CAPDO Logs:\n%s\n", capdoLogs)

	kindcluster.Teardown()
	os.RemoveAll(suiteTmpDir)
})

func buildCAPDOComponents(manifest *string) {
	capdoManifests, err := exec.Command(*kustomizeBinary, "build", "../../config/default").Output() // nolint
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(GinkgoWriter, "Error: %s\n", string(exitError.Stderr))
		}
	}
	Expect(err).NotTo(HaveOccurred())

	// envsubst the credentials
	b64credentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	os.Setenv("DO_B64ENCODED_CREDENTIALS", b64credentials)
	manifestsContent := os.ExpandEnv(string(capdoManifests))

	// write out the manifests
	manifestFile := path.Join(suiteTmpDir, "infrastructure-components.yaml")
	Expect(ioutil.WriteFile(manifestFile, []byte(manifestsContent), 0600)).To(Succeed())
	*manifest = manifestFile
}

func buildCloudControllerManager(manifest *string) {
	ccmManifests, err := ioutil.ReadFile("manifests/digitalocean-cloud-controller-manager.yaml")
	Expect(err).NotTo(HaveOccurred())

	// envsubst the credentials
	manifestsContent := os.ExpandEnv(string(ccmManifests))
	manifestFile := path.Join(suiteTmpDir, "digitalocean-cloud-controller-manager.yaml")
	Expect(ioutil.WriteFile(manifestFile, []byte(manifestsContent), 0600)).To(Succeed())
	*manifest = manifestFile
}

func retrieveLogs(namespace, deploymentName string) string {
	deployment := &appsv1.Deployment{}
	Expect(kindclient.Get(context.TODO(), crclient.ObjectKey{Namespace: namespace, Name: deploymentName}, deployment)).To(Succeed())

	pods := &corev1.PodList{}

	selector, err := metav1.LabelSelectorAsMap(deployment.Spec.Selector)
	Expect(err).NotTo(HaveOccurred())

	Expect(kindclient.List(context.TODO(), pods, crclient.InNamespace(namespace), crclient.MatchingLabels(selector))).To(Succeed())
	Expect(pods.Items).NotTo(BeEmpty())

	clientset, err := kubernetes.NewForConfig(kindcluster.RestConfig())
	Expect(err).NotTo(HaveOccurred())

	podLogs, err := clientset.CoreV1().Pods(namespace).GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{Container: "manager"}).Stream()
	Expect(err).NotTo(HaveOccurred())
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	Expect(err).NotTo(HaveOccurred())

	return buf.String()
}
