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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var _ = logf.Log.WithName("domachinetemplate-resource")

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-domachinetemplate,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=domachinetemplates,versions=v1beta1,name=validation.domachinetemplate.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-domachinetemplate,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=domachinetemplates,versions=v1beta1,name=default.domachinetemplate.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1

// DOMachineTemplateWebhook is a collection of webhooks for DOMachineTemplate objects.
type DOMachineTemplateWebhook struct{}

var (
	_ admission.Defaulter[*v1beta1.DOMachineTemplate] = &DOMachineTemplateWebhook{}
	_ admission.Validator[*v1beta1.DOMachineTemplate] = &DOMachineTemplateWebhook{}
)

func (w *DOMachineTemplateWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &v1beta1.DOMachineTemplate{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default implements admission.Defaulter so a webhook will be registered for the type.
func (w *DOMachineTemplateWebhook) Default(_ context.Context, doMachineTemplate *v1beta1.DOMachineTemplate) error {
	defaultDOMachineSpec(&doMachineTemplate.Spec.Template.Spec)
	return nil
}

// ValidateCreate implements admission.Validator so a webhook will be registered for the type.
func (w *DOMachineTemplateWebhook) ValidateCreate(_ context.Context, doMachineTemplate *v1beta1.DOMachineTemplate) (admission.Warnings, error) {
	var allErrs field.ErrorList

	if doMachineTemplate.Spec.Template.Spec.ProviderID != nil {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "template", "spec", "providerID"), "cannot be set in templates"))
	}

	if len(allErrs) != 0 {
		return nil, apierrors.NewInvalid(doMachineTemplate.GroupVersionKind().GroupKind(), doMachineTemplate.Name, allErrs)
	}

	return privateDropletWarnings(doMachineTemplate.Spec.Template.Spec), nil
}

// ValidateUpdate implements admission.Validator so a webhook will be registered for the type.
func (w *DOMachineTemplateWebhook) ValidateUpdate(_ context.Context, oldDOMachineTemplate, newDOMachineTemplate *v1beta1.DOMachineTemplate) (admission.Warnings, error) {
	var allErrs field.ErrorList

	if !reflect.DeepEqual(newDOMachineTemplate.Spec, oldDOMachineTemplate.Spec) {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec"), "DOMachineTemplateSpec is immutable"))
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(newDOMachineTemplate.GroupVersionKind().GroupKind(), newDOMachineTemplate.Name, allErrs)
}

// ValidateDelete implements admission.Validator so a webhook will be registered for the type.
func (w *DOMachineTemplateWebhook) ValidateDelete(context.Context, *v1beta1.DOMachineTemplate) (admission.Warnings, error) {
	return nil, nil
}
