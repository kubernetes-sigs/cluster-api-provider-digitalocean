package ubuntu

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	doconfigv1 "github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/providerconfig/v1alpha1"

	"github.com/pmezard/go-difflib/difflib"
)

var update = flag.Bool("update", true, "update .golden files")

func TestNodeUserData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		cluster        *clusterv1.Cluster
		machine        *clusterv1.Machine
		providerConfig *doconfigv1.DigitalOceanMachineProviderConfig
	}{
		{
			name: "simple-master-userdata-test",
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cluster-1",
				},
				Status: clusterv1.ClusterStatus{
					APIEndpoints: []clusterv1.APIEndpoint{
						{
							Host: "0.0.0.0",
						},
					},
				},
			},
			machine: &clusterv1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-1",
				},
				Spec: clusterv1.MachineSpec{
					Versions: clusterv1.MachineVersionInfo{
						Kubelet:      "1.11.2",
						ControlPlane: "1.11.2",
					},
				},
			},
			providerConfig: &doconfigv1.DigitalOceanMachineProviderConfig{
				Image:         "ubuntu-18-04-x64",
				SSHPublicKeys: []string{"ssh-rsa AAAAA"},
			},
		},
		{
			name: "simple-node-userdata-test",
			cluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cluster-1",
				},
				Status: clusterv1.ClusterStatus{
					APIEndpoints: []clusterv1.APIEndpoint{
						{
							Host: "0.0.0.0",
						},
					},
				},
			},
			machine: &clusterv1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-2",
				},
				Spec: clusterv1.MachineSpec{
					Versions: clusterv1.MachineVersionInfo{
						Kubelet: "1.9.4",
					},
				},
			},
			providerConfig: &doconfigv1.DigitalOceanMachineProviderConfig{
				Image:         "ubuntu-18-04-x64",
				SSHPublicKeys: []string{"ssh-rsa BBBBB"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Provider{}
			userdata, err := p.UserData(tc.cluster, tc.machine, tc.providerConfig, "abcdef.1234567890abcdef")
			if err != nil {
				t.Fatal(err)
			}

			golden := filepath.Join("testdata", tc.name+".golden")
			if *update {
				ioutil.WriteFile(golden, []byte(userdata), 0644)
			}

			expected, err := ioutil.ReadFile(golden)
			if err != nil {
				t.Errorf("failed to read .golden file: %v", err)
			}

			diff := difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(expected)),
				B:        difflib.SplitLines(userdata),
				FromFile: "Fixture",
				ToFile:   "Current",
				Context:  3,
			}
			diffStr, err := difflib.GetUnifiedDiffString(diff)
			if err != nil {
				t.Fatal(err)
			}

			if diffStr != "" {
				t.Errorf("got diff between expected and actual result: \n%s\n", diffStr)
			}
		})
	}
}
