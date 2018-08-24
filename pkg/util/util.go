package util

import (
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// IsMachineMaster checks is a provided machine master, by checking is ControlPlane version set.
func IsMachineMaster(machine *clusterv1.Machine) bool {
	return machine.Spec.Versions.ControlPlane != ""
}
