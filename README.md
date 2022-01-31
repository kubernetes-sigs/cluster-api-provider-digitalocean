# Kubernetes Cluster API Provider DigitalOcean

<p align="center"><img alt="capi" src="https://github.com/kubernetes-sigs/cluster-api/raw/master/docs/book/src/images/introduction.png" width="160x" /><img alt="capi" src="https://upload.wikimedia.org/wikipedia/commons/f/ff/DigitalOcean_logo.svg" width="192x" /></p>
<p align="center">
<!-- prow build badge, godoc, and go report card-->
</a> <a href="https://pkg.go.dev/sigs.k8s.io/cluster-api-provider-digitalocean"><img src="https://pkg.go.dev/badge/sigs.k8s.io/cluster-api-provider-digitalocean.svg" alt="Go Reference"></a> <a href="https://goreportcard.com/report/sigs.k8s.io/cluster-api-provider-digitalocean"><img alt="Go Report Card" src="https://goreportcard.com/badge/sigs.k8s.io/cluster-api-provider-digitalocean" /></a></p>

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

This project is currently a work-in-progress, in an Alpha state, so it may not be production ready. There is no backwards-compatibility guarantee at this point. For more details on the roadmap and upcoming features, check out [the project's issue tracker on GitHub][issue].

## Launching a Kubernetes cluster on DigitalOcean

Check out the [getting started guide](./docs/getting-started.md) for launching a cluster on DigitalOcean.

## Features

- Native Kubernetes manifests and API
- Support for single and multi-node control plane clusters
- Choice of Linux distribution (as long as a current [cloud-init](https://cloudinit.readthedocs.io/en/latest/topics/examples.html) is available)

------

## Compatibility with Cluster API and Kubernetes Versions

This provider's versions are compatible with the following versions of Cluster API:

|                                       | Cluster API v1alpha1 (v0.1) | Cluster API v1alpha2 (v0.2) | Cluster API v1alpha3 (v0.3) | Cluster API v1alpha4 (v0.4) | Cluster API v1 (v1.0) |
| ------------------------------------- | --------------------------- | --------------------------- | --------------------------- |-----------------------------| --------------------- |
| DigitalOcean Provider v1alpha1 (v0.1) | ✓                           |                             |                             |                             |                       |
| DigitalOcean Provider v1alpha1 (v0.2) | ✓                           |                             |                             |                             |                       |
| DigitalOcean Provider v1alpha2 (v0.3) |                             | ✓                           |                             |                             |                       |
| DigitalOcean Provider v1alpha3 (v0.4) |                             |                             | ✓                           |                             |                       |
| DigitalOcean Provider v1alpha4 (v0.5) |                             |                             |                             | ✓                           |                       |
| DigitalOcean Provider v1       (v1.0) |                             |                             |                             |                             | ✓                     |

This provider's versions are able to install and manage the following versions of Kubernetes:

|                 | DigitalOcean Provider v1alpha1 (v0.1) | DigitalOcean Provider v1alpha1 (v0.2) | DigitalOcean Provider v1alpha2 (v0.3) | DigitalOcean Provider v1alpha3 (v0.4) | DigitalOcean Provider v1alpha4 (v0.5) | DigitalOcean Provider v1 (v1.0) |
| --------------- | ------------------------------------- | ------------------------------------- | ------------------------------------- | ------------------------------------- | ------------------------------------- | ------------------------------- |
| Kubernetes 1.19 |                                       |                                       | ✓                                     | ✓                                     | ✓                                     | ✓                               |
| Kubernetes 1.20 |                                       |                                       | ✓                                     | ✓                                     | ✓                                     | ✓                               |
| Kubernetes 1.21 |                                       |                                       | ✓                                     | ✓                                     | ✓                                     | ✓                               |
| Kubernetes 1.22 |                                       |                                       | ✓                                     | ✓                                     | ✓                                     | ✓                               |
| Kubernetes 1.23 |                                       |                                       |                                       | ✓                                     | ✓                                     | ✓                               |

**NOTE:** As the versioning for this project is tied to the versioning of Cluster API, future modifications to this policy may be made to more closely align with other providers in the Cluster API ecosystem.


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
