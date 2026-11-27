/*
Copyright 2021 The Kubernetes Authors.

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

// Package webhooks contains implementations of validating and mutating webhooks for infrastructure.cluster.x-k8s.io resources.
package webhooks

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var _ = logf.Log.WithName("docluster-resource")

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-docluster,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=doclusters,versions=v1beta1,name=validation.docluster.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-docluster,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=doclusters,versions=v1beta1,name=default.docluster.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1

// DOClusterWebhook is a collection of webhooks for DOCluster objects.
type DOClusterWebhook struct{}

var (
	_ admission.Defaulter[*v1beta1.DOCluster] = &DOClusterWebhook{}
	_ admission.Validator[*v1beta1.DOCluster] = &DOClusterWebhook{}
)

func (w *DOClusterWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &v1beta1.DOCluster{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default implements admission.Defaulter so a webhook will be registered for the type.
func (w *DOClusterWebhook) Default(context.Context, *v1beta1.DOCluster) error {
	return nil
}

// ValidateCreate implements admission.Validator so a webhook will be registered for the type.
func (w *DOClusterWebhook) ValidateCreate(context.Context, *v1beta1.DOCluster) (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate implements admission.Validator so a webhook will be registered for the type.
func (w *DOClusterWebhook) ValidateUpdate(_ context.Context, oldDOCluster, newDOCluster *v1beta1.DOCluster) (admission.Warnings, error) {
	var allErrs field.ErrorList

	if newDOCluster.Spec.Region != oldDOCluster.Spec.Region {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "region"), newDOCluster.Spec.Region, "field is immutable"))
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(newDOCluster.GroupVersionKind().GroupKind(), newDOCluster.Name, allErrs)
}

// ValidateDelete implements admission.Validator so a webhook will be registered for the type.
func (w *DOClusterWebhook) ValidateDelete(context.Context, *v1beta1.DOCluster) (admission.Warnings, error) {
	return nil, nil
}
