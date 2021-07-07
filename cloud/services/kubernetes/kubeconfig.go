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

package kubernetes

import (
	"context"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controlplanev1 "sigs.k8s.io/cluster-api-provider-digitalocean/controlplane/doks/api/v1alpha4"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/cluster-api/util/record"
	"sigs.k8s.io/cluster-api/util/secret"
)

// GetCluster get a cluster instance.
func (s *Service) ReconcileKubeconfig(ctx context.Context, controlplane *controlplanev1.DOKSControlPlane) error {
	s.scope.V(2).Info("Reconciling DOKS kubeconfig for cluster", "cluster-name", s.scope.Name())

	clusterRef := types.NamespacedName{
		Name:      s.scope.Cluster.Name,
		Namespace: s.scope.Cluster.Namespace,
	}

	_, err := secret.GetFromNamespacedName(ctx, s.scope.Client, clusterRef, secret.Kubeconfig)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get kubeconfig secret")
		}

		if createErr := s.createCAPIKubeconfigSecret(ctx, controlplane, &clusterRef); createErr != nil {
			return errors.Errorf("creating kubeconfig secret: %w", err)
		}
	} else if updateErr := s.updateCAPIKubeconfigSecret(ctx, controlplane, &clusterRef); updateErr != nil {
		return errors.Errorf("updating kubeconfig secret: %w", err)
	}

	return nil
}

func (s *Service) createCAPIKubeconfigSecret(ctx context.Context, controlplane *controlplanev1.DOKSControlPlane, clusterRef *types.NamespacedName) error {
	controllerOwnerRef := *metav1.NewControllerRef(controlplane, controlplanev1.GroupVersion.WithKind("DOKSControlPlane"))

	clusterKubeconfig, _, err := s.scope.Kubernetes.GetKubeConfig(ctx, s.scope.GetInstanceID())
	if err != nil {
		return errors.Wrap(err, "failed to aquire kubeconfig from DigitalOcean")
	}

	// kubeconfig secret name + namespace MUST be identical to the CAPI Cluster resource
	kubeconfigSecret := kubeconfig.GenerateSecretWithOwner(*clusterRef, clusterKubeconfig.KubeconfigYAML, controllerOwnerRef)
	if err := s.scope.Client.Create(ctx, kubeconfigSecret); err != nil {
		return errors.Wrap(err, "failed to create kubeconfig secret")
	}

	record.Eventf(controlplane, "SucessfulCreateKubeconfig", "Created kubeconfig for cluster %q", s.scope.Name())
	return nil
}

func (s *Service) updateCAPIKubeconfigSecret(ctx context.Context, controlplane *controlplanev1.DOKSControlPlane, clusterRef *types.NamespacedName) error {
	controllerOwnerRef := *metav1.NewControllerRef(controlplane, controlplanev1.GroupVersion.WithKind("DOKSControlPlane"))

	clusterKubeconfig, _, err := s.scope.Kubernetes.GetKubeConfig(ctx, s.scope.GetInstanceID())
	if err != nil {
		return errors.Wrap(err, "failed to aquire kubeconfig from DigitalOcean")
	}

	// kubeconfig secret name + namespace MUST be identical to the CAPI Cluster resource
	kubeconfigSecret := kubeconfig.GenerateSecretWithOwner(*clusterRef, clusterKubeconfig.KubeconfigYAML, controllerOwnerRef)
	if err := s.scope.Client.Update(ctx, kubeconfigSecret); err != nil {
		return errors.Wrap(err, "failed to update kubeconfig secret")
	}

	record.Eventf(controlplane, "SucessfulCreateKubeconfig", "Created kubeconfig for cluster %q", s.scope.Name())
	return nil
}
