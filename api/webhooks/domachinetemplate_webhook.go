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
	"reflect"

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
var _ = logf.Log.WithName("domachinetemplate-resource")

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-domachinetemplate,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=domachinetemplates,versions=v1beta1,name=validation.domachinetemplate.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1beta1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-domachinetemplate,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=domachinetemplates,versions=v1beta1,name=default.domachinetemplate.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1beta1

// DOMachineTemplateWebhook is a collection of webhooks for DOMachineTemplate objects.
type DOMachineTemplateWebhook struct{}

var (
	_ webhook.CustomDefaulter = &DOMachineTemplateWebhook{}
	_ webhook.CustomValidator = &DOMachineTemplateWebhook{}
)

func (w *DOMachineTemplateWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1beta1.DOMachineTemplate{}).
		WithDefaulter(&DOMachineTemplateWebhook{}).
		WithValidator(&DOMachineTemplateWebhook{}).
		Complete()
}

// Default implements webhookutil.CustomDefaulter so a webhook will be registered for the type.
func (w *DOMachineTemplateWebhook) Default(context.Context, runtime.Object) error {
	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOMachineTemplateWebhook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	var allErrs field.ErrorList

	doMachineTemplate, ok := obj.(*v1beta1.DOMachineTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an DOMachineTemplate object but got a %T", obj))
	}

	if doMachineTemplate.Spec.Template.Spec.ProviderID != nil {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec", "template", "spec", "providerID"), "cannot be set in templates"))
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(doMachineTemplate.GroupVersionKind().GroupKind(), doMachineTemplate.Name, allErrs)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOMachineTemplateWebhook) ValidateUpdate(_ context.Context, objOld, objNew runtime.Object) (admission.Warnings, error) {
	var allErrs field.ErrorList

	oldDOMachineTemplate, ok := objOld.(*v1beta1.DOMachineTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an DOMachineTemplate old object but got a %T", objOld))
	}

	newDOMachineTemplate, ok := objNew.(*v1beta1.DOMachineTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an DOMachineTemplate new object but got a %T", objNew))
	}

	if !reflect.DeepEqual(newDOMachineTemplate.Spec, oldDOMachineTemplate.Spec) {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec"), "DOMachineTemplateSpec is immutable"))
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(newDOMachineTemplate.GroupVersionKind().GroupKind(), newDOMachineTemplate.Name, allErrs)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOMachineTemplateWebhook) ValidateDelete(context.Context, runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
