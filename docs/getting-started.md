# Getting started

## Prerequisites

- A [DigitalOcean][DigitalOcean] Account
- Install [clusterctl][clusterctl]
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
# Export the DigitalOcean access token
$ export DIGITALOCEAN_ACCESS_TOKEN=<access_token>

# Init doctl
$ doctl auth init --access-token ${DIGITALOCEAN_ACCESS_TOKEN}
```

## Building images

Clone the image builder repository if you haven't already:

    $ git clone https://github.com/kubernetes-sigs/image-builder.git

Change directory to images/capi within the image builder repository:

    $ cd image-builder/images/capi

Choose a DigitalOcean image build target from the list returned by `make | grep build-do` and generate a DigitalOcean image (choosing Ubuntu in the example below):

    $ make build-do-ubuntu-2004

Verify that the image is available in your account and remember the corresponding image ID:

    $ doctl compute image list-user


## Initialize the management cluster

```bash
$ export DO_B64ENCODED_CREDENTIALS="$(echo -n "${DIGITALOCEAN_ACCESS_TOKEN}" | base64 | tr -d '\n')"

# Initialize a management cluster with digitalocean infrastructure provider.
$ clusterctl init --infrastructure digitalocean
```

The output will be similar to this:

```bash
Fetching providers
Installing cert-manager Version="v0.16.1"
Waiting for cert-manager to be available...
Installing Provider="cluster-api" Version="v0.3.11" TargetNamespace="capi-system"
Installing Provider="bootstrap-kubeadm" Version="v0.3.11" TargetNamespace="capi-kubeadm-bootstrap-system"
Installing Provider="control-plane-kubeadm" Version="v0.3.11" TargetNamespace="capi-kubeadm-control-plane-system"
Installing Provider="infrastructure-digitalocean" Version="v0.4.0" TargetNamespace="capdo-system"

Your management cluster has been initialized successfully!

You can now create your first workload cluster by running the following:

  clusterctl config cluster [name] --kubernetes-version [version] | kubectl apply -f -

```

## Creating a workload cluster

Setting up environment variable

```bash
$ export DO_REGION=<region>
$ export DO_SSH_KEY_FINGERPRINT=<your-ssh-key-fingerprint>
$ export DO_CONTROL_PLANE_MACHINE_TYPE=<droplet-size>
$ export DO_CONTROL_PLANE_MACHINE_IMAGE=<image-id> # created in the step above.
$ export DO_NODE_MACHINE_TYPE=<droplet-size>
$ export DO_NODE_MACHINE_IMAGE=<image-id> # created in the step above.
```
Generate templates for creating workload clusters.

```bash
$ clusterctl config cluster capdo-quickstart \
    --infrastructure digitalocean \
    --kubernetes-version v1.17.11 \
    --control-plane-machine-count 1 \
    --worker-machine-count 3 > capdo-quickstart-cluster.yaml
```

*You may need to inspect and make some changes to the generated template.*

Create the workload cluster on the management cluster.

```bash
$ kubectl apply -f capdo-quickstart-cluster.yaml
```

You can see the workload cluster resources using

```bash
$ kubectl get cluster-api
```

> Note: The control planes won’t be ready until you install the CNI and DigitalOcean Cloud Controller Manager.

To verify the first control plane is up:

```bash
$ kubectl get kubeadmcontrolplane

NAME                                                                               INITIALIZED   API SERVER AVAILABLE   VERSION    REPLICAS   READY   UPDATED   UNAVAILABLE
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/capdo-quickstart-control-plane   true                                 v1.17.11   1                  1         1
```

After the first control plane node has initialized status, you can retrieve the workload cluster Kubeconfig:

```bash
$ clusterctl get kubeconfig capdo-quickstart > capdo-quickstart.kubeconfig
```

You can verify the kubernetes node in the workload cluster

```bash
$ KUBECONFIG=capdo-quickstart.kubeconfig kubectl get node

NAME                                   STATUS     ROLES    AGE     VERSION
capdo-quickstart-control-plane-pt926   NotReady   master   10m     v1.17.11
capdo-quickstart-md-0-2vnwv            NotReady   <none>   5m31s   v1.17.11
capdo-quickstart-md-0-5295f            NotReady   <none>   5m30s   v1.17.11
capdo-quickstart-md-0-pm8np            NotReady   <none>   5m28s   v1.17.11
```

### Deploy CNI

Calico is used here as an example.

```bash
$ KUBECONFIG=capdo-quickstart.kubeconfig kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
```

### Deploy DigitalOcean CCM and CSI

```bash
# Create digitalocean secret
$ KUBECONFIG=capdo-quickstart.kubeconfig kubectl create secret generic digitalocean --namespace kube-system --from-literal access-token=$DIGITALOCEAN_ACCESS_TOKEN

# Deploy DigitalOcean Cloud Controller Manager
$ KUBECONFIG=capdo-quickstart.kubeconfig kubectl apply -f https://raw.githubusercontent.com/digitalocean/digitalocean-cloud-controller-manager/master/releases/v0.1.27.yml

# Deploy DigitalOcean CSI (optional)
$ KUBECONFIG=capdo-quickstart.kubeconfig kubectl apply -f https://raw.githubusercontent.com/digitalocean/csi-digitalocean/master/deploy/kubernetes/releases/csi-digitalocean-v1.3.0.yaml
```

After CNI and CCM deployed, your workload cluster nodes should be in the ready state. You can verify using:

```bash
$ KUBECONFIG=capdo-quickstart.kubeconfig kubectl get node

NAME                                   STATUS   ROLES    AGE   VERSION
capdo-quickstart-control-plane-pt926   Ready    master   25m   v1.17.11
capdo-quickstart-md-0-2vnwv            Ready    <none>   21m   v1.17.11
capdo-quickstart-md-0-5295f            Ready    <none>   21m   v1.17.11
capdo-quickstart-md-0-pm8np            Ready    <none>   21m   v1.17.11
```

## Deleting a workload cluster

You can delete the workload cluster from the management cluster using:

```bash
$ kubectl delete cluster capdo-quickstart
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
[clusterctl]: https://github.com/kubernetes-sigs/cluster-api/releases
