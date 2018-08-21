package ubuntu

import (
	"bytes"
	"fmt"
	"text/template"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	doconfigv1 "github.com/kubermatic/cluster-api-provider-digitalocean/cloud/digitalocean/providerconfig/v1alpha1"
	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/containerruntime"
	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/containerruntime/docker"
	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/machinetemplate"

	"github.com/Masterminds/semver"
)

// Provider is userdata.Provider implementation.
type Provider struct{}

// SupportedContainerRuntimes return list of container runtimes
func (p Provider) GetDockerVersion(kubernetesVersion string) (string, error) {
	defaultVersions, err := docker.GetOfficiallySupportedVersions(kubernetesVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get a officially supported docker version for the given kubelet version: %v", err)
	}

	var runtimes []containerruntime.RuntimeInfo
	for _, ic := range dockerInstallCandidates {
		for _, v := range ic.versions {
			runtimes = append(runtimes, containerruntime.RuntimeInfo{Name: "docker", Version: v})
		}
	}

	var newVersion string
	for _, v := range defaultVersions {
		for _, sv := range runtimes {
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

// NodeUserData generates cloud-init file for a Master.
func (p Provider) MasterUserData(cluster *clusterv1.Cluster, machine *clusterv1.Machine, providerConfig *doconfigv1.DigitalOceanMachineProviderConfig, bootstrapToken string) (string, error) {
	return "", fmt.Errorf("TODO: Not yet implemented")
}

// NodeUserData generates cloud-init file for a Node.
func (p Provider) NodeUserData(cluster *clusterv1.Cluster, machine *clusterv1.Machine, providerConfig *doconfigv1.DigitalOceanMachineProviderConfig, bootstrapToken string) (string, error) {
	// Parse template for machine data
	tmpl, err := template.New("user-data").Funcs(machinetemplate.TxtFuncMap()).Parse(userdataNodeTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse user-data template: %v", err)
	}

	// Convert kubeletVersion struct to a semver struct.
	kubeletVersion, err := semver.NewVersion(machine.Spec.Versions.Kubelet)
	if err != nil {
		return "", fmt.Errorf("invalid kubelet version: %v", err)
	}

	// Get name of the kubeadm DropIn file.
	var kubeadmDropInFilename string
	if kubeletVersion.Minor() > 8 {
		kubeadmDropInFilename = "10-kubeadm.conf"
	} else {
		kubeadmDropInFilename = "kubeadm-10.conf"
	}

	var crPkg, crPkgVersion string
	crVersion, err := p.GetDockerVersion(kubeletVersion.String())
	if err != nil {
		return "", fmt.Errorf("failed to get docker install candidate for %s: %v", machine.Spec.Versions.Kubelet, err)
	}
	crPkg, crPkgVersion, err = getDockerInstallCandidate(crVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get docker install candidate for %s: %v", machine.Spec.Versions.Kubelet, err)
	}

	// TODO: KubeadmCACertHash, CloudProvider, CloudConfig, OSConfig, ClusterDNSIPs.
	data := struct {
		MachineSpec           clusterv1.MachineSpec
		ProviderConfig        *doconfigv1.DigitalOceanMachineProviderConfig
		BoostrapToken         string
		CRAptPackage          string
		CRAptPackageVersion   string
		KubernetesVersion     string
		KubeadmDropInFilename string
		ServerAddr            string
	}{
		MachineSpec:           machine.Spec,
		ProviderConfig:        providerConfig,
		BoostrapToken:         bootstrapToken,
		CRAptPackage:          crPkg,
		CRAptPackageVersion:   crPkgVersion,
		KubernetesVersion:     kubeletVersion.String(),
		KubeadmDropInFilename: kubeadmDropInFilename,
		ServerAddr:            cluster.Status.APIEndpoints[0].Host,
	}
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute user-data template: %v", err)
	}

	return string(b.String()), nil
}
