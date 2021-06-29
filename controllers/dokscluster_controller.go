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

package controllers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/predicates"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/kubernetes"
)

// DOKSClusterReconciler reconciles a DOKSCluster object
type DOKSClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksclusters/finalizers,verbs=update

// SetupWithManager sets up the controller with the Manager.
func (r *DOKSClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.DOKSCluster{}).
		WithEventFilter(predicates.ResourceNotPaused(ctrl.LoggerFrom(ctx))). // don't queue reconcile if resource is paused
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DOKSCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DOKSClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)

	dokscluster := &infrav1.DOKSCluster{}
	if err := r.Get(ctx, req.NamespacedName, dokscluster); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, dokscluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, nil
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}

	// Create the cluster scope.
	clusterScope, err := scope.NewDOKSClusterScope(scope.DOKSClusterScopeParams{
		Client:      r.Client,
		Logger:      log,
		Cluster:     cluster,
		DOKSCluster: dokscluster,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any changes.
	defer func() {
		if err := clusterScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !dokscluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, clusterScope)
	}

	// Handle new clusters
	kubernetessvc := kubernetes.NewService(ctx, clusterScope)
	docluster, err := kubernetessvc.GetCluster(clusterScope.GetInstanceID())
	if err != nil {
		return reconcile.Result{}, err
	}
	if docluster == nil {
		return r.reconcileCreate(ctx, clusterScope)
	}

	return r.reconcile(ctx, clusterScope)
}

func (r *DOKSClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.DOKSClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Deleting DOKSCluster")
	return reconcile.Result{}, nil
}

func (r *DOKSClusterReconciler) reconcileCreate(ctx context.Context, clusterScope *scope.DOKSClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Creating DOKSCluster")
	// TODO(zetaron): List DOKSNodePools by cluster Label

	clusterReq, err := labels.NewRequirement(
		"cluster.x-k8s.io/cluster-name",
		selection.Equals,
		[]string{clusterScope.Name()},
	)
	if err != nil {
		return reconcile.Result{}, errors.Errorf("bad requirement: %+v", err)
	}
	ls := labels.NewSelector()
	ls = ls.Add(*clusterReq)

	npl := &infrav1.DOKSNodePoolList{}
	lo := &client.ListOptions{
		LabelSelector: ls,
	}

	err = r.List(ctx, npl, lo)
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to list cluster node pools: %+v", err)
	}

	clusterScope.Info(fmt.Sprintf("Found %d MachinePools", len(npl.Items)))

	// TODO(zetaron): at least 1 machine pool
	// TODO(zetaron): all machine pools must own a node pool via infrastructureRef

	// TODO(zetaron): map MachinePools to godo.NodePool
	// TODO(zetaron): map Cluster to godo.KubernetesCluster

	return reconcile.Result{}, nil
}

func (r *DOKSClusterReconciler) reconcile(ctx context.Context, clusterScope *scope.DOKSClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling DOKSCluster")
	dokscluster := clusterScope.DOKSCluster
	// If the DOKSCluster doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(dokscluster, infrav1.DOKSClusterFinalizer)
	fmt.Println("my msg heres")

	return reconcile.Result{}, nil
}
