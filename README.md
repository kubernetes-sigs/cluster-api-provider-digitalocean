# Kubernetes Cluster API Provider DigitalOcean

<p align="center"><img alt="capi" src="https://github.com/kubernetes-sigs/cluster-api/raw/master/docs/book/src/images/introduction.png" width="160x" /><img alt="capi" src="https://upload.wikimedia.org/wikipedia/commons/f/ff/DigitalOcean_logo.svg" width="192x" /></p>
<p align="center">
<!-- prow build badge, godoc, and go report card-->
</a> <a href="https://godoc.org/sigs.k8s.io/cluster-api-provider-digitalocean"><img src="https://godoc.org/sigs.k8s.io/cluster-api-provider-digitalocean?status.svg"></a> <a href="https://goreportcard.com/report/sigs.k8s.io/cluster-api-provider-digitalocean"><img alt="Go Report Card" src="https://goreportcard.com/badge/sigs.k8s.io/cluster-api-provider-digitalocean" /></a></p>

------

Kubernetes-native declarative infrastructure for DigitalOcean.

## What is the Cluster API Provider DigitalOcean

The [Cluster API][cluster_api] brings
declarative, Kubernetes-style APIs to cluster creation, configuration and
management.

The API itself is shared across multiple cloud providers allowing for true DigitalOcean
hybrid deployments of Kubernetes. It is built atop the lessons learned from
previous cluster managers such as [kops][kops] and
[kubicorn][kubicorn].

## Project Status

This project is currently work-in-progress and in Alpha, so it may not be production ready. There is no backwards-compatibility guarantee at this point. For more details on the roadmap and upcoming features, check out [the project's issue tracker on GitHub][issue].

## Launching a Kubernetes cluster on DigitalOcean

Check out the [getting started guide](./docs/getting-started.md) for launching a cluster on DigitalOcean. 

## Features

- Native Kubernetes manifests and API
- Support for single and multi-node control plane clusters
- Choice your Linux distribution (as long as a current cloud-init is available)

------

## Compatibility with Cluster API and Kubernetes Versions

TODO

## Documentation

Documentation is in the `/docs` directory.

## Getting involved and contributing

More about development and contributing practices can be found in [`CONTRIBUTING.md`](./CONTRIBUTING.md).

<!-- References -->

[prow]: https://go.k8s.io/bot-commands
[issue]: https://github.com/kubernetes-sigs/cluster-api-provider-digitalocean/issues
[new_issue]: https://github.com/kubernetes-sigs/cluster-api-provider-digitalocean/issues/new
[good_first_issue]: https://github.com/kubernetes-sigs/cluster-api-provider-digitalocean/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22good+first+issue%22
[cluster_api]: https://github.com/kubernetes-sigs/cluster-api
[kops]: https://github.com/kubernetes/kops
[kubicorn]: http://kubicorn.io/
[tilt]: https://tilt.dev
[cluster_api_tilt]: https://master.cluster-api.sigs.k8s.io/developer/tilt.html