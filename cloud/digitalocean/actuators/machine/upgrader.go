package machine

import (
	"fmt"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func (do *DOClient) upgradeCommandMasterControlPlane(machine *clusterv1.Machine) ([]string, error) {
	machineConfig, err := do.decodeMachineProviderConfig(machine.Spec.ProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("error decoding provided machineConfig: %v", err)
	}

	commandList := []string{}
	switch machineConfig.Image {
	case "ubuntu-18-04-x64":
		commandList = []string{
			fmt.Sprintf("sudo apt-get install kubeadm=%s-00", machine.Spec.Versions.ControlPlane),
			fmt.Sprintf("sudo kubeadm upgrade apply v%s -y", machine.Spec.Versions.ControlPlane),
		}
	default:
		return nil, fmt.Errorf("upgrade command list not available for image '%s'", machineConfig.Image)
	}

	return commandList, nil
}

func (do *DOClient) upgradeCommandMasterKubelet(machine *clusterv1.Machine) ([]string, error) {
	machineConfig, err := do.decodeMachineProviderConfig(machine.Spec.ProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("error decoding provided machineConfig: %v", err)
	}

	commandList := []string{}
	switch machineConfig.Image {
	case "ubuntu-18-04-x64":
		commandList = []string{
			fmt.Sprintf("sudo kubectl drain --kubeconfig=/etc/kubernetes/admin.conf --ignore-daemonsets %s", machine.Name),
			fmt.Sprintf("sudo apt-get install kubelet=%s-00", machine.Spec.Versions.Kubelet),
			fmt.Sprintf("sudo systemctl restart kubelet"),
			fmt.Sprintf("sudo kubeadm uncordon --kubeconfig=/etc/kubernetes/admin.conf %s", machine.Name),
		}
	default:
		return nil, fmt.Errorf("upgrade command list not available for image '%s'", machineConfig.Image)
	}

	return commandList, nil
}
