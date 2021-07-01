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
	"time"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	exp_util "sigs.k8s.io/cluster-api/exp/util"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/predicates"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/kubernetes"
)

// DOKSClusterReconciler reconciles a DOKSCluster object
type DOKSClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinepools;machinepools/status,verbs=get;list

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

	return r.reconcile(ctx, clusterScope)
}

func (r *DOKSClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.DOKSClusterScope) (_ reconcile.Result, reterr error) {
	clusterScope.Info("Deleting DOKSCluster")
	dokscluster := clusterScope.DOKSCluster

	// Handle cluster deletion
	clusterId := clusterScope.GetInstanceID()
	kubernetessvc := kubernetes.NewService(ctx, clusterScope)
	docluster, err := kubernetessvc.GetCluster(clusterId)
	if err != nil {
		return reconcile.Result{}, err
	}

	if docluster != nil {
		if err := kubernetessvc.DeleteCluster(clusterId); err != nil {
			return reconcile.Result{}, err
		}
	} else {
		clusterScope.V(2).Info("Unable to locate cluster")
		r.Recorder.Eventf(dokscluster, corev1.EventTypeWarning, "NoInstanceFound", "Skip deleting")
	}

	r.Recorder.Eventf(dokscluster, corev1.EventTypeNormal, "ClusterDeleted", "Deleted a cluster - %s", clusterScope.Name())
	controllerutil.RemoveFinalizer(dokscluster, infrav1.DOKSClusterFinalizer)

	return reconcile.Result{}, nil
}

func (r *DOKSClusterReconciler) reconcile(ctx context.Context, clusterScope *scope.DOKSClusterScope) (_ reconcile.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	clusterScope.Info("Reconciling DOKSCluster")
	dokscluster := clusterScope.DOKSCluster

	// If the DOKSCluster is in an error state, return early.
	if dokscluster.Status.FailureReason != nil || dokscluster.Status.FailureMessage != nil {
		clusterScope.Info("Error state detected, skipping reconciliation")
		return reconcile.Result{}, nil
	}

	// If the DOKSCluster doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(dokscluster, infrav1.DOKSClusterFinalizer)

	// Handle new clusters
	kubernetessvc := kubernetes.NewService(ctx, clusterScope)
	docluster, err := kubernetessvc.GetCluster(clusterScope.GetInstanceID())
	if err != nil {
		return reconcile.Result{}, err
	}
	if docluster == nil {
		// List NodePools that are related to the cluster
		labels := map[string]string{clusterv1.ClusterLabelName: clusterScope.Name()}
		nodePoolList := &infrav1.DOKSNodePoolList{}
		if err := r.List(ctx, nodePoolList, client.MatchingLabels(labels)); err != nil {
			return reconcile.Result{}, errors.Errorf("failed to list cluster node pools: %+v", err)
		}

		clusterScope.Info(fmt.Sprintf("Found %d MachinePools", len(nodePoolList.Items)))

		var nodePoolScopes []*scope.DOKSNodePoolScope
		for _, nodePool := range nodePoolList.Items {
			machinePool, err := exp_util.GetOwnerMachinePool(ctx, r.Client, nodePool.ObjectMeta)
			if err != nil {
				return reconcile.Result{}, errors.Errorf("failed to locate owning MachinePool: %+v", err)
			}

			nodePoolScope, err := scope.NewDOKSNodePoolScope(scope.DOKSNodePoolScopeParams{
				Client:       r.Client,
				Logger:       log,
				Cluster:      clusterScope.Cluster,
				DOKSCluster:  clusterScope.DOKSCluster,
				MachinePool:  machinePool,
				DOKSNodePool: &nodePool,
			})
			if err != nil {
				return reconcile.Result{}, errors.Errorf("failed to create NodePool Scope: %+v", err)
			}

			nodePoolScopes = append(nodePoolScopes, nodePoolScope)
		}

		// Always close the scopes when exiting this function so we can persist any changes.
		defer func() {
			for _, nodePoolScope := range nodePoolScopes {
				if err := nodePoolScope.Close(); err != nil && reterr == nil {
					reterr = err
				}
			}
		}()

		kubernetessvc := kubernetes.NewService(ctx, clusterScope)
		docluster, err := kubernetessvc.CreateCluster(clusterScope, nodePoolScopes)
		if err != nil {
			err := errors.Errorf("failed to create cluster with DigitalOcean API: %+v", err)
			r.Recorder.Event(dokscluster, corev1.EventTypeWarning, "ClusterCreatingError", err.Error())
			clusterScope.SetProviderStatus(&godo.KubernetesClusterStatus{
				State:   godo.KubernetesClusterStatusError,
				Message: err.Error(),
			})

			return reconcile.Result{}, err
		}
		r.Recorder.Eventf(dokscluster, corev1.EventTypeNormal, "ClusterCreated", "Created new cluster - %s", docluster.Name)

		// Link the DigitalOcean Cluster to our DOKSCluster
		clusterScope.SetProviderID(docluster.ID)

		// Link the DigitalOcean NodePools to our clusters DOKSNodePools
		for _, nodePoolScope := range nodePoolScopes {
			for _, nodePool := range docluster.NodePools {
				if nodePool.Name == nodePoolScope.Name() {
					nodePoolScope.SetProviderID(nodePool.ID)
					break
				}
			}
		}
	}

	clusterScope.SetControlPlaneEndpoint(docluster.Endpoint)
	clusterScope.SetProviderStatus(docluster.Status)

	switch docluster.Status.State {
	case godo.KubernetesClusterStatusRunning:
		clusterScope.Info("Set DOKSCluster status to ready")
		clusterScope.SetReady()
		r.Recorder.Eventf(dokscluster, corev1.EventTypeNormal, "DOKSClusterReady", "DOKSCluster %s - has ready status", clusterScope.Name())
		return reconcile.Result{}, nil
	case godo.KubernetesClusterStatusUpgrading:
		r.Recorder.Eventf(dokscluster, corev1.EventTypeNormal, "DOKSClusterUpgrading", "DOKSCluster %s - is upgrading", clusterScope.Name())
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	case godo.KubernetesClusterStatusProvisioning:
		r.Recorder.Eventf(dokscluster, corev1.EventTypeNormal, "DOKSClusterProvisioning", "DOKSCluster %s - is provisioning", clusterScope.Name())
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	case godo.KubernetesClusterStatusDegraded:
		r.Recorder.Eventf(dokscluster, corev1.EventTypeWarning, "DOKSClusterDegraded", "DOKSCluster %s - has degraded status", clusterScope.Name())
		return reconcile.Result{}, nil
	case godo.KubernetesClusterStatusInvalid:
		r.Recorder.Eventf(dokscluster, corev1.EventTypeWarning, "DOKSClusterInvalid", "DOKSCluster %s - has invalid state", clusterScope.Name())
		return reconcile.Result{}, nil
	case godo.KubernetesClusterStatusError:
		clusterScope.SetFailureMessage(errors.Errorf(docluster.Status.Message))
		r.Recorder.Eventf(dokscluster, corev1.EventTypeWarning, "DOKSClusterError", "DOKSCluster %s - has error status", clusterScope.Name())
		return reconcile.Result{}, nil
	default:
		clusterScope.SetFailureMessage(errors.Errorf("Cluster status %q is unexpected", string(docluster.Status.State)))
		return reconcile.Result{}, nil
	}

}
