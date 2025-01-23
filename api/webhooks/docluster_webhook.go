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

package webhooks

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var _ = logf.Log.WithName("docluster-resource")

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-docluster,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=doclusters,versions=v1beta1,name=validation.docluster.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1beta1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-docluster,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=doclusters,versions=v1beta1,name=default.docluster.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1beta1

type DOClusterWebhook struct{}

var (
	_ webhook.CustomDefaulter = &DOClusterWebhook{}
	_ webhook.CustomValidator = &DOClusterWebhook{}
)

func (w *DOClusterWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1beta1.DOCluster{}).
		WithDefaulter(&DOClusterWebhook{}).
		WithValidator(&DOClusterWebhook{}).
		Complete()
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (w *DOClusterWebhook) Default(context.Context, runtime.Object) error {
	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOClusterWebhook) ValidateCreate(context.Context, runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOClusterWebhook) ValidateUpdate(_ context.Context, objOld, objNew runtime.Object) (admission.Warnings, error) {
	var allErrs field.ErrorList

	oldDOCluster, ok := objOld.(*v1beta1.DOCluster)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an DOCluster old object but got a %T", objOld))
	}

	newDOCluster, ok := objNew.(*v1beta1.DOCluster)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an DOCluster new object but got a %T", objNew))
	}

	if newDOCluster.Spec.Region != oldDOCluster.Spec.Region {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec", "region"), newDOCluster.Spec.Region, "field is immutable"))
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(newDOCluster.GroupVersionKind().GroupKind(), newDOCluster.Name, allErrs)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOClusterWebhook) ValidateDelete(context.Context, runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
