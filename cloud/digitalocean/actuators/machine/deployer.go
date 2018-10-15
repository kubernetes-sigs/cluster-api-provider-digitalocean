/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machine

import (
	"errors"
	"fmt"
	"strings"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	apiutil "sigs.k8s.io/cluster-api/pkg/util"
)

// GetIP returns public IP address of the node in the cluster.
func (do *DOClient) GetIP(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (string, error) {
	droplet, err := do.instanceExists(machine)
	if err != nil {
		return "", err
	}

	if droplet == nil {
		return "", fmt.Errorf("instance %s doesn't exist", droplet.Name)
	}

	return droplet.PublicIPv4()
}

// GetKubeConfig returns kubeconfig from the master.
func (do *DOClient) GetKubeConfig(cluster *clusterv1.Cluster, master *clusterv1.Machine) (string, error) {
	droplet, err := do.instanceExists(master)
	if err != nil {
		return "", err
	}
	if droplet == nil {
		return "", errors.New("instance does not exists")
	}
	if len(cluster.Status.APIEndpoints) == 0 {
		return "", errors.New("unable to find cluster api endpoint address")
	}

	// We're using system SSH to download kubeconfig file from master.
	// The reasons for that are:
	//   * GetKubeConfig is executed by clusterctl on local machine. We don't know location of SSH key used for
	//     authentication and we can't add clusterctl flag for SSH key path, as it may not be possible to pass it to
	//     machine-controller and GetKubeConfig function.
	//   * Because we don't have SSH key, we can't use Go SSH implementation here.
	//   * Using Go SSH implementation and SSH agent together misbehaves, so is not an option.
	// Therefore, we're using system SSH here, so it correctly handles keys and authentication.
	result := strings.TrimSpace(apiutil.ExecCommand(
		"ssh", "-q",
		"-o", "StrictHostKeyChecking no",
		"-o", "UserKnownHostsFile /dev/null",
		fmt.Sprintf("%s@%s", "root", cluster.Status.APIEndpoints[0].Host),
		"echo STARTFILE; sudo cat /etc/kubernetes/admin.conf"))
	kubeconfig := strings.Split(result, "STARTFILE")

	if len(kubeconfig) < 2 {
		return "", errors.New("kubeconfig not available")
	}

	return strings.TrimSpace(kubeconfig[1]), nil
}
