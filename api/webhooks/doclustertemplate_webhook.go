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
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var _ = logf.Log.WithName("doclustertemplate-resource")

func (w *DOClusterTemplateWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &v1beta1.DOClusterTemplate{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

//+kubebuilder:webhook:verbs=create;update,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-doclustertemplate,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=doclustertemplates,versions=v1beta1,name=default.doclustertemplate.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1
//+kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-doclustertemplate,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=doclustertemplates,versions=v1beta1,name=validation.doclustertemplate.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1

// DOClusterTemplateWebhook is a collection of webhooks for DOClusterTemplate objects.
type DOClusterTemplateWebhook struct{}

var (
	_ admission.Defaulter[*v1beta1.DOClusterTemplate] = &DOClusterTemplateWebhook{}
	_ admission.Validator[*v1beta1.DOClusterTemplate] = &DOClusterTemplateWebhook{}
)

// Default implements admission.Defaulter so a webhook will be registered for the type.
func (w *DOClusterTemplateWebhook) Default(context.Context, *v1beta1.DOClusterTemplate) error {
	return nil
}

// ValidateCreate implements admission.Validator so a webhook will be registered for the type.
func (w *DOClusterTemplateWebhook) ValidateCreate(context.Context, *v1beta1.DOClusterTemplate) (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate implements admission.Validator so a webhook will be registered for the type.
func (w *DOClusterTemplateWebhook) ValidateUpdate(_ context.Context, oldTemplate, newTemplate *v1beta1.DOClusterTemplate) (admission.Warnings, error) {
	if !reflect.DeepEqual(newTemplate.Spec, oldTemplate.Spec) {
		return nil, apierrors.NewBadRequest("DOClusterTemplate.Spec is immutable")
	}
	return nil, nil
}

// ValidateDelete implements admission.Validator so a webhook will be registered for the type.
func (w *DOClusterTemplateWebhook) ValidateDelete(context.Context, *v1beta1.DOClusterTemplate) (admission.Warnings, error) {
	return nil, nil
}
