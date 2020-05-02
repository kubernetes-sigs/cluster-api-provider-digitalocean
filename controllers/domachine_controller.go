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
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha2"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/computes"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DOMachineReconciler reconciles a DOMachine object.
type DOMachineReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *DOMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.DOMachine{}).
		Watches(
			&source.Kind{Type: &clusterv1.Machine{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: util.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("DOMachine")),
			},
		).
		Watches(
			&source.Kind{Type: &infrav1.DOCluster{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(r.DOClusterToDOMachines)},
		).
		Complete(r)
}

func (r *DOMachineReconciler) DOClusterToDOMachines(o handler.MapObject) []ctrl.Request {
	result := []ctrl.Request{}

	c, ok := o.Object.(*infrav1.DOCluster)
	if !ok {
		r.Log.Error(errors.Errorf("expected a DOCluster but got a %T", o.Object), "failed to get DOMachine for DOCluster")
		return nil
	}
	log := r.Log.WithValues("DOCluster", c.Name, "Namespace", c.Namespace)

	cluster, err := util.GetOwnerCluster(context.TODO(), r.Client, c.ObjectMeta)
	switch {
	case apierrors.IsNotFound(err) || cluster == nil:
		return result
	case err != nil:
		log.Error(err, "failed to get owning cluster")
		return result
	}

	labels := map[string]string{clusterv1.MachineClusterLabelName: cluster.Name}
	machineList := &clusterv1.MachineList{}
	if err := r.List(context.TODO(), machineList, client.InNamespace(c.Namespace), client.MatchingLabels(labels)); err != nil {
		log.Error(err, "failed to list Machines")
		return nil
	}
	for _, m := range machineList.Items {
		if m.Spec.InfrastructureRef.Name == "" {
			continue
		}
		name := client.ObjectKey{Namespace: m.Namespace, Name: m.Spec.InfrastructureRef.Name}
		result = append(result, ctrl.Request{NamespacedName: name})
	}

	return result
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=domachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=domachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch

func (r *DOMachineReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	logger := r.Log.WithValues("doCluster", req.NamespacedName.Name, "namespace", req.NamespacedName.Namespace)

	domachine := &infrav1.DOMachine{}
	if err := r.Get(ctx, req.NamespacedName, domachine); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	logger = logger.WithName(domachine.APIVersion)

	// Fetch the Machine.
	machine, err := util.GetOwnerMachine(ctx, r.Client, domachine.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if machine == nil {
		logger.Info("Machine Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}

	logger = logger.WithValues("machine", machine.Name)

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		logger.Info("Machine is missing cluster label or cluster does not exist")
		return reconcile.Result{}, nil
	}

	logger = logger.WithValues("cluster", cluster.Name)

	docluster := &infrav1.DOCluster{}
	doclusterNamespacedName := client.ObjectKey{
		Namespace: domachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Get(ctx, doclusterNamespacedName, docluster); err != nil {
		logger.Info("DOluster is not available yet")
		return reconcile.Result{}, nil
	}

	logger = logger.WithValues("docluster", docluster.Name)

	// Create the cluster scope
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:    r.Client,
		Logger:    logger,
		Cluster:   cluster,
		DOCluster: docluster,
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	// Create the machine scope
	machineScope, err := scope.NewMachineScope(scope.MachineScopeParams{
		Logger:    logger,
		Client:    r.Client,
		Cluster:   cluster,
		Machine:   machine,
		DOCluster: docluster,
		DOMachine: domachine,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any DOMachine changes.
	defer func() {
		if err := machineScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted machines
	if !domachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machineScope, clusterScope)
	}

	return r.reconcile(ctx, machineScope, clusterScope)
}

func (r *DOMachineReconciler) reconcile(ctx context.Context, machineScope *scope.MachineScope, clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	machineScope.Info("Reconciling DOMachine")
	domachine := machineScope.DOMachine
	// If the DOMachine is in an error state, return early.
	if domachine.Status.ErrorReason != nil || domachine.Status.ErrorMessage != nil {
		machineScope.Info("Error state detected, skipping reconciliation")
		return reconcile.Result{}, nil
	}

	// If the DOMachine doesn't have our finalizer, add it.
	if !util.Contains(domachine.Finalizers, infrav1.MachineFinalizer) {
		domachine.Finalizers = append(domachine.Finalizers, infrav1.MachineFinalizer)
	}

	if !machineScope.Cluster.Status.InfrastructureReady {
		machineScope.Info("Cluster infrastructure is not ready yet")
		return reconcile.Result{}, nil
	}

	// Make sure bootstrap data is available and populated.
	if machineScope.Machine.Spec.Bootstrap.Data == nil {
		machineScope.Info("Bootstrap data is not yet available")
		return reconcile.Result{}, nil
	}

	computesvc := computes.NewService(ctx, clusterScope)
	droplet, err := computesvc.GetDroplet(machineScope.GetInstanceID())
	if err != nil {
		return reconcile.Result{}, err
	}
	if droplet == nil {
		droplet, err = computesvc.CreateDroplet(machineScope)
		if err != nil {
			errs := errors.Errorf("Failed to create droplet instance for DOMachine %s/%s: %v", domachine.Namespace, domachine.Name, err)
			machineScope.SetErrorReason(capierrors.CreateMachineError)
			machineScope.SetErrorMessage(errs)
			r.Recorder.Event(domachine, corev1.EventTypeWarning, "InstanceCreatingError", errs.Error())
			return reconcile.Result{}, errs
		}
		r.Recorder.Eventf(domachine, corev1.EventTypeNormal, "InstanceCreated", "Created new droplet instance - %s", droplet.Name)
	}

	machineScope.SetProviderID(strconv.Itoa(droplet.ID))
	machineScope.SetInstanceStatus(infrav1.DOResourceStatus(droplet.Status))

	addrs, err := computesvc.GetDropletAddress(droplet)
	if err != nil {
		machineScope.SetErrorMessage(errors.New("failed to getting droplet address"))
		return reconcile.Result{}, err
	}
	machineScope.SetAddresses(addrs)

	// Proceed to reconcile the DOMachine state.
	switch infrav1.DOResourceStatus(droplet.Status) {
	case infrav1.DOResourceStatusNew:
		machineScope.Info("Machine instance is pending", "instance-id", machineScope.GetInstanceID())
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	case infrav1.DOResourceStatusRunning:
		machineScope.Info("Machine instance is active", "instance-id", machineScope.GetInstanceID())
		machineScope.SetReady()
		r.Recorder.Eventf(domachine, corev1.EventTypeNormal, "DOMachineReady", "DOMachine %s - has ready status", droplet.Name)
		return reconcile.Result{}, nil
	default:
		machineScope.SetErrorReason(capierrors.UpdateMachineError)
		machineScope.SetErrorMessage(errors.Errorf("Instance status %q is unexpected", droplet.Status))
		return reconcile.Result{}, nil
	}
}

func (r *DOMachineReconciler) reconcileDelete(ctx context.Context, machineScope *scope.MachineScope, clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	machineScope.Info("Reconciling delete DOMachine")
	domachine := machineScope.DOMachine

	computesvc := computes.NewService(ctx, clusterScope)
	droplet, err := computesvc.GetDroplet(machineScope.GetInstanceID())
	if err != nil {
		return reconcile.Result{}, err
	}

	if droplet == nil {
		clusterScope.V(2).Info("Unable to locate droplet instance")
		r.Recorder.Eventf(domachine, corev1.EventTypeWarning, "NoInstanceFound", "Skip deleting")
		return reconcile.Result{}, nil
	}

	if err := computesvc.DeleteDroplet(machineScope.GetInstanceID()); err != nil {
		return reconcile.Result{}, err
	}

	r.Recorder.Eventf(domachine, corev1.EventTypeNormal, "InstanceDeleted", "Deleted a instance - %s", droplet.Name)
	domachine.Finalizers = util.Filter(domachine.Finalizers, infrav1.MachineFinalizer)
	return reconcile.Result{}, nil
}
