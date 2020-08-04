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

package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof" //nolint
	"os"

	// +kubebuilder:scaffold:imports
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	infrav1alpha2 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha2"
	infrav1alpha3 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha3"
	"sigs.k8s.io/cluster-api-provider-digitalocean/controllers"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	klog.InitFlags(nil)

	_ = clientgoscheme.AddToScheme(scheme)
	_ = infrav1alpha2.AddToScheme(scheme)
	_ = infrav1alpha3.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

var (
	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	healthAddr              string
	watchNamespace          string
	profilerAddress         string
)

func InitFlags(fs *pflag.FlagSet) {
	fs.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	fs.BoolVar(&enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	fs.StringVar(&leaderElectionNamespace, "leader-election-namespace", "", "Namespace that the controller performs leader election in. If unspecified, the controller will discover which namespace it is running in.")
	fs.StringVar(&healthAddr, "health-addr", ":9440", "The address the health endpoint binds to.")
	fs.StringVar(&watchNamespace, "namespace", "", "Namespace that the controller watches to reconcile cluster-api objects. If unspecified, the controller watches for cluster-api objects across all namespaces.")
	fs.StringVar(&profilerAddress, "profiler-address", "", "Bind address to expose the pprof profiler (e.g. localhost:6060)")
}

func main() {
	InitFlags(pflag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if watchNamespace != "" {
		setupLog.Info("Watching cluster-api objects only in namespace for reconciliation", "namespace", watchNamespace)
	}

	if profilerAddress != "" {
		setupLog.Info("Profiler listening for requests", "profiler-address", profilerAddress)
		go func() {
			setupLog.Error(http.ListenAndServe(profilerAddress, nil), "listen and serve error")
		}()
	}

	ctrl.SetLogger(klogr.New())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      metricsAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "controller-leader-election-capdo",
		LeaderElectionNamespace: leaderElectionNamespace,
		Namespace:               watchNamespace,
		Port:                    9443,
		HealthProbeBindAddress:  healthAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize event recorder.
	record.InitFromRecorder(mgr.GetEventRecorderFor("digitalocean-controller"))

	if err = (&controllers.DOClusterReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("DOCluster"),
		Recorder: mgr.GetEventRecorderFor("docluster-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DOCluster")
		os.Exit(1)
	}
	if err = (&controllers.DOMachineReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("DOMachine"),
		Recorder: mgr.GetEventRecorderFor("domachine-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DOMachine")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create ready check")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create health check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
