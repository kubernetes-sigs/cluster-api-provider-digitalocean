module sigs.k8s.io/cluster-api-provider-digitalocean

go 1.16

require (
	github.com/digitalocean/godo v1.65.0
	github.com/go-logr/logr v0.4.0
	github.com/miekg/dns v1.1.26
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/klog/v2 v2.9.0
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	sigs.k8s.io/cluster-api v1.0.0
	sigs.k8s.io/cluster-api/test v1.0.0
	sigs.k8s.io/controller-runtime v0.10.2
)

replace sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.0
