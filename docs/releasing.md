# Releasing Guide

## Prerequisites

- Make sure your git has following set of repositories ("remotes")

```shell
$ git remote -v
origin	    git@github.com:{YOUR_GITHUB_USERNAME}/cluster-api-provider-digitalocean.git (fetch)
origin	    git@github.com:{YOUR_GITHUB_USERNAME}/cluster-api-provider-digitalocean.git (push)
upstream	git@github.com:kubernetes-sigs/cluster-api-provider-digitalocean.git (fetch)
upstream	git@github.com:kubernetes-sigs/cluster-api-provider-digitalocean.git (push)
```

- Make sure your working tree (this project) is clean from untracked files

```shell
$ git status
On branch master
nothing to commit, working tree clean
```

## Release Process

1. Create an env var `RELEASE_VERSION=v0.x.x` Ensure the version is prefixed with a `v`
2. If this is a new minor release, create a new release branch and push to github, for example

```shell
git checkout -b release-0.4
git push upstream release-0.4
```

Otherwise, switch to it

```shell
git checkout release-0.4
```

3. Tag the repository and push the tag

```shell
git tag -a $RELEASE_VERSION -m $RELEASE_VERSION
git push upstream $RELEASE_VERSION
```

The cloudbuild will automatically trigger the build and push the images to the [staging registry][staging-registry], Once the image is available, promote the image to [production registry][production registry]. How to promote the image can be found [here][image-promoter]

4. Create a draft release in github and associate it with the tag that was just created.
5. Run `make release RELEASE_TAG=$RELEASE_VERSION` to generate release manifests in the `./out` directory.
6. Attach files under `./out` directory to the draft release.
7. Publish release.

[staging-registry]: https://gcr.io/k8s-staging-cluster-api-do
[production registry]: https://us.gcr.io/k8s-artifacts-prod/cluster-api-do
[image-promoter]: https://github.com/kubernetes/k8s.io/tree/master/k8s.gcr.io#image-promoter
