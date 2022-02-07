module sigs.k8s.io/cluster-api-provider-digitalocean

go 1.16

require (
	github.com/digitalocean/godo v1.73.0
	github.com/go-logr/logr v0.4.0
	github.com/miekg/dns v1.1.3
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	k8s.io/api v0.21.4
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v0.21.4
	k8s.io/klog/v2 v2.9.0
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
	sigs.k8s.io/cluster-api v0.4.7
	sigs.k8s.io/cluster-api/test v0.4.7
	sigs.k8s.io/controller-runtime v0.9.7
)

replace sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.4.7
