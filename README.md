# Kubernetes cluster-api-provider-digitalocean Project [![Build Status](https://circleci.com/gh/kubermatic/cluster-api-provider-digitalocean.svg?style=shield)](https://circleci.com/gh/kubermatic/cluster-api-provider-digitalocean/) [![Go Report Card](https://goreportcard.com/badge/github.com/kubermatic/cluster-api-provider-digitalocean)](https://goreportcard.com/report/github.com/kubermatic/cluster-api-provider-digitalocean)

This repository hosts a concrete implementation of a provider for [DigitalOcean](https://www.digitalocean.com/) for the [cluster-api project](https://github.com/kubernetes-sigs/cluster-api).

## Project Status

This project is currently Work-in-Progress and may not be production ready. There is no backwards-compatibility guarantee at this point.
Checkout the Features portion of the README for details about the project status.

## Getting Started

### Prerequisites

In order to create a cluster using `clusterctl`, you need the following tools installed on your local machine:

* `kubectl`, which can be done by following [this tutorial](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [`minikube`](https://kubernetes.io/docs/tasks/tools/install-minikube/) and the appropriate [`minikube` driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md). We recommend `kvm2` for Linux and `virtualbox` for macOS
* [DigitalOcean API Access Token generated](https://www.digitalocean.com/docs/api/create-personal-access-token/) and set as the `DIGITALOCEAN_ACCESS_TOKEN` environment variable,
* Go toolchain [installed and configured](https://golang.org/doc/install), needed in order to compile the `clusterctl` binary,
* `cluster-api-provider-digitalocean` repository cloned:
```bash
git clone https://github.com/kubermatic/cluster-api-provider-digitalocean $(go env GOPATH)/src/github.com/kubermatic/cluster-api-provider-digitalocean
```

### Building `clusterctl`

The `clusterctl` tool is used to bootstrap an Kubernetes cluster from zero. Currently, we have not released binaries, so you need to compile it manually.

Compiling is done by invoking the `compile` Make target:
```bash
make compile
```

This command generates three binaries: `clusterctl`, `machine-controller` and `cluster-controller`, in the `./bin` directory. In order to bootstrap the cluster, you only need the `clusterctl` binary.

The `clusterctl` can also be compiled manually, such as:
```bash
cd $(go env GOPATH)/github.com/kubermatic/cluster-api-provider-digitalocean/clusterctl
go install
```

## Creating a Cluster

To create your first cluster using `cluster-api-provider-digitalocean`, you need to use the `clusterctl`. It takes the following four manifests as input:

* `cluster.yaml` - defines Cluster properties, such as Pod and Services CIDR, Services Domain, etc.
* `machines.yaml` - defines Machine properties, such as machine size, image, tags, SSH keys, enabled features, as well as what Kubernetes version will be used for each machine.
* `provider-components.yaml` - contains deployment manifest for controllers, userdata used to bootstrap machines, a secret with SSH key for the `machine-controller` and a secret with DigitalOcean API Access Token.
* [Optional] `addons.yaml` - used to deploy additional components once the cluster is bootstrapped, such as [DigitalOcean Cloud Controller Manager](https://github.com/digitalocean/digitalocean-cloud-controller-manager) and [DigitalOcean CSI plugin](https://github.com/digitalocean/csi-digitalocean).

The manifests can be generated automatically by using the [`generate-yaml.sh`](./clusterctl/examples/digitalocean/generate-yaml.sh) script, located in the `clusterctl/examples/digitalocean` directory:
```bash
cd clusterctl/examples/digitalocean
./generate-yaml.sh
cd ../..
```
The result of the script is an `out` directory with generated manifests and a generated SSH key to be used by the `machine-controller`. More details about how it generates manifests and how to customize them can be found in the [README file in `clusterctl/examples/digitalocean` directory](./clusterctl/examples/digitalocean).

As a temporary workaround for a bug with SSH keys, you need to copy the generated private key to the `/etc/sshkeys` directory, name it `private`, and give it read permissions for your user:
```
sudo mkdir /etc/sshkeys
sudo cp clusterctl/examples/digitalocean/out/test-1_rsa /etc/sshkeys/private
sudo chown $USER /etc/sshkeys/private
```
Issue [#39](https://github.com/kubermatic/cluster-api-provider-digitalocean/issues/39) tracks this problem.

Once you have manifests generated, you can create a cluster using the following command. Make sure to replace the value of `vm-driver` flag with the name of your actual `minikube` driver.
```bash
./bin/clusterctl create cluster \
    --provider digitalocean \
    --vm-driver kvm2 \
    -c ./clusterctl/examples/digitalocean/out/cluster.yaml \
    -m ./clusterctl/examples/digitalocean/out/machines.yaml \
    -p ./clusterctl/examples/digitalocean/out/provider-components.yaml \
    -a ./clusterctl/examples/digitalocean/out/addons.yaml
```

More details about the `create cluster` command can be found by invoking help:
```bash
./bin/clusterctl create cluster --help
```

The `clusterctl`'s workflow is:
* Create a Minikube bootstrap cluster,
* Deploy the `cluster-api-controller`, `digitalocean-machine-controller` and `digitalocean-cluster-controller`, on the bootstrap cluster,
* Create a Master, download `kubeconfig` file, and deploy controllers on the Master,
* Create other specified machines (nodes),
* Deploy addon components ([`digitalocean-cloud-controller-manager`](https://github.com/digitalocean/digitalocean-cloud-controller-manager) and [`csi-digitalocean`](https://github.com/digitalocean/csi-digitalocean)),
* Remove the local Minikube cluster.

To learn more about the process and how each component work, check out the [diagram in `cluster-api` repostiory](https://github.com/kubernetes-sigs/cluster-api#what-is-the-cluster-api).

### Interacting With Your New Cluster

`clusterctl` downloads the `kubeconfig` file in your current directory from the cluster automatically. You can use it with `kubectl` to interact with your cluster:
```bash
kubectl --kubeconfig kubeconfig get nodes
kubectl --kubeconfig kubeconfig get all --all-namespaces
```

## Upgrading the Cluster

Upgrading Master is currently not possible automatically (by updating the Machine object) as Update method is not fully implemented. More details can be found in [issue #32](https://github.com/kubermatic/cluster-api-provider-digitalocean/issues/32).

Workers can be upgraded by updating the appropriate Machine object for that node. Workers are upgraded by replacing nodes—first the old node is removed and then a new one with new properties is created.

To ensure non-disturbing maintenance we recommend having at least 2+ worker nodes at the time of upgrading, so another node can take tasks from the node being upgraded. The node that is going to be upgraded should be marked unschedulable and drained, so there are no pods running and scheduled.

```bash
# Make node unschedulable.
kubectl --kubeconfig kubeconfig cordon <node-name>
# Drain all pods from the node.
kubectl --kubeconfig kubeconfig drain <node-name>
```

Now that you prepared node for upgrading, you can proceed with editing the Machine object:
```bash
kubectl --kubeconfig kubeconfig edit machine <node-name>
```

This opens the Machine manifest such as the following one, in your default text editor. You can choose editor by setting the `EDITOR` environment variable.

There you can change machine properties, including Kubernetes (`kubelet`) version.

```yaml
apiVersion: cluster.k8s.io/v1alpha1
kind: Machine
metadata:  
  creationTimestamp: 2018-09-14T11:02:16Z
  finalizers:
  - machine.cluster.k8s.io
  generateName: digitalocean-fra1-node-
  generation: 3
  labels:
    set: node
  name: digitalocean-fra1-node-tzzgm
  namespace: default
  resourceVersion: "5"
  selfLink: /apis/cluster.k8s.io/v1alpha1/namespaces/default/machines/digitalocean-fra1-node-tzzgm
  uid: a41f83ad-b80d-11e8-aeef-0242ac110003
spec:
  metadata:
    creationTimestamp: null
  providerConfig:
    ValueFrom: null
    value:
      backups: false
      image: ubuntu-18-04-x64
      ipv6: false
      monitoring: true
      private_networking: true
      region: fra1
      size: s-2vcpu-2gb
      sshPublicKeys:
      - ssh-rsa AAAA
      tags:
      - machine-2
  versions:
    kubelet: 1.11.3
status:
  lastUpdated: null
  providerStatus: null
```

Saving changes to the Machine object deletes the old machine and then creates a new one. After some time, a new machine will be part of your Kubernetes cluster. You can track progress by watching list of nodes. Once new node appears and is Ready, upgrade has finished.

```bash
watch -n1 kubectl get nodes
```

## Deleting the Cluster

To delete Master and confirm all relevant resources are deleted from the cloud, we're going to use [`doctl`—DigitalOcean CLI](https://github.com/digitalocean/doctl). You can also use DigitalOcean Cloud Control Panel or API instead of `doctl.

First, save the Droplet ID of Master, as we'll use it later to delete the control plane machine:

```bash
export MASTER_ID=$(kubectl --kubeconfig=kubeconfig get machines -l set=master -o jsonpath='{.items[0].metadata.annotations.droplet-id}')
```

Now, delete all Workers in the cluster by removing all Machine object with label `set=node`:

```
kubectl --kubeconfig=kubeconfig delete machines -l set=node
```

You can confirm are nodes deleted by checking list of nodes. After some time, only Master should be present:

```bash
kubectl --kubeconfig=kubeconfig get nodes
```

Then, delete all Services and PersistentVolumeClaims, so all Load Balancers and Volumes in the cloud are deleted:

```bash
kubectl --kubeconfig=kubeconfig delete svc --all
kubectl --kubeconfig=kubeconfig delete pvc --all
```

Finally, we can delete the Master using `doctl` and `$MASTER_ID` environment variable we set earlier:

```bash
doctl compute droplet delete $MASTER_ID
```

You can use `doctl` to confirm that Droplets, Load Balancers and Volumes relevant to the cluster are deleted:

```bash
doctl compute droplet list
doctl compute load-balancer list
doctl compute volume list
```

## Features

The core of the machine-controller is fully implemented and can be used to manage machines. The `clusterctl` can be used to bootstrap the cluster.

The following features are to be implemented soon:
* Cluster Controller - currently `cluster-controller` does nothing. Follow the issue [#31](https://github.com/kubermatic/cluster-api-provider-digitalocean/issues/31) for more details,
* Master upgrades - updating master instances have no effect. Follow the issue [#32](https://github.com/kubermatic/cluster-api-provider-digitalocean/issues/32) for more details,

Follow [the issue tracker](https://github.com/kubermatic/cluster-api-provider-digitalocean/issues) for details about other upcoming features.

## Development

This portion of the README file contains information about the development process.

### Building Controllers

The following Make targets can be used to build controllers:
* `make compile` compiles all three components: `clusterctl`, `machine-controller` and `cluster-controller`
* `make build` runs `dep ensure` and installs `machine-controller` and `cluster-controller`
* `make all` runs code generation, `dep ensure`, install `machine-controller` and `cluster-controller`, and build `machine-controller` and `cluster-controller` Docker images

The `make generate` target runs code generation.

### Building and Pushing Images

We're using two Quay.io repositories for hosting cluster-api-provider-digitalocean images: `quay.io/kubermatic/digitalocean-machine-controller` and `quay.io/kubermatic/digitalocean-cluster-controller`.

Before building and pushing images to the registry, make sure to replace the version number in the [`machine-controller`](./cmd/machine-controller/Makefile) and [`cluster-controller`](./cmd/machine-controller/Makefile) Makefiles, so you don't overwrite existing images!

The images are built using the `make images` target, which runs `dep ensures` and build images. There is also `make images-nodep` target, which doesn't run `dep ensure`, and is used by the CI.

The images can be pushed to the registry using the `make push` and `make push-nodep` targets. Both targets build images, so you don't need to run `make images` before.

### Code Style Guidelines

In order for the pull request to get accepted, the code must pass `gofmt`, `govet` and `gometalinter` checks. You can run all checks by invoking the `make check` Makefile target.

### Dependency Management

This project uses [`dep`](https://github.com/golang/dep) for dependency management. Before pushing the code to GitHub, make sure your vendor directory is in sync by running `dep ensure`, or otherwise CI will fail.

### Testing

Unit tests can be invoked by running the `make test-unit` Makefile target. Integration and End-to-End tests are currently not implemented.

