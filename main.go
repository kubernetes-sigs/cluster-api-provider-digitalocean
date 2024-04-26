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

// Package main
package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	// +kubebuilder:scaffold:imports

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	cgrecord "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/flags"
	"sigs.k8s.io/cluster-api/util/record"

	infrav1alpha4 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"
	infrav1beta1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-digitalocean/controllers"
	dnsutil "sigs.k8s.io/cluster-api-provider-digitalocean/util/dns"
	dnsresolver "sigs.k8s.io/cluster-api-provider-digitalocean/util/dns/resolver"
	"sigs.k8s.io/cluster-api-provider-digitalocean/util/reconciler"
	"sigs.k8s.io/cluster-api-provider-digitalocean/version"
)

var (
	scheme             = runtime.NewScheme()
	setupLog           = ctrl.Log.WithName("setup")
	diagnosticsOptions = flags.DiagnosticsOptions{}
)

func init() {
	klog.InitFlags(nil)

	_ = clientgoscheme.AddToScheme(scheme)
	_ = infrav1alpha4.AddToScheme(scheme)
	_ = infrav1beta1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

var (
	leaderElectionLeaseDuration time.Duration
	leaderElectionRenewDeadline time.Duration
	leaderElectionRetryPeriod   time.Duration
	syncPeriod                  time.Duration
	reconcileTimeout            time.Duration
	leaderElectionNamespace     string
	healthAddr                  string
	watchNamespace              string
	profilerAddress             string
	webhookCertDir              string
	webhookPort                 int
	enableLeaderElection        bool
	restConfigQPS               float32
	restConfigBurst             int
)

func initFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&leaderElectionNamespace, "leader-election-namespace", "", "Namespace that the controller performs leader election in. If unspecified, the controller will discover which namespace it is running in.")
	fs.DurationVar(&leaderElectionLeaseDuration, "leader-elect-lease-duration", 15*time.Second, "Interval at which non-leader candidates will wait to force acquire leadership (duration string)")
	fs.DurationVar(&leaderElectionRenewDeadline, "leader-elect-renew-deadline", 10*time.Second, "Duration that the leading controller manager will retry refreshing leadership before giving up (duration string)")
	fs.DurationVar(&leaderElectionRetryPeriod, "leader-elect-retry-period", 2*time.Second, "Duration the LeaderElector clients should wait between tries of actions (duration string)")
	fs.StringVar(&healthAddr, "health-addr", ":9440", "The address the health endpoint binds to.")
	fs.StringVar(&watchNamespace, "namespace", "", "Namespace that the controller watches to reconcile cluster-api objects. If unspecified, the controller watches for cluster-api objects across all namespaces.")
	fs.StringVar(&profilerAddress, "profiler-address", "", "Bind address to expose the pprof profiler (e.g. localhost:6060)")
	fs.DurationVar(&syncPeriod, "sync-period", 10*time.Minute, "The minimum interval at which watched resources are reconciled (e.g. 10m)")
	fs.IntVar(&webhookPort, "webhook-port", 9443, "Webhook Server port, disabled by default. When enabled, the manager will only work as webhook server, no reconcilers are installed.")
	fs.DurationVar(&reconcileTimeout, "reconcile-timeout", reconciler.DefaultLoopTimeout, "The maximum duration a reconcile loop can run (e.g. 90m)")
	fs.Float32Var(&restConfigQPS, "kube-api-qps", 20, "Maximum queries per second from the controller client to the Kubernetes API server.")
	fs.IntVar(&restConfigBurst, "kube-api-burst", 30, "Maximum number of queries that should be allowed in one burst from the controller client to the Kubernetes API server.")
	fs.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs", "Webhook Server Certificate Directory, is the directory that contains the server key and certificate")
	flags.AddDiagnosticsOptions(fs, &diagnosticsOptions)
}

// Add RBAC for the authorized diagnostics endpoint.
// +kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create

func main() {
	initFlags(pflag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	diagnosticsOpts := flags.GetDiagnosticsOptions(diagnosticsOptions)

	var watchNamespaces map[string]cache.Config
	if watchNamespace != "" {
		watchNamespaces = map[string]cache.Config{
			watchNamespace: {},
		}
		setupLog.Info("Watching cluster-api objects only in namespace for reconciliation", "namespace", watchNamespace)
	}

	if profilerAddress != "" {
		setupLog.Info("Profiler listening for requests", "profiler-address", profilerAddress)
		go func() {
			server := &http.Server{
				Addr: profilerAddress,

				// Timeouts
				ReadTimeout:       60 * time.Second,
				ReadHeaderTimeout: 60 * time.Second,
				WriteTimeout:      60 * time.Second,
				IdleTimeout:       60 * time.Second,
			}
			err := server.ListenAndServe()
			if err != nil {
				setupLog.Error(err, "listen and serve error")
			}
		}()
	}

	ctrl.SetLogger(klog.Background())
	ctx := ctrl.SetupSignalHandler()

	// Machine and cluster operations can create enough events to trigger the event recorder spam filter
	// Setting the burst size higher ensures all events will be recorded and submitted to the API
	broadcaster := cgrecord.NewBroadcasterWithCorrelatorOptions(cgrecord.CorrelatorOptions{
		BurstSize: 100,
	})

	restConfig := ctrl.GetConfigOrDie()
	restConfig.QPS = restConfigQPS
	restConfig.Burst = restConfigBurst

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                     scheme,
		Metrics:                    diagnosticsOpts,
		LeaderElection:             enableLeaderElection,
		LeaderElectionID:           "controller-leader-election-capdo",
		LeaderElectionNamespace:    leaderElectionNamespace,
		LeaseDuration:              &leaderElectionLeaseDuration,
		RenewDeadline:              &leaderElectionRenewDeadline,
		RetryPeriod:                &leaderElectionRetryPeriod,
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
		Cache: cache.Options{
			DefaultNamespaces: watchNamespaces,
			SyncPeriod:        &syncPeriod,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:    webhookPort,
			CertDir: webhookCertDir,
		}),
		HealthProbeBindAddress: healthAddr,
		EventBroadcaster:       broadcaster,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize event recorder.
	record.InitFromRecorder(mgr.GetEventRecorderFor("digitalocean-controller"))

	dnsresolver, err := dnsresolver.NewDNSResolver()
	if err != nil {
		setupLog.Error(err, "unable to create dns resolver")
		os.Exit(1)
	}

	dnsutil.InitFromDNSResolver(dnsresolver)

	if err = (&controllers.DOClusterReconciler{
		Client:           mgr.GetClient(),
		Recorder:         mgr.GetEventRecorderFor("docluster-controller"),
		ReconcileTimeout: reconcileTimeout,
	}).SetupWithManager(ctx, mgr, controller.Options{}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DOCluster")
		os.Exit(1)
	}
	if err = (&controllers.DOMachineReconciler{
		Client:           mgr.GetClient(),
		Recorder:         mgr.GetEventRecorderFor("domachine-controller"),
		ReconcileTimeout: reconcileTimeout,
	}).SetupWithManager(ctx, mgr, controller.Options{}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DOMachine")
		os.Exit(1)
	}

	if err := (&infrav1beta1.DOCluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "DOCluster")
		os.Exit(1)
	}

	if err = (&infrav1beta1.DOClusterTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "DOClusterTemplate")
		os.Exit(1)
	}

	if err := (&infrav1beta1.DOMachine{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "DOMachine")
		os.Exit(1)
	}

	if err := (&infrav1beta1.DOMachineTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "DOMachineTemplate")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddReadyzCheck("webhook", mgr.GetWebhookServer().StartedChecker()); err != nil {
		setupLog.Error(err, "unable to create ready check")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("webhook", mgr.GetWebhookServer().StartedChecker()); err != nil {
		setupLog.Error(err, "unable to create health check")
		os.Exit(1)
	}

	setupLog.Info("starting manager", "version", version.Get().String(), "extended_info", version.Get())
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
