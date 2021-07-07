/*
Copyright 2021.

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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	kubernetessvc "sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/kubernetes"
	controlplanev1 "sigs.k8s.io/cluster-api-provider-digitalocean/controlplane/doks/api/v1alpha4"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
)

// DOKSControlPlaneReconciler reconciles a DOKSControlPlane object
type DOKSControlPlaneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=dokscontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=dokscontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=dokscontrolplanes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DOKSControlPlane object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DOKSControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := log.FromContext(ctx)

	// Get the control plane instance
	doksControlPlane := &controlplanev1.DOKSControlPlane{}
	if err := r.Client.Get(ctx, req.NamespacedName, doksControlPlane); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Get the cluster
	cluster, err := util.GetOwnerCluster(ctx, r.Client, doksControlPlane.ObjectMeta)
	if err != nil {
		log.Error(err, "Failed to retrieve owner Cluster from the API Server")
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, err
	}

	if annotations.IsPaused(cluster, doksControlPlane) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, err
	}

	// Get the infrastructure cluster
	dokscluster := &infrav1.DOKSCluster{}
	doksclusterNamespacedName := client.ObjectKey{
		Namespace: cluster.Spec.InfrastructureRef.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Get(ctx, doksclusterNamespacedName, dokscluster); err != nil {
		log.Error(err, "Failed to retrieve Infrastructure Cluster from the API Server")
		return ctrl.Result{}, err
	}
	if dokscluster == nil {
		log.Info("DOKSCluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, err
	}

	doksControlPlaneScope, err := scope.NewDOKSControlPlaneScope(scope.DOKSControlPlaneScopeParams{
		Client:           r.Client,
		Cluster:          cluster,
		DOKSCluster:      dokscluster,
		DOKSControlPlane: doksControlPlane,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	defer func() {
		if err := doksControlPlaneScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted control plane
	if !doksControlPlane.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, doksControlPlaneScope)
	}

	return r.reconcile(ctx, doksControlPlaneScope)
}

func (r *DOKSControlPlaneReconciler) reconcileDelete(ctx context.Context, doksControlPlaneScope *scope.DOKSControlPlaneScope) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *DOKSControlPlaneReconciler) reconcile(ctx context.Context, doksControlPlaneScope *scope.DOKSControlPlaneScope) (ctrl.Result, error) {
	doksControlPlaneScope.Info("Reconciling DOKSControlPlane")

	dokscontrolplane := doksControlPlaneScope.DOKSControlPlane

	controllerutil.AddFinalizer(dokscontrolplane, controlplanev1.DOKSControlPlaneFinalizer)
	if err := doksControlPlaneScope.PatchObject(); err != nil {
		return ctrl.Result{}, err
	}

	doksClusterScope, err := doksControlPlaneScope.ToDOKSClusterScope()
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to derive DOKSClusterScope from DOKSControlPlaneScope", err)
	}

	kubernetesService := kubernetessvc.NewService(ctx, doksClusterScope)
	if err := kubernetesService.ReconcileKubeconfig(ctx, dokscontrolplane); err != nil {
		return ctrl.Result{}, errors.Errorf("failed to reconcile kubeconfig for DOKSControlPlane %s/%s: %w", dokscontrolplane.Namespace, dokscontrolplane.Name, err)
	}

	doksControlPlaneScope.SetInitialized()

	if doksClusterScope.DOKSCluster.Status.Ready {
		doksControlPlaneScope.SetReady()
	} else {
		doksControlPlaneScope.SetNotReady()
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DOKSControlPlaneReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	log := ctrl.LoggerFrom(ctx)
	doksManagedControlPlane := &controlplanev1.DOKSControlPlane{}

	c, err := ctrl.NewControllerManagedBy(mgr).
		For(doksManagedControlPlane).
		WithEventFilter(predicates.ResourceNotPaused(log)).
		Watches(
			&source.Kind{Type: &infrav1.DOKSCluster{}},
			handler.EnqueueRequestsFromMapFunc(r.managedClusterToManagedControlPlane(log)),
		).
		Build(r)
	if err != nil {
		return fmt.Errorf("failed setting up the DOKSControlPlane controller manager: %w", err)
	}

	if err := c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(util.ClusterToInfrastructureMapFunc(doksManagedControlPlane.GroupVersionKind())),
		predicates.ClusterUnpausedAndInfrastructureReady(log),
	); err != nil {
		return fmt.Errorf("failed adding Watch for ready Clusters: %w", err)
	}

	return nil
}

func (r *DOKSControlPlaneReconciler) managedClusterToManagedControlPlane(log logr.Logger) handler.MapFunc {
	return func(o client.Object) []ctrl.Request {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		doksCluster, ok := o.(*infrav1.DOKSCluster)
		if !ok {
			panic(fmt.Sprintf("Expected a DOKSCluster but got a %T", o))
		}

		if !doksCluster.ObjectMeta.DeletionTimestamp.IsZero() {
			log.V(4).Info("DOKSCluster has a deletion timestamp, skipping mapping")
			return nil
		}

		cluster, err := util.GetOwnerCluster(ctx, r.Client, doksCluster.ObjectMeta)
		if err != nil {
			log.Error(err, "failed to get owning cluster")
			return nil
		}
		if cluster == nil {
			log.V(4).Info("Owning cluster not set on DOKSCluster, skipping mapping")
			return nil
		}

		controlPlaneRef := cluster.Spec.ControlPlaneRef
		if controlPlaneRef == nil || controlPlaneRef.Kind != "DOKSControlPlane" {
			log.V(4).Info("ControlPlaneRef is nil or not DOKSControlPlane, skipping mapping")
			return nil
		}

		return []ctrl.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      controlPlaneRef.Name,
					Namespace: controlPlaneRef.Namespace,
				},
			},
		}
	}
}
