package userdata

import (
	"errors"

	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/containerruntime"
)

var (
	ErrProviderNotFound = errors.New("no user data provider for the given os found")

	//providers = map[providerconfig.OperatingSystem]Provider{
	//	//providerconfig.OperatingSystemCoreos: coreos.Provider{},
	//	//providerconfig.OperatingSystemUbuntu: ubuntu.Provider{},
	//	//providerconfig.OperatingSystemCentOS: centos.Provider{},
	//}
)

//func ForOS(os providerconfig.OperatingSystem) (Provider, error) {
//	if p, found := providers[os]; found {
//		return p, nil
//	}
//	return nil, ErrProviderNotFound
//}

type Provider interface {
	MasterUserData() (string, error)
	NodeUserData() (string, error)
	SupportedContainerRuntimes() []containerruntime.RuntimeInfo
}
