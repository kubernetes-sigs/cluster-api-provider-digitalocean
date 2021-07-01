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

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	exp_clusterv1 "sigs.k8s.io/cluster-api/exp/api/v1alpha4"
	exp_util "sigs.k8s.io/cluster-api/exp/util"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/kubernetes"
)

// DOKSNodePoolReconciler reconciles a DOKSNodePool object
type DOKSNodePoolReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksnodepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksnodepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doksnodepools/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=doksclusters;doksclusters/status,verbs=get
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinepools;machinepools/status,verbs=get
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get

// SetupWithManager sets up the controller with the Manager.
func (r *DOKSNodePoolReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	log := ctrl.LoggerFrom(ctx)

	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.DOKSNodePool{}).
		WithEventFilter(predicates.ResourceNotPaused(ctrl.LoggerFrom(ctx))). // don't queue reconcile if resource is paused
		Watches(
			&source.Kind{Type: &exp_clusterv1.MachinePool{}},
			handler.EnqueueRequestsFromMapFunc(exp_util.MachinePoolToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("DOKSNodePool"), log)),
		).
		Build(r)
	if err != nil {
		return errors.Wrapf(err, "error creating controller")
	}

	clusterToObjectFunc, err := util.ClusterToObjectsMapper(r.Client, &infrav1.DOKSNodePoolList{}, mgr.GetScheme())
	if err != nil {
		return errors.Wrapf(err, "failed to create mapper for Cluster to DOKSNodePool")
	}

	// Add a watch on clusterv1.Cluster object for unpause & ready notifications.
	if err := c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(clusterToObjectFunc),
		predicates.ClusterUnpausedAndInfrastructureReady(ctrl.LoggerFrom(ctx)),
	); err != nil {
		return errors.Wrapf(err, "failed adding a watch for ready clusters")
	}

	return nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DOKSNodePool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DOKSNodePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := log.FromContext(ctx)

	doksnodepool := &infrav1.DOKSNodePool{}
	if err := r.Get(ctx, req.NamespacedName, doksnodepool); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	machinePool, err := exp_util.GetOwnerMachinePool(ctx, r.Client, doksnodepool.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, nil
	}
	if machinePool == nil {
		log.Info("MachinePool Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}

	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, doksnodepool.ObjectMeta)
	if err != nil {
		log.Info("DOKSNodePool is missing cluster label or cluster does not exist")
		return reconcile.Result{}, nil
	}

	dokscluster := &infrav1.DOKSCluster{}
	doksclusterNamespacedName := client.ObjectKey{
		Namespace: cluster.Spec.InfrastructureRef.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Get(ctx, doksclusterNamespacedName, dokscluster); err != nil {
		log.Info("DOKSCluster is not available yet")
		return reconcile.Result{}, err
	}

	nodePoolScope, err := scope.NewDOKSNodePoolScope(scope.DOKSNodePoolScopeParams{
		Client:       r.Client,
		Logger:       log,
		Cluster:      cluster,
		DOKSCluster:  dokscluster,
		MachinePool:  machinePool,
		DOKSNodePool: doksnodepool,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	defer func() {
		if err := nodePoolScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted node pools
	if !doksnodepool.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, nodePoolScope)
	}

	return r.reconcile(ctx, nodePoolScope)
}

func (r *DOKSNodePoolReconciler) reconcile(ctx context.Context, nodePoolScope *scope.DOKSNodePoolScope) (reconcile.Result, error) {
	doksnodepool := nodePoolScope.DOKSNodePool
	machinepool := nodePoolScope.MachinePool

	// If the DOKSNodePool is in an error state, return early.
	if doksnodepool.Status.FailureReason != nil || doksnodepool.Status.FailureMessage != nil {
		nodePoolScope.Info("Error state detected, skipping reconciliation")
		return reconcile.Result{}, nil
	}

	// If the DOKSNodePool doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(doksnodepool, infrav1.DOKSNodePoolFinalizer)

	if !nodePoolScope.Cluster.Status.InfrastructureReady {
		nodePoolScope.Info("Cluster infrastructure is not ready yet")
		return reconcile.Result{}, nil
	}

	clusterScope, err := nodePoolScope.ToDOKSClusterScope()
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create DOKSClusterScope from DOKSNodePoolScope: %+v", err)
	}

	// Handle new node pools
	doClusterId := clusterScope.GetInstanceID()
	kubernetessvc := kubernetes.NewService(ctx, clusterScope)
	nodePool, err := kubernetessvc.GetNodePool(doClusterId, nodePoolScope.GetInstanceID())
	if err != nil {
		return reconcile.Result{}, err
	}
	if nodePool == nil {
		nodePool, err = kubernetessvc.CreateNodePool(doClusterId, nodePoolScope)
		if err != nil {
			err = errors.Errorf("Failed to create node pool %s/%s: %v", doksnodepool.Namespace, doksnodepool.Name, err)
			r.Recorder.Event(doksnodepool, corev1.EventTypeWarning, "InstanceCreatingError", err.Error())
			return reconcile.Result{}, err
		}
		r.Recorder.Eventf(doksnodepool, corev1.EventTypeNormal, "InstanceCreated", "Created new node pool - %s", nodePool.Name)
	}

	nodePoolScope.SetProviderID(nodePool.ID)

	// Handle existing node pools
	updateRequest := godo.KubernetesNodePoolUpdateRequest{
		Name:      nodePoolScope.Name(),
		AutoScale: &doksnodepool.Spec.AutoScale,
		MinNodes:  doksnodepool.Spec.MinNodes,
		MaxNodes:  doksnodepool.Spec.MaxNodes,
	}

	// When DOKSNodePool is set to autoscale sync what is currently set in DigitalOcean
	if doksnodepool.Spec.AutoScale {
		count := int32(nodePool.Count)
		machinepool.Spec.Replicas = &count
	} else { // When not autoscaling update to our desired replica count
		updateRequest.Count = nodePoolScope.Replicas()
	}

	if _, _, err := nodePoolScope.Kubernetes.UpdateNodePool(ctx, clusterScope.GetInstanceID(), nodePoolScope.GetInstanceID(), &updateRequest); err != nil {
		return reconcile.Result{}, errors.Errorf("failed to update node pool: %+v", err)
	}

	return reconcile.Result{}, nil
}

func (r *DOKSNodePoolReconciler) reconcileDelete(ctx context.Context, nodePoolScope *scope.DOKSNodePoolScope) (reconcile.Result, error) {
	nodePoolScope.Info("Deleting DOKSNodePool")
	doksnodepool := nodePoolScope.DOKSNodePool

	clusterScope, err := nodePoolScope.ToDOKSClusterScope()
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create DOKSClusterScope from DOKSNodePoolScope: %+v", err)
	}

	if !nodePoolScope.Cluster.DeletionTimestamp.IsZero() {
		nodePoolScope.Info("The Cluster of this NodePool is beeing deleted. Skipping NodePool deletion.")
		controllerutil.RemoveFinalizer(doksnodepool, infrav1.DOKSNodePoolFinalizer)
		return reconcile.Result{}, nil
	}

	clusterId := clusterScope.GetInstanceID()
	nodePoolId := nodePoolScope.GetInstanceID()
	kubernetessvc := kubernetes.NewService(ctx, clusterScope)
	nodePool, err := kubernetessvc.GetNodePool(clusterId, nodePoolId)
	if err != nil {
		return reconcile.Result{}, err
	}

	if nodePool != nil {
		if err := kubernetessvc.DeleteNodePool(clusterId, nodePoolId); err != nil {
			return reconcile.Result{}, err
		}
	} else {
		nodePoolScope.V(2).Info("Unable to locate node pool")
		r.Recorder.Eventf(doksnodepool, corev1.EventTypeWarning, "NoInstanceFound", "Skip deleting")
	}

	r.Recorder.Eventf(doksnodepool, corev1.EventTypeNormal, "InstanceDeleted", "Deleted a node pool - %s", nodePoolScope.Name())
	controllerutil.RemoveFinalizer(doksnodepool, infrav1.DOKSNodePoolFinalizer)

	return reconcile.Result{}, nil
}
