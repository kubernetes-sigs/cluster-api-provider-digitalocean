package generic

import (
	"fmt"

	"github.com/Masterminds/semver"
)

// InstallCandidate describes Docker installation candidate.
type InstallCandidate struct {
	Versions   []string
	PkgVersion string
	Pkg        string
}

// getOfficiallySupportedVersions returns list of officially supported docker version for the given Kubelet version.
func getOfficiallySupportedVersions(kubernetesVersion string) ([]string, error) {
	v, err := semver.NewVersion(kubernetesVersion)
	if err != nil {
		return nil, err
	}

	majorMinorString := fmt.Sprintf("%d.%d", v.Major(), v.Minor())
	switch majorMinorString {
	case "1.8", "1.9", "1.10", "1.11":
		return []string{"1.11.2", "1.12.6", "1.13.1", "17.03.2", "17.12", "18.03", "18.06"}, nil
	default:
		return nil, fmt.Errorf("no supported docker version for provided kubernetes version found")
	}
}

// GetDockerVersion returns Docker version based on provided install candidates and Kubelet version.
func GetDockerVersion(kubernetesVersion string, dockerInstallCandidates []InstallCandidate) (string, error) {
	defaultVersions, err := getOfficiallySupportedVersions(kubernetesVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get a officially supported docker version for the given kubelet version: %v", err)
	}

	var runtimes []string
	for _, ic := range dockerInstallCandidates {
		runtimes = append(runtimes, ic.Versions...)
	}

	var newVersion string
	for _, v := range defaultVersions {
		for _, sv := range runtimes {
			if sv == v {
				// we should not return asap as we prefer the highest supported version
				newVersion = sv
			}
		}
	}
	if newVersion == "" {
		return "", fmt.Errorf("no supported versions available for kubelet '%s'", kubernetesVersion)
	}

	return newVersion, nil
}
