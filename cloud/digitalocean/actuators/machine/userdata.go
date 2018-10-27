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
	"bytes"
	"encoding/base64"
	"fmt"
	"text/template"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/cert"

	"sigs.k8s.io/cluster-api-provider-digitalocean/pkg/docker"

	"github.com/golang/glog"
)

// userdataParams are parameters used to parse the environment variables for bootstrap scripts.
type userdataParams struct {
	Machine        *clusterv1.Machine
	Cluster        *clusterv1.Cluster
	Token          string
	MasterEndpoint string
	PodCIDR        string
	ServiceCIDR    string
	CRPackage      string
	CRVersion      string
	CACertificate  string
	CAPrivateKey   string
}

// masterUserdata builds master bootstrap script based on script provided in providerConfig and based on environment template.
func masterUserdata(cluster *clusterv1.Cluster, machine *clusterv1.Machine, certificateAuthority *cert.CertificateAuthority, osImage, token, metadata string) (string, error) {
	var crPkg, crPkgVersion string
	dockerVersion, err := docker.ForOS(osImage)
	if err == nil {
		crPkg, crPkgVersion, err = dockerVersion.GetDockerInstallCandidate(machine.Spec.Versions.Kubelet)
		if err != nil {
			return "", fmt.Errorf("failed to get docker install candidate for %s: %v", machine.Spec.Versions.Kubelet, err)
		}
	} else {
		glog.Info(err)
	}

	var caCertificate, caPrivateKey string
	if certificateAuthority != nil {
		caCertificate = base64.StdEncoding.EncodeToString(certificateAuthority.Certificate)
		caPrivateKey = base64.StdEncoding.EncodeToString(certificateAuthority.PrivateKey)
	}

	params := userdataParams{
		Cluster:       cluster,
		Machine:       machine,
		Token:         token,
		PodCIDR:       subnet(cluster.Spec.ClusterNetwork.Pods),
		ServiceCIDR:   subnet(cluster.Spec.ClusterNetwork.Services),
		CRPackage:     crPkg,
		CRVersion:     crPkgVersion,
		CACertificate: caCertificate,
		CAPrivateKey:  caPrivateKey,
	}
	tmpl := template.Must(template.New("masterEnvironment").Parse(masterEnvironmentVariables))

	b := &bytes.Buffer{}
	err = tmpl.Execute(b, params)
	if err != nil {
		return "", fmt.Errorf("failed to execute user-data template: %v", err)
	}
	b.Write([]byte(metadata))

	return b.String(), nil
}

// nodeUserdata builds node bootstrap script based on script provided in providerConfig and based on environment template.
func nodeUserdata(cluster *clusterv1.Cluster, machine *clusterv1.Machine, osImage, token, metadata string) (string, error) {
	var crPkg, crPkgVersion string
	dockerVersion, err := docker.ForOS(osImage)
	if err == nil {
		crPkg, crPkgVersion, err = dockerVersion.GetDockerInstallCandidate(machine.Spec.Versions.Kubelet)
		if err != nil {
			return "", fmt.Errorf("failed to get docker install candidate for %s: %v", machine.Spec.Versions.Kubelet, err)
		}
	} else {
		glog.Info(err)
	}

	params := userdataParams{
		Cluster:        cluster,
		Machine:        machine,
		Token:          token,
		MasterEndpoint: endpoint(cluster.Status.APIEndpoints[0]),
		PodCIDR:        subnet(cluster.Spec.ClusterNetwork.Pods),
		ServiceCIDR:    subnet(cluster.Spec.ClusterNetwork.Services),
		CRPackage:      crPkg,
		CRVersion:      crPkgVersion,
	}
	tmpl := template.Must(template.New("nodeEnvironment").Parse(nodeEnvironmentVariables))

	b := &bytes.Buffer{}
	err = tmpl.Execute(b, params)
	if err != nil {
		return "", fmt.Errorf("failed to execute user-data template: %v", err)
	}
	b.Write([]byte(metadata))

	return b.String(), nil
}

func endpoint(apiEndpoint clusterv1.APIEndpoint) string {
	return fmt.Sprintf("%s:%d", apiEndpoint.Host, apiEndpoint.Port)
}

// subnet gets first IP of the subnet.
func subnet(netRange clusterv1.NetworkRanges) string {
	if len(netRange.CIDRBlocks) == 0 {
		return ""
	}
	return netRange.CIDRBlocks[0]
}

const (
	// masterEnvironment is the environment variables template for master instances.
	masterEnvironmentVariables = `#!/bin/bash
KUBELET_VERSION={{ .Machine.Spec.Versions.Kubelet }}
TOKEN={{ .Token }}
PORT=443
NAMESPACE={{ .Machine.ObjectMeta.Namespace }}
MACHINE=$NAMESPACE
MACHINE+="/"
MACHINE+={{ .Machine.ObjectMeta.Name }}
CONTROL_PLANE_VERSION={{ .Machine.Spec.Versions.ControlPlane }}
CLUSTER_DNS_DOMAIN={{ .Cluster.Spec.ClusterNetwork.ServiceDomain }}
POD_CIDR={{ .PodCIDR }}
SERVICE_CIDR={{ .ServiceCIDR }}
MASTER_CA_CERTIFICATE={{ .CACertificate }}
MASTER_CA_PRIVATE_KEY={{ .CAPrivateKey }}
CR_PACKAGE={{ .CRPackage }}
CR_VERSION={{ .CRVersion }}
`
	nodeEnvironmentVariables = `#!/bin/bash
KUBELET_VERSION={{ .Machine.Spec.Versions.Kubelet }}
MASTER={{ .MasterEndpoint }}
TOKEN={{ .Token }}
NAMESPACE={{ .Machine.ObjectMeta.Namespace }}
MACHINE=$NAMESPACE
MACHINE+="/"
MACHINE+={{ .Machine.ObjectMeta.Name }}
CLUSTER_DNS_DOMAIN={{ .Cluster.Spec.ClusterNetwork.ServiceDomain }}
POD_CIDR={{ .PodCIDR }}
SERVICE_CIDR={{ .ServiceCIDR }}
CR_PACKAGE={{ .CRPackage }}
CR_VERSION={{ .CRVersion }}
`
)
