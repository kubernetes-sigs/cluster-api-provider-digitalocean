module sigs.k8s.io/cluster-api-provider-digitalocean

go 1.13

require (
	github.com/digitalocean/godo v1.47.0
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	k8s.io/api v0.17.12
	k8s.io/apimachinery v0.17.12
	k8s.io/client-go v0.17.12
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19
	sigs.k8s.io/cluster-api v0.3.11
	sigs.k8s.io/controller-runtime v0.5.11
)
