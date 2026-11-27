# Building machine images

CAPDO creates droplets from DigitalOcean snapshots that already contain the
Kubernetes components (kubeadm, kubelet, containerd). These snapshots are built
with [image-builder](https://github.com/kubernetes-sigs/image-builder/tree/main/images/capi).

This page describes how to build them in any DigitalOcean account, copy them to
other regions, and use them in cluster templates or the e2e tests. The scripts
live in [`hack/images`](../hack/images).

## Prerequisites

- [`doctl`](https://github.com/digitalocean/doctl#installing-doctl), `git`, `jq`, `make` and `python3`.
- On macOS, GNU sed: `brew install gnu-sed`. The scripts find it automatically.
- A DigitalOcean API token in `DIGITALOCEAN_ACCESS_TOKEN`. Both the scripts and
  packer read it from there.

The build creates a temporary droplet, SSH key and tag, then takes a snapshot. A
[custom-scoped token](https://docs.digitalocean.com/reference/api/scopes/) needs at least:

| Resource | Scopes |
|---|---|
| droplet | create, read, delete |
| ssh_key | create, read, delete |
| tag | create, read |
| snapshot, image | create, read |
| actions, regions, sizes | read |

The token used to *run* clusters, for example in CI, usually doesn't need `tag` or
`snapshot` scopes. Using a separate token for builds keeps it that way.

If you use `doctl` contexts, you can take the token from one without printing it:

```sh
export DIGITALOCEAN_ACCESS_TOKEN="$(yq -r '.["auth-contexts"]["<context>"]' "<doctl config.yaml>")"
```

On macOS the doctl config is in `~/Library/Application Support/doctl/config.yaml`,
and on Linux it's in `~/.config/doctl/config.yaml`.

## Build images

```sh
hack/images/build-do-images.sh -r sfo2 1.34.13 1.35.11
```

- One snapshot is built per Kubernetes version, using image-builder's
  `build-do-<os>` target. The default OS is `ubuntu-2404`; change it with `-o ubuntu-2204`.
- `-r` sets the region the build runs in, which is where the snapshot is stored.
  Pick the region you'll create clusters in, or copy the snapshot later (see below).
- image-builder is cloned into a temporary directory at `main`. Set `IMAGE_BUILDER_REF`
  to use a tag or commit, or pass `-d <dir>` to use an existing checkout.
- Each build takes about 15–25 minutes. At the end the script prints the snapshot IDs:

```
=== Snapshots
123456789  Cluster API Kubernetes v1.34.13 on Ubuntu 24.04
123456790  Cluster API Kubernetes v1.35.11 on Ubuntu 24.04
```

Use the current patch release of each Kubernetes minor you need (see
`https://dl.k8s.io/release/stable-<minor>.txt`), and check that the Cluster API
version you run supports it as a workload cluster version
([support matrix](https://cluster-api.sigs.k8s.io/reference/versions)).

## Copy images to other regions

Snapshots are regional. To make them available everywhere:

```sh
hack/images/transfer-do-images.sh 123456789 123456790            # all available regions
hack/images/transfer-do-images.sh -r nyc3,ams3 123456789         # specific regions
```

Regions a snapshot is already in are skipped, and transfers run in parallel. Each
region you copy to is billed as separate snapshot storage.

## Use the images

Set the snapshot ID in your DOMachineTemplate (`spec.template.spec.image`), or, for the
templates in [`templates/`](../templates), in `DO_CONTROL_PLANE_MACHINE_IMAGE` and
`DO_NODE_MACHINE_IMAGE`. The Kubernetes version in the cluster must match the image.

### e2e tests

The e2e configuration in [`test/e2e/config/digitalocean-ci.yaml`](../test/e2e/config/digitalocean-ci.yaml) uses:

| Variable | Image |
|---|---|
| `KUBERNETES_VERSION`, `DO_CONTROL_PLANE_MACHINE_IMAGE`, `DO_NODE_MACHINE_IMAGE` | Default version for all specs |
| `KUBERNETES_VERSION_UPGRADE_FROM`, `KUBERNETES_IMAGE_UPGRADE_FROM` | Upgrade test: starting version |
| `KUBERNETES_VERSION_UPGRADE_TO`, `KUBERNETES_IMAGE_UPGRADE_TO` | Upgrade test: target version (one minor later) |

The upgrade pair must be adjacent minors, and both must be supported by the Cluster API
version the e2e tests use. Keep a comment with the Kubernetes version and OS next to each
image ID.

## Clean up

When no configuration references a snapshot anymore, delete it:

```sh
doctl compute snapshot delete <snapshot-id>
```
