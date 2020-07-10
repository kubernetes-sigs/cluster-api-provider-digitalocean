/*
Copyright 2020 The Kubernetes Authors.

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

package e2e

import (
	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha3"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	bootstrapkubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/types/v1beta1"
	"sigs.k8s.io/cluster-api/util"
)

type ClusterGenerator struct{}

func (gen ClusterGenerator) Generate(clusterNamespace, clusterName string) (*clusterv1.Cluster, *infrav1.DOCluster) {
	clusterRegion := "nyc1"
	docluster := &infrav1.DOCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
			Name:      clusterName,
		},
		Spec: infrav1.DOClusterSpec{
			Region: clusterRegion,
		},
	}

	cluster := &clusterv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
			Name:      clusterName,
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{},
			InfrastructureRef: &corev1.ObjectReference{
				APIVersion: infrav1.GroupVersion.String(),
				Kind:       TypeToKind(docluster),
				Namespace:  docluster.GetNamespace(),
				Name:       docluster.GetName(),
			},
		},
	}

	return cluster, docluster
}

type MachineGenerator struct{}

func (gen MachineGenerator) Generate(namespace, clusterName string, isControlPlane bool) (*clusterv1.Machine, *bootstrapkubeadmv1.KubeadmConfig, *infrav1.DOMachine) {
	name := clusterName + "-node-" + util.RandomString(6)
	if isControlPlane {
		name = clusterName + "-controlplane-" + util.RandomString(6)
	}

	kubernetesVersion := *kubernetesVersion
	kubeadmconfig := &bootstrapkubeadmv1.KubeadmConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: bootstrapkubeadmv1.KubeadmConfigSpec{
			InitConfiguration: &kubeadmv1beta1.InitConfiguration{
				NodeRegistration: kubeadmv1beta1.NodeRegistrationOptions{
					Name: `{{ ds.meta_data["local_hostname"] }}`,
					KubeletExtraArgs: map[string]string{
						"cloud-provider": "external",
						"provider-id":    `digitalocean://{{ ds.meta_data["instance_id"] }}`,
					},
				},
			},
			JoinConfiguration: &kubeadmv1beta1.JoinConfiguration{
				NodeRegistration: kubeadmv1beta1.NodeRegistrationOptions{
					Name: `{{ ds.meta_data["local_hostname"] }}`,
					KubeletExtraArgs: map[string]string{
						"cloud-provider": "external",
						"provider-id":    `digitalocean://{{ ds.meta_data["instance_id"] }}`,
					},
				},
			},
		},
	}

	domachine := &infrav1.DOMachine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: infrav1.DOMachineSpec{
			Size:           machineSize,
			Image:          intstr.Parse(machineImage),
			SSHKeys:        []intstr.IntOrString{intstr.Parse(machineSSHKey)},
			AdditionalTags: infrav1.Tags{"e2e-test"},
		},
	}

	machine := &clusterv1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: clusterv1.MachineSpec{
			ClusterName: clusterName,
			Version:     &kubernetesVersion,
			Bootstrap: clusterv1.Bootstrap{
				ConfigRef: &corev1.ObjectReference{
					APIVersion: bootstrapkubeadmv1.GroupVersion.String(),
					Kind:       TypeToKind(kubeadmconfig),
					Namespace:  kubeadmconfig.GetNamespace(),
					Name:       kubeadmconfig.GetName(),
				},
			},
			InfrastructureRef: corev1.ObjectReference{
				APIVersion: infrav1.GroupVersion.String(),
				Kind:       TypeToKind(domachine),
				Namespace:  domachine.GetNamespace(),
				Name:       domachine.GetName(),
			},
		},
	}

	if isControlPlane {
		machine.ObjectMeta.Labels = labels.Merge(machine.ObjectMeta.Labels, map[string]string{
			clusterv1.MachineControlPlaneLabelName: "true",
		})
	}

	return machine, kubeadmconfig, domachine
}

type MachineDeploymentGenerator struct{}

func (gen MachineDeploymentGenerator) Generate(namespace, clusterName string) (*clusterv1.MachineDeployment, *bootstrapkubeadmv1.KubeadmConfigTemplate, *infrav1.DOMachineTemplate) {
	name := clusterName + "-md-0-" + util.RandomString(6)
	kubeadmConfigTemplte := &bootstrapkubeadmv1.KubeadmConfigTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: bootstrapkubeadmv1.KubeadmConfigTemplateSpec{
			Template: bootstrapkubeadmv1.KubeadmConfigTemplateResource{
				Spec: bootstrapkubeadmv1.KubeadmConfigSpec{
					JoinConfiguration: &kubeadmv1beta1.JoinConfiguration{
						NodeRegistration: kubeadmv1beta1.NodeRegistrationOptions{
							Name: `{{ ds.meta_data["local_hostname"] }}`,
							KubeletExtraArgs: map[string]string{
								"cloud-provider": "external",
								"provider-id":    `digitalocean://{{ ds.meta_data["instance_id"] }}`,
							},
						},
					},
				},
			},
		},
	}

	domachineTemplate := &infrav1.DOMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: infrav1.DOMachineTemplateSpec{
			Template: infrav1.DOMachineTemplateResource{
				Spec: infrav1.DOMachineSpec{
					Size:           machineSize,
					Image:          intstr.Parse(machineImage),
					SSHKeys:        []intstr.IntOrString{intstr.Parse(machineSSHKey)},
					AdditionalTags: infrav1.Tags{"e2e-test"},
				},
			},
		},
	}

	replicas := int32(1)
	machineDeployment := &clusterv1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				"nodepool": "nodepool-0",
			},
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: clusterName,
			Replicas:    &replicas,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"nodepool": "nodepool-0",
				},
			},
			Template: clusterv1.MachineTemplateSpec{
				ObjectMeta: clusterv1.ObjectMeta{
					Namespace: namespace,
					Labels: map[string]string{
						"nodepool": "nodepool-0",
					},
				},
				Spec: clusterv1.MachineSpec{
					ClusterName: clusterName,
					Version:     kubernetesVersion,
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &corev1.ObjectReference{
							APIVersion: bootstrapkubeadmv1.GroupVersion.String(),
							Kind:       TypeToKind(kubeadmConfigTemplte),
							Namespace:  kubeadmConfigTemplte.GetNamespace(),
							Name:       kubeadmConfigTemplte.GetName(),
						},
					},
					InfrastructureRef: corev1.ObjectReference{
						APIVersion: infrav1.GroupVersion.String(),
						Kind:       TypeToKind(domachineTemplate),
						Namespace:  domachineTemplate.GetNamespace(),
						Name:       domachineTemplate.GetName(),
					},
				},
			},
		},
	}

	return machineDeployment, kubeadmConfigTemplte, domachineTemplate
}
