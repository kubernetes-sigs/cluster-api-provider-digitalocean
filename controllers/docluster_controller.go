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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha3"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/networking"
	dnsutil "sigs.k8s.io/cluster-api-provider-digitalocean/util/dns"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DOClusterReconciler reconciles a DOCluster object.
type DOClusterReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *DOClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := r.Log.WithValues("controller", "DOCluster")
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.DOCluster{}).
		WithEventFilter(predicates.ResourceNotPaused(log)). // don't queue reconcile if resource is paused
		Build(r)
	if err != nil {
		return errors.Wrapf(err, "error creating controller")
	}

	// Add a watch on clusterv1.Cluster object for unpause notifications.
	if err = c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: util.ClusterToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("DOCluster")),
		},
		predicates.ClusterUnpaused(log),
	); err != nil {
		return errors.Wrapf(err, "failed adding a watch for ready clusters")
	}

	return nil
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=doclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch

func (r *DOClusterReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	log := r.Log.WithValues("doCluster", req.NamespacedName.Name, "namespace", req.NamespacedName.Namespace)

	docluster := &infrav1.DOCluster{}
	if err := r.Get(ctx, req.NamespacedName, docluster); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	log = log.WithName(docluster.APIVersion)

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, docluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	// Create the cluster scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:    r.Client,
		Logger:    log,
		Cluster:   cluster,
		DOCluster: docluster,
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
	if !docluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, clusterScope)
	}

	return r.reconcile(ctx, clusterScope)
}

func (r *DOClusterReconciler) reconcile(ctx context.Context, clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling DOCluster")
	docluster := clusterScope.DOCluster
	// If the DOCluster doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(docluster, infrav1.ClusterFinalizer)

	networkingsvc := networking.NewService(ctx, clusterScope)
	apiServerLoadbalancer := clusterScope.APIServerLoadbalancers()
	apiServerLoadbalancer.ApplyDefault()

	apiServerLoadbalancerRef := clusterScope.APIServerLoadbalancersRef()
	loadbalancer, err := networkingsvc.GetLoadBalancer(apiServerLoadbalancerRef.ResourceID)
	if err != nil {
		return reconcile.Result{}, err
	}
	if loadbalancer == nil {
		loadbalancer, err = networkingsvc.CreateLoadBalancer(apiServerLoadbalancer)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "failed to create load balancers for DOCluster %s/%s", docluster.Namespace, docluster.Name)
		}

		r.Recorder.Eventf(docluster, corev1.EventTypeNormal, "LoadBalancerCreated", "Created new load balancers - %s", loadbalancer.Name)
	}

	apiServerLoadbalancerRef.ResourceID = loadbalancer.ID
	apiServerLoadbalancerRef.ResourceStatus = infrav1.DOResourceStatus(loadbalancer.Status)

	if apiServerLoadbalancerRef.ResourceStatus != infrav1.DOResourceStatusRunning && loadbalancer.IP == "" {
		clusterScope.Info("Waiting on API server Global IP Address")
		return reconcile.Result{RequeueAfter: 15 * time.Second}, nil
	}

	r.Recorder.Eventf(docluster, corev1.EventTypeNormal, "LoadBalancerReady", "LoadBalancer got an IP Address - %s", loadbalancer.IP)

	var controlPlaneEndpoint = loadbalancer.IP
	if docluster.Spec.ControlPlaneDNS != nil {
		clusterScope.Info("Verifying LB DNS Record")
		// ensure DNS record is created and use it as control plane endpoint
		recordSpec := docluster.Spec.ControlPlaneDNS
		controlPlaneEndpoint = fmt.Sprintf("%s.%s", recordSpec.Name, recordSpec.Domain)
		dRecord, err := networkingsvc.GetDomainRecord(
			recordSpec.Domain,
			recordSpec.Name,
			"A",
		)

		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "failed verify DNS record for LB Name %s.%s",
				recordSpec.Name, recordSpec.Domain)
		}

		if dRecord == nil || dRecord.Data != loadbalancer.IP {
			clusterScope.Info("Ensuring LB DNS Record is in place")
			clusterScope.SetControlPlaneDNSRecordReady(false)
			if err := networkingsvc.UpsertDomainRecord(
				recordSpec.Domain,
				recordSpec.Name,
				"A",
				loadbalancer.IP,
			); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to reconcile LB DNS record")
			}
		}

		// If the record has never been ready we need to check whether it has
		// been propagated or not. Updating the record in the DNS API does not
		// mean it is already advertised at the DNS server. If the DNS is slower
		// than our reconciliation is, we'd fall into a case where our
		// reconciler hits an NXDOMAIN which is then stored in the negative
		// cache, so all our retries would fail until the cache TTL is up. This
		// propagation check works around the DNS cache problem by directly
		// making DNS queries and not going through system resolvers.
		if !clusterScope.DOCluster.Status.ControlPlaneDNSRecordReady {
			propagated, err := dnsutil.CheckDNSPropagated(dnsutil.ToFQDN(recordSpec.Name, recordSpec.Domain), loadbalancer.IP)
			if err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to check DNS propagation")
			}

			if !propagated {
				clusterScope.Info("Waiting for DNS record to be propagated")
				return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
			}

			clusterScope.Info("DNS record is propagated - set DOCluster ControlPlaneDNSRecordReady status to ready")
			clusterScope.SetControlPlaneDNSRecordReady(true)
		}

		clusterScope.Info("LB DNS Record is already ready")
		r.Recorder.Eventf(docluster, corev1.EventTypeNormal, "DomainRecordReady", "DNS Record '%s.%s' with IP '%s'", recordSpec.Name, recordSpec.Domain, loadbalancer.IP)
	}

	clusterScope.SetControlPlaneEndpoint(clusterv1.APIEndpoint{
		Host: controlPlaneEndpoint,
		Port: int32(apiServerLoadbalancer.Port),
	})

	clusterScope.Info("Set DOCluster status to ready")
	clusterScope.SetReady()
	r.Recorder.Eventf(docluster, corev1.EventTypeNormal, "DOClusterReady", "DOCluster %s - has ready status", clusterScope.Name())
	return reconcile.Result{}, nil
}

func (r *DOClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling delete DOCluster")
	docluster := clusterScope.DOCluster
	networkingsvc := networking.NewService(ctx, clusterScope)
	apiServerLoadbalancerRef := clusterScope.APIServerLoadbalancersRef()

	if docluster.Spec.ControlPlaneDNS != nil {
		recordSpec := docluster.Spec.ControlPlaneDNS
		if err := networkingsvc.DeleteDomainRecord(recordSpec.Domain, recordSpec.Name, "A"); err != nil {
			return reconcile.Result{}, err
		}
	}

	loadbalancer, err := networkingsvc.GetLoadBalancer(apiServerLoadbalancerRef.ResourceID)
	if err != nil {
		return reconcile.Result{}, err
	}

	if loadbalancer == nil {
		clusterScope.V(2).Info("Unable to locate load balancer")
		r.Recorder.Eventf(docluster, corev1.EventTypeWarning, "NoLoadBalancerFound", "Unable to find matching load balancer")
		controllerutil.RemoveFinalizer(docluster, infrav1.ClusterFinalizer)
		return reconcile.Result{}, nil
	}

	if err := networkingsvc.DeleteLoadBalancer(loadbalancer.ID); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "error deleting load balancer for DOCluster %s/%s", docluster.Namespace, docluster.Name)
	}

	r.Recorder.Eventf(docluster, corev1.EventTypeNormal, "LoadBalancerDeleted", "Deleted an LoadBalancer - %s", loadbalancer.Name)
	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(docluster, infrav1.ClusterFinalizer)
	return reconcile.Result{}, nil
}
