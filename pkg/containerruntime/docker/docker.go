package docker

import (
	"fmt"

	"github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/actuators/machine/userdata"

	"github.com/Masterminds/semver"
)

// GetOfficiallySupportedVersions returns the officially supported docker version for the given kubernetes version
func getOfficiallySupportedVersions(kubernetesVersion string) ([]string, error) {
	v, err := semver.NewVersion(kubernetesVersion)
	if err != nil {
		return nil, err
	}

	majorMinorString := fmt.Sprintf("%d.%d", v.Major(), v.Minor())
	switch majorMinorString {
	case "1.8", "1.9", "1.10", "1.11":
		return []string{"1.11.2", "1.12.6", "1.13.1", "17.03.2"}, nil
	default:
		return nil, fmt.Errorf("no supported docker version for provided kubernetes version found")
	}
}

func GetDockerVersion(kubernetesVersion string, provider userdata.Provider) (string, error) {
	defaultVersions, err := getOfficiallySupportedVersions(kubernetesVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get a officially supported docker version for the given kubelet version: %v", err)
	}

	var newVersion string
	providerSupportedVersions := provider.SupportedContainerRuntimes()
	for _, v := range defaultVersions {
		for _, sv := range providerSupportedVersions {
			if sv.Version == v {
				// we should not return asap as we prefer the highest supported version
				newVersion = sv.Version
			}
		}
	}
	if newVersion == "" {
		return "", fmt.Errorf("no supported versions available for kubelet '%s'", kubernetesVersion)
	}

	return newVersion, nil
}
