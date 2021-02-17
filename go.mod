module sigs.k8s.io/cluster-api-provider-digitalocean

go 1.16

require (
	github.com/digitalocean/godo v1.54.0
	github.com/go-logr/logr v0.4.0
	github.com/golang/mock v1.4.4
	github.com/miekg/dns v1.1.3
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20201203001011-0b49973bad19
	k8s.io/api v0.21.0-beta.1
	k8s.io/apimachinery v0.21.0-beta.1
	k8s.io/client-go v0.21.0-beta.1
	k8s.io/klog/v2 v2.8.0
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10
	sigs.k8s.io/cluster-api v0.0.0-20210325161731-7124a659accd
	sigs.k8s.io/controller-runtime v0.9.0-alpha.1
)
