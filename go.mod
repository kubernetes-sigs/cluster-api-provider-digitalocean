module sigs.k8s.io/cluster-api-provider-digitalocean

go 1.16

require (
	github.com/digitalocean/godo v1.54.0
	github.com/go-logr/logr v0.4.0
	github.com/miekg/dns v1.1.3
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20210615190721-d04028783cf1
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/klog/v2 v2.9.0
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	sigs.k8s.io/cluster-api v0.4.0
	sigs.k8s.io/cluster-api/test v0.4.0
	sigs.k8s.io/controller-runtime v0.9.1
)

replace sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.4.0
