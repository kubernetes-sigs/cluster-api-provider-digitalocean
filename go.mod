module sigs.k8s.io/cluster-api-provider-digitalocean

go 1.13

require (
	github.com/digitalocean/godo v1.54.0
	github.com/go-logr/logr v0.3.0
	github.com/miekg/dns v1.1.3
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20201203001011-0b49973bad19
	k8s.io/api v0.17.14
	k8s.io/apimachinery v0.17.14
	k8s.io/client-go v0.17.14
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/cluster-api v0.3.12
	sigs.k8s.io/controller-runtime v0.5.14
)
