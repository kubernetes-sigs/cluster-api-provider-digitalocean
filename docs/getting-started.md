# Getting started

## Prerequisites

- Linux or MacOS (Windows isn't supported at the moment).
- A [DigitalOcean][DigitalOcean] Account
- Install [kubectl][kubectl]
- Install [kustomize][kustomize] `v3.1.0+`
- [Packer][Packer] and [Ansible][Ansible] to build images
- Make to use `Makefile` targets
- A management cluster. You can use either a VM, container or existing Kubernetes cluster as management cluster.
   - If you want to use VM, install [Minikube][Minikube], version 0.30.0 or greater. Also install a [driver][Minikube Driver]. For Linux, we recommend `kvm2`. For MacOS, we recommend `VirtualBox`.
   - If you want to use a container, install [Kind][kind].
   - If you want to use an existing Kubernetes cluster, prepare a kubeconfig which for this cluster.
- Install [doctl][doctl] (optional)

## Setup Environment

```bash
# Export the DigitalOcean access token and region
export DIGITALOCEAN_ACCESS_TOKEN=<access_token>
export DO_REGION=<region>

# Init doctl
doctl auth init --access-token ${DIGITALOCEAN_ACCESS_TOKEN}
```

## Building images

Clone the image builder repository if you haven't already.

    $ git clone https://sigs.k8s.io/image-builder.git image-builder

Change directory to images/capi within the image builder repository

    $ cd image-builder/images/capi

Run the Make target to generate DigitalOcean images.

    $ make build-do-default

Check the image already available in your account.

    $ doctl compute image list-user


## Cluster Creation

> We assume you already have a running a management cluster

```bash
export CLUSTER_NAME=capdo-quickstart # change this name as you prefer.
export MACHINE_IMAGE=<image-id> # created in the step above.
```

For the purpose of this tutorial, we’ll name our cluster `capdo-quickstart`.

Generate examples files.

```
make generate-examples
```

Install core-component(CAPI & CABPK) and provider-component (CAPDO)

```
kubectl apply -f examples/_out/core-components.yaml
kubectl apply -f examples/_out/provider-components.yaml
```

Create cluster and control plane machine

```
kubectl apply -f examples/_out/cluster.yaml
kubectl apply -f examples/_out/controlplane.yaml
```

After the controlplane is up and running, Retrive cluster kubeconfig

```bash
kubectl --namespace=default get secret/${CLUSTER_NAME}-kubeconfig -o json \
  | jq -r .data.value \
  | base64 --decode \
  > ./${CLUSTER_NAME}.kubeconfig
```

Deploy a CNI solution, Calico is used here as an example.

```bash
kubectl --kubeconfig=./${CLUSTER_NAME}.kubeconfig apply -f https://docs.projectcalico.org/v3.8/manifests/calico.yaml
```

Deploy DigitalOcean Cloud Controller Manager

```bash
kubectl --kubeconfig=./${CLUSTER_NAME}.kubeconfig apply -f examples/digitalocean-ccm.yaml
```

Optional Deploy DigitalOcean CSI

```bash
kubectl --kubeconfig=./${CLUSTER_NAME}.kubeconfig apply -f examples/digitalocean-csi.yaml
```

Check the status of control-plane using kubectl get nodes

```bash
kubectl --kubeconfig=./${CLUSTER_NAME}.kubeconfig get nodes
```

Finishing up, we’ll create a single node MachineDeployment.

```
kubectl apply -f examples/_out/machinedeployment.yaml
```

<!-- References -->
[kubectl]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[kustomize]: https://github.com/kubernetes-sigs/kustomize/releases
[kind]: https://github.com/kubernetes-sigs/kind#installation-and-usage
[doctl]: https://github.com/digitalocean/doctl#installing-doctl
[Minikube]: https://kubernetes.io/docs/tasks/tools/install-minikube/
[Minikube Driver]: https://github.com/kubernetes/minikube/blob/master/docs/drivers.md
[Packer]: https://www.packer.io/intro/getting-started/install.html
[Ansible]: https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html
[DigitalOcean]: https://cloud.digitalocean.com/