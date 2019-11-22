# Contributing guidelines

This document contains guidelines for contributing to the `cluster-api-provider-digitalocean` project.

## Code Style Guidelines

In order for the pull request to get accepted the code must pass `gofmt`, `govet` and `golangci-lint` checks. You can run all checks by invoking `make check`.

## Dependency Management

This project uses [`dep`](https://github.com/golang/dep) for dependency management. Before pushing the code to GitHub, make sure your vendor directory is in sync by running `dep ensure` or otherwise CI will fail.

## Testing

Unit tests can be run by invoking `make test-unit`. Integration and End-to-End tests are currently not implemented.

## Building

The following Make targets can be used to build controllers:

* `all` runs code generation, `dep ensure`, builds `machine-controller`, `cluster-controller` and `clusterctl` binaries, as well as builds `machine-controller` and `cluster-controller` Docker images
* `compile` compiles `clusterctl`, `machine-controller` and `cluster-controller`, and stores them in the `./bin` directory
* `install` compiles `clusterctl`, `machine-controller` and `cluster-controller`, and stores them in the `$GOPTAH/bin` directory

## Versioning

This project follows [semantic versioning 2.0.0](https://semver.org/).

GitHub releases follow notation of `v0.x.y`. Docker images doesn't have the `v` prefix, e.g. `0.x.y`. Development versions of Docker images have the `-dev` suffix.

## Docker Images

We host the following images on `quay.io`:

* [`quay.io/kubermatic/digitalocean-machine-controller`](https://quay.io/repository/kubermatic/digitalocean-machine-controller)
* [`quay.io/kubermatic/digitalocean-cluster-controller`](https://quay.io/repository/kubermatic/digitalocean-cluster-controller)

### Building images locally

The `make images` and `make images-dev` commands can be used to build Docker images locally. The `images-dev` target build a development version of the image, annotated with the `-dev` suffix.

### Pushing images to `quay.io`

The `make push` and `make push-dev` targets build images and then push them to `quay.io` repositories used by the project.

**NOTE:** Make sure to set appropriate tag name for images in [`machine-controller`](./cmd/machine-controller/Makefile) and in [`cluster-controller`](./cmd/cluster-controller/Makefile) Makefiles, so you don't overwrite existing images!

## Release process

This section of the document describes how to handle the release process. Releasing new version of Cluster-API provider for DigitalOcean requires write access to GitHub and Quay.io repositories.

### Creating tag

The first step is to create a new Git tag and then push it to GitHub. The following Git command creates new tag based on current branch/latest commit:

```bash
git tag -as v0.x.y -m "Release v0.x.y of Cluster API provider for DigitalOcean"
```

If you want to specify the commit hash to be used for the tag, you can use command such as:

```bash
git tag -as v0.x.y <short-commit-hash> -m "Release v0.x.y of Cluster API provider for DigitalOcean"
```

Once the tag is created you need to push it to GitHub, such as:

```bash
git push --tags
```

### Building binaries

The next step is to build and compress binaries that will be uploaded to GitHub. Note that only `clusterctl` binaries are uploaded to GitHub, while `machine-controller` and `cluster-controller` are only published as Docker images.

#### Building MacOS `clusterctl` binary

```bash
# Build MacOS 64-bit clusterctl binary
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./release/clusterctl sigs.k8s.io/cluster-api-provider-digitalocean/clusterctl

# Create tarball
tar -czf ./release/clusterctl-darwin-amd64.tar.gz clusterctl
```

#### Building Linux `clusterctl` binary

```
# Build Linux 64-bit clusterctl binary
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o ./release/clusterctl-linux-amd64 sigs.k8s.io/cluster-api-provider-digitalocean/cmd/cluster-controller

# Create tarball
tar -czf ./release/clusterctl-linux-amd64.tar.gz clusterctl
```

### Pushing Docker Images

The Docker images can be pushed by invoking `make push`.

**NOTE:** Make sure to set appropriate tag name for images in [`machine-controller`](./cmd/machine-controller/Makefile) and in [`cluster-controller`](./cmd/cluster-controller/Makefile) Makefiles, so you don't overwrite existing images!

### Creating GitHub release

The last step is to create a GitHub release and push `clusterctl` binaries to GitHub.

To learn about GitHub Releases, check out the [official GitHub Releases documentation](https://help.github.com/articles/creating-releases/).

#### Writing Release Notes

When creating a release, you need to specify the release notes for that release. The release notes should include:

- Changelog, which can be generated using [`gchl`](https://github.com/kubermatic/gchl)
- List of Docker images that are relevant to the release
- SHA256 checksums of tarballs

[The following example](https://github.com/kubernetes-sigs/cluster-api-provider-digitalocean/releases/tag/v0.2.0) can be used while writing release notes.
