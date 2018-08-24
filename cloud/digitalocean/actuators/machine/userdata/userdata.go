package userdata

import (
	"errors"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/actuators/machine/userdata/ubuntu"
	doconfigv1 "github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/providerconfig/v1alpha1"
)

var (
	ErrProviderNotFound = errors.New("no user data provider for the given os found")

	providers = map[string]Provider{
		"ubuntu-16-04-x64": ubuntu.Provider{},
		"ubuntu-18-04-x64": ubuntu.Provider{},
	}
)

func ForOS(os string) (Provider, error) {
	if p, found := providers[os]; found {
		return p, nil
	}
	return nil, ErrProviderNotFound
}

type Provider interface {
	UserData(cluster *clusterv1.Cluster, machine *clusterv1.Machine, providerConfig *doconfigv1.DigitalOceanMachineProviderConfig, bootstrapToken string) (string, error)
	GetDockerVersion(kubernetesVersion string) (string, error)
}
