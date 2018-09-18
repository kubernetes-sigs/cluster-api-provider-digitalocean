package ubuntu

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/docker/generic"
)

// RuntimeVersion implements RuntimeVersion interface.
type RuntimeVersion struct{}

// dockerInstallCandidates are Docker versions possible to install.
var dockerInstallCandidates = []generic.InstallCandidate{
	{
		Versions:   []string{"17.12", "17.12.1"},
		Pkg:        "docker.io",
		PkgVersion: "17.12.1-0ubuntu1",
	},
	{
		Versions:   []string{"18.03", "18.03.1"},
		Pkg:        "docker-ce",
		PkgVersion: "18.03.1~ce~3-0~ubuntu",
	},
	{
		Versions:   []string{"18.06.0"},
		Pkg:        "docker-ce",
		PkgVersion: "18.06.0~ce~3-0~ubuntu",
	},
	{
		Versions:   []string{"18.06", "18.06.1"},
		Pkg:        "docker-ce",
		PkgVersion: "18.06.1~ce~3-0~ubuntu",
	},
}

// GetDockerInstallCandidate returns Docker package and version cadidate.
func (d RuntimeVersion) GetDockerInstallCandidate(kubernetesVersion string) (string, string, error) {
	crVersion, err := generic.GetDockerVersion(kubernetesVersion, dockerInstallCandidates)
	if err != nil {
		return "", "", fmt.Errorf("failed to get docker install candidate for %s: %v", kubernetesVersion, err)
	}

	for _, ic := range dockerInstallCandidates {
		if sets.NewString(ic.Versions...).Has(crVersion) {
			return ic.Pkg, ic.PkgVersion, nil
		}
	}

	return "", "", fmt.Errorf("no install candidate available for the desired version")
}
