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

	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var _ = logf.Log.WithName("domachine-resource")

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-domachine,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=domachines,versions=v1beta1,name=validation.domachine.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1beta1
// +kubebuilder:webhook:verbs=create;update,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-domachine,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=infrastructure.cluster.x-k8s.io,resources=domachines,versions=v1beta1,name=default.domachine.infrastructure.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1beta1

// DOMachineWebhook is a collection of webhooks for DOMachine objects.
type DOMachineWebhook struct{}

var (
	_ webhook.CustomDefaulter = &DOMachineWebhook{}
	_ webhook.CustomValidator = &DOMachineWebhook{}
)

func (w *DOMachineWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1beta1.DOMachine{}).
		WithDefaulter(&DOMachineWebhook{}).
		WithValidator(&DOMachineWebhook{}).
		Complete()
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (w *DOMachineWebhook) Default(context.Context, runtime.Object) error {
	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOMachineWebhook) ValidateCreate(context.Context, runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOMachineWebhook) ValidateUpdate(_ context.Context, objOld, objNew runtime.Object) (admission.Warnings, error) {
	var allErrs field.ErrorList

	newDOMachine, ok := objNew.(*v1beta1.DOMachine)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an DOMachine new object but got a %T", objNew))
	}

	newDOMachineUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(objNew)
	if err != nil {
		return nil, apierrors.NewInternalError(errors.Wrap(err, "failed to convert new DOMachine to unstructured object"))
	}

	oldDOMachineUnstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(objOld)
	if err != nil {
		return nil, apierrors.NewInternalError(errors.Wrap(err, "failed to convert old DOMachine to unstructured object"))
	}

	newDOMachineSpec := newDOMachineUnstr["spec"].(map[string]interface{})
	oldDOMachineSpec := oldDOMachineUnstr["spec"].(map[string]interface{})

	// allow changes to providerID
	delete(oldDOMachineSpec, "providerID")
	delete(newDOMachineSpec, "providerID")

	// allow changes to additionalTags
	delete(oldDOMachineSpec, "additionalTags")
	delete(newDOMachineSpec, "additionalTags")

	if !reflect.DeepEqual(oldDOMachineSpec, newDOMachineSpec) {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec"), "cannot be modified"))
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(newDOMachine.GroupVersionKind().GroupKind(), newDOMachine.Name, allErrs)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *DOMachineWebhook) ValidateDelete(context.Context, runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
