package ubuntu

import (
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/kubermatic/cluster-api-provider-digitalocean/pkg/containerruntime"
)

// Provider is userdata.Provider implementation.
type Provider struct{}

// SupportedContainerRuntimes return list of container runtimes
func (p Provider) SupportedContainerRuntimes() (runtimes []containerruntime.RuntimeInfo) {
	for _, ic := range dockerInstallCandidates {
		for _, v := range ic.versions {
			runtimes = append(runtimes, containerruntime.RuntimeInfo{Name: "docker", Version: v})
		}
	}

	return runtimes
}

func (p Provider) MasterUserData() (string, error) {
	return "", fmt.Errorf("TODO: Not yet implemented")
}

func (p Provider) NodeUserData(machine *clusterv1.Machine, bootstrapToken string) (string, error) {
	// Parse template for machine data
	//tmpl, err := template.New("user-data").Funcs(machinetemplate.TxtFuncMap()).Parse(userdataNodeTemplate)
	//if err != nil {
	//	return "", fmt.Errorf("failed to parse user-data template: %v", err)
	//}
	//
	//// Convert kubeletVersion struct to a semver struct.
	//kubeletVersion, err := semver.NewVersion(machine.Spec.Versions.Kubelet)
	//if err != nil {
	//	return "", fmt.Errorf("invalid kubelet version: %v", err)
	//}
	//
	//// Get name of the kubeadm DropIn file.
	//var kubeadmDropInFilename string
	//if kubeletVersion.Minor() > 8 {
	//	kubeadmDropInFilename = "10-kubeadm.conf"
	//} else {
	//	kubeadmDropInFilename = "kubeadm-10.conf"
	//}

	return "", fmt.Errorf("TODO: Not yet implemented.")
}
