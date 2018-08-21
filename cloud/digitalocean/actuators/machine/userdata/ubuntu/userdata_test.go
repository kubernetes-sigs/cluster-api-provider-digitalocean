package ubuntu

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	doconfigv1 "github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/providerconfig/v1alpha1"
)

func TestNodeUserData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		cluster        *clusterv1.Cluster
		machine        *clusterv1.Machine
		providerConfig *doconfigv1.DigitalOceanMachineProviderConfig
	}{
		{
			name: "simple nodeuserdata test",
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
						Kubelet: "1.9.4",
					},
				},
			},
			providerConfig: &doconfigv1.DigitalOceanMachineProviderConfig{
				Image:         "ubuntu-16-04-x64",
				SSHPublicKeys: []string{"test-key-1"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Provider{}
			_, err := p.NodeUserData(tc.cluster, tc.machine, tc.providerConfig, "123456.123456")
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
