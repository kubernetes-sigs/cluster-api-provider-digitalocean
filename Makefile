# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=/usr/bin/env bash

.DEFAULT_GOAL:=help

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://proxy.golang.org
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on

# Directories.
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/bin)
BIN_DIR := bin
TEST_E2E_DIR := test/e2e
REPO_ROOT := $(shell git rev-parse --show-toplevel)
ARTIFACTS ?= $(REPO_ROOT)/_artifacts
E2E_DATA_DIR ?= $(ROOT_DIR)/$(TEST_E2E_DIR)/data
KUBETEST_CONF_PATH ?= $(abspath $(E2E_DATA_DIR)/kubetest/conformance.yaml)
GO_INSTALL = ./scripts/go_install.sh

# Files
E2E_CONF_FILE ?= $(REPO_ROOT)/test/e2e/config/digitalocean-dev.yaml
E2E_CONF_FILE_ENVSUBST := $(ROOT_DIR)/test/e2e/config/digitalocean-dev-envsubst.yaml

# Define Docker related variables. Releases should modify and double check these vars.
REGISTRY ?= gcr.io/$(shell gcloud config get-value project)
STAGING_REGISTRY := gcr.io/k8s-staging-cluster-api-do
PROD_REGISTRY := us.gcr.io/k8s-artifacts-prod/cluster-api-do
IMAGE_NAME ?= cluster-api-do-controller
CONTROLLER_IMG ?= $(REGISTRY)/$(IMAGE_NAME)
TAG ?= dev
ARCH ?= amd64
ALL_ARCH = amd64 arm arm64 ppc64le s390x

# Allow overriding manifest generation destination directory
MANIFEST_ROOT ?= config
CRD_ROOT ?= $(MANIFEST_ROOT)/crd/bases
WEBHOOK_ROOT ?= $(MANIFEST_ROOT)/webhook
RBAC_ROOT ?= $(MANIFEST_ROOT)/rbac


# Allow overriding the e2e configurations
GINKGO_FOCUS ?= Workload cluster creation
GINKGO_NODES ?= 3
GINKGO_NOCOLOR ?= false

# Allow overriding the imagePullPolicy
PULL_POLICY ?= Always

# Docker buildkit disabled for now until we figure out how to fix when
# building the image in the post stage
DOCKER_BUILDKIT ?= 0
export DOCKER_BUILDKIT

KIND_CLUSTER_NAME ?= capdo

## --------------------------------------
##@ Help
## --------------------------------------

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
##@ Testing
## --------------------------------------

$(ARTIFACTS):
	mkdir -p $@

.PHONY: test
test: generate lint ## Run tests
	go test -v ./api/... ./controllers/... ./cloud/...

.PHONY: test-e2e ## Run e2e tests using clusterctl
test-e2e: $(ENVSUBST) $(GINKGO) $(KIND) $(KUSTOMIZE) e2e-image ## Run e2e tests
	$(ENVSUBST) < $(E2E_CONF_FILE) > $(E2E_CONF_FILE_ENVSUBST) && \
	time $(GINKGO) -trace -progress -v -tags=e2e -focus=$(GINKGO_FOCUS) -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) ./test/e2e/... -- \
			-e2e.config="$(E2E_CONF_FILE_ENVSUBST)" \
			-e2e.artifacts-folder="$(ARTIFACTS)" $(E2E_ARGS)

.PHONY: test-conformance
test-conformance: $(ENVSUBST) $(GINKGO) $(KIND) $(KUSTOMIZE) e2e-image ## Run conformance test on workload cluster
	$(ENVSUBST) < $(E2E_CONF_FILE) > $(E2E_CONF_FILE_ENVSUBST) && \
	time $(GINKGO) -v -trace -stream -progress -tags=e2e -focus=$(GINKGO_FOCUS) -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) ./test/e2e/... -- \
			-e2e.config="$(E2E_CONF_FILE_ENVSUBST)" \
			-kubetest.config-file=$(KUBETEST_CONF_PATH) \
			-e2e.artifacts-folder="$(ARTIFACTS)" $(E2E_ARGS)

.PHONY: e2e-image
e2e-image:
	docker build --build-arg ldflags="$(LDFLAGS)" --tag="gcr.io/k8s-staging-cluster-api/capdo-manager:e2e" .

## --------------------------------------
##@ Binaries
## --------------------------------------

# Binaries.
CONTROLLER_GEN_VER := v0.3.0
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER)

CONVERSION_GEN_VER := v0.18.4
CONVERSION_GEN_BIN := conversion-gen
CONVERSION_GEN := $(TOOLS_BIN_DIR)/$(CONVERSION_GEN_BIN)-$(CONVERSION_GEN_VER)

ENVSUBST_BIN := envsubst
ENVSUBST := $(TOOLS_BIN_DIR)/$(ENVSUBST_BIN)-drone

GOLANGCI_LINT_VER := v1.32.2
GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER)

KUSTOMIZE_VER := v3.5.4
KUSTOMIZE_BIN := kustomize
KUSTOMIZE := $(TOOLS_BIN_DIR)/$(KUSTOMIZE_BIN)-$(KUSTOMIZE_VER)

# Using unreleased version until https://github.com/golang/mock/pull/405 is part of a release.
MOCKGEN_VER := 8a3d5958550701de9e6650b84b75a118771e7b49
MOCKGEN_BIN := mockgen
MOCKGEN := $(TOOLS_BIN_DIR)/$(MOCKGEN_BIN)-$(MOCKGEN_VER)

RELEASE_NOTES_VER := v0.3.4
RELEASE_NOTES_BIN := release-notes
RELEASE_NOTES := $(TOOLS_BIN_DIR)/$(RELEASE_NOTES_BIN)-$(RELEASE_NOTES_VER)

GINKGO_VER := v1.13.0
GINKGO_BIN := ginkgo
GINKGO := $(TOOLS_BIN_DIR)/$(GINKGO_BIN)-$(GINKGO_VER)

KUBECTL_VER := v1.16.13
KUBECTL_BIN := kubectl
KUBECTL := $(TOOLS_BIN_DIR)/$(KUBECTL_BIN)-$(KUBECTL_VER)



.PHONY: binaries
binaries: manager ## Builds and installs all binaries

.PHONY: manager
manager: ## Build manager binary.
	go build -o $(BIN_DIR)/manager .

## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(CONTROLLER_GEN): ## Build controller-gen from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) sigs.k8s.io/controller-tools/cmd/controller-gen $(CONTROLLER_GEN_BIN) $(CONTROLLER_GEN_VER)

$(CONVERSION_GEN): ## Build conversion-gen.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) k8s.io/code-generator/cmd/conversion-gen $(CONVERSION_GEN_BIN) $(CONVERSION_GEN_VER)

$(ENVSUBST): ## Build envsubst from tools folder.
	rm -f $(TOOLS_BIN_DIR)/$(ENVSUBST_BIN)*
	mkdir -p $(TOOLS_DIR) && cd $(TOOLS_DIR) && go build -tags=tools -o $(ENVSUBST) github.com/drone/envsubst/cmd/envsubst
	ln -sf $(ENVSUBST) $(TOOLS_BIN_DIR)/$(ENVSUBST_BIN)

.PHONY: $(ENVSUBST_BIN)
$(ENVSUBST_BIN): $(ENVSUBST) ## Build envsubst from tools folder.

$(GOLANGCI_LINT): ## Build golangci-lint from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint $(GOLANGCI_LINT_BIN) $(GOLANGCI_LINT_VER)

$(KUSTOMIZE): ## Build kustomize from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) sigs.k8s.io/kustomize/kustomize/v3 $(KUSTOMIZE_BIN) $(KUSTOMIZE_VER)

$(MOCKGEN): ## Build mockgen from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/golang/mock/mockgen $(MOCKGEN_BIN) $(MOCKGEN_VER)

$(RELEASE_NOTES): ## Build release notes.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) k8s.io/release/cmd/release-notes $(RELEASE_NOTES_BIN) $(RELEASE_NOTES_VER)

$(GINKGO): ## Build ginkgo.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/onsi/ginkgo/ginkgo $(GINKGO_BIN) $(GINKGO_VER)

$(KUBECTL): ## Build kubectl
	mkdir -p $(TOOLS_BIN_DIR)
	rm -f "$(KUBECTL)*"
	curl -fsL https://storage.googleapis.com/kubernetes-release/release/$(KUBECTL_VER)/bin/$(GOOS)/$(GOARCH)/kubectl -o $(KUBECTL)
	ln -sf "$(KUBECTL)" "$(TOOLS_BIN_DIR)/$(KUBECTL_BIN)"
	chmod +x "$(TOOLS_BIN_DIR)/$(KUBECTL_BIN)" "$(KUBECTL)"

## --------------------------------------
##@ Linting
## --------------------------------------

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint codebase
	$(GOLANGCI_LINT) run -v

lint-full: $(GOLANGCI_LINT) ## Run slower linters to detect possible issues
	$(GOLANGCI_LINT) run -v --fast=false

## --------------------------------------
##@ Generate
## --------------------------------------

.PHONY: modules
modules: ## Runs go mod to ensure proper vendoring.
	go mod tidy

.PHONY: generate
generate: ## Generate code
	$(MAKE) generate-go
	$(MAKE) generate-manifests

.PHONY: generate-go
generate-go: $(CONTROLLER_GEN) $(MOCKGEN) $(CONVERSION_GEN) ## Runs Go related generate targets
	go generate ./...
	$(CONTROLLER_GEN) \
		paths=./api/... \
		object:headerFile=./hack/boilerplate/boilerplate.generatego.txt

	$(CONVERSION_GEN) \
		--input-dirs=./api/v1alpha2 \
		--output-file-base=zz_generated.conversion \
		--go-header-file=./hack/boilerplate/boilerplate.generatego.txt

.PHONY: generate-manifests
generate-manifests: $(CONTROLLER_GEN) ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./api/... \
		crd:crdVersions=v1 \
		output:crd:dir=$(CRD_ROOT) \
		output:webhook:dir=$(WEBHOOK_ROOT) \
		webhook
	$(CONTROLLER_GEN) \
		paths=./controllers/... \
		output:rbac:dir=$(RBAC_ROOT) \
		rbac:roleName=manager-role

.PHONY: generate-examples
generate-examples: clean-examples ## Generate examples configurations to run a cluster.
	./examples/generate.sh

## --------------------------------------
##@ Docker
## --------------------------------------

.PHONY: docker-build
docker-build: ## Build the docker image for controller-manager
	docker build --pull --build-arg ARCH=$(ARCH) . -t $(CONTROLLER_IMG)-$(ARCH):$(TAG)
	MANIFEST_IMG=$(CONTROLLER_IMG)-$(ARCH) MANIFEST_TAG=$(TAG) $(MAKE) set-manifest-image
	$(MAKE) set-manifest-pull-policy

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(CONTROLLER_IMG)-$(ARCH):$(TAG)

## --------------------------------------
##@ Docker — All ARCH
## --------------------------------------

.PHONY: docker-build-all ## Build all the architecture docker images
docker-build-all: $(addprefix docker-build-,$(ALL_ARCH))

docker-build-%:
	$(MAKE) ARCH=$* docker-build

.PHONY: docker-push-all ## Push all the architecture docker images
docker-push-all: $(addprefix docker-push-,$(ALL_ARCH))
	$(MAKE) docker-push-manifest

docker-push-%:
	$(MAKE) ARCH=$* docker-push

.PHONY: docker-push-manifest
docker-push-manifest: ## Push the fat manifest docker image.
	## Minimum docker version 18.06.0 is required for creating and pushing manifest images.
	docker manifest create --amend $(CONTROLLER_IMG):$(TAG) $(shell echo $(ALL_ARCH) | sed -e "s~[^ ]*~$(CONTROLLER_IMG)\-&:$(TAG)~g")
	@for arch in $(ALL_ARCH); do docker manifest annotate --arch $${arch} ${CONTROLLER_IMG}:${TAG} ${CONTROLLER_IMG}-$${arch}:${TAG}; done
	docker manifest push --purge ${CONTROLLER_IMG}:${TAG}
	MANIFEST_IMG=$(CONTROLLER_IMG) MANIFEST_TAG=$(TAG) $(MAKE) set-manifest-image
	$(MAKE) set-manifest-pull-policy

.PHONY: set-manifest-image
set-manifest-image:
	$(info Updating kustomize image patch file for manager resource)
	sed -i'' -e 's@image: .*@image: '"${MANIFEST_IMG}:$(MANIFEST_TAG)"'@' ./config/manager/manager_image_patch.yaml

.PHONY: set-manifest-pull-policy
set-manifest-pull-policy:
	$(info Updating kustomize pull policy file for manager resource)
	sed -i'' -e 's@imagePullPolicy: .*@imagePullPolicy: '"$(PULL_POLICY)"'@' ./config/manager/manager_pull_policy.yaml

## --------------------------------------
##@ Release
## --------------------------------------

RELEASE_TAG := $(shell git describe --abbrev=0 2>/dev/null)
RELEASE_DIR := out

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)/

.PHONY: release
release: clean-release  ## Builds and push container images using the latest git tag for the commit.
	@if [ -z "${RELEASE_TAG}" ]; then echo "RELEASE_TAG is not set"; exit 1; fi
	@if ! [ -z "$$(git status --porcelain)" ]; then echo "Your local git repository contains uncommitted changes, use git clean before proceeding."; exit 1; fi
	git checkout "${RELEASE_TAG}"
	# Set the manifest image to the production bucket.
	$(MAKE) set-manifest-image MANIFEST_IMG=$(PROD_REGISTRY)/$(IMAGE_NAME) MANIFEST_TAG=$(RELEASE_TAG)
	$(MAKE) set-manifest-pull-policy PULL_POLICY=IfNotPresent
	$(MAKE) release-manifests
	$(MAKE) release-templates

.PHONY: release-manifests
release-manifests: $(KUSTOMIZE) $(RELEASE_DIR) ## Builds the manifests to publish with a release
	cp metadata.yaml $(RELEASE_DIR)/metadata.yaml
	kustomize build config > $(RELEASE_DIR)/infrastructure-components.yaml

.PHONY: release-templates
release-templates: $(RELEASE_DIR)
	cp templates/cluster-template* $(RELEASE_DIR)/

.PHONY: release-binary
release-binary: $(RELEASE_DIR)
	docker run \
		--rm \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-v "$$(pwd):/workspace" \
		-w /workspace \
		golang:1.13.8 \
		go build -a -ldflags '-extldflags "-static"' \
		-o $(RELEASE_DIR)/$(notdir $(RELEASE_BINARY))-$(GOOS)-$(GOARCH) $(RELEASE_BINARY)

.PHONY: release-staging
release-staging: ## Builds and push container images to the staging bucket.
	REGISTRY=$(STAGING_REGISTRY) $(MAKE) docker-build-all docker-push-all release-alias-tag

RELEASE_ALIAS_TAG=$(PULL_BASE_REF)

.PHONY: release-alias-tag
release-alias-tag: # Adds the tag to the last build tag.
	gcloud container images add-tag $(CONTROLLER_IMG):$(TAG) $(CONTROLLER_IMG):$(RELEASE_ALIAS_TAG)

.PHONY: release-notes
release-notes: $(RELEASE_NOTES)
	$(RELEASE_NOTES)

## --------------------------------------
##@ Development
## --------------------------------------

.PHONY: create-cluster
create-cluster: $(CLUSTERCTL) ## Create a development Kubernetes cluster on DigitalOcean using examples
	$(CLUSTERCTL) \
	create cluster -v 4 \
	--bootstrap-flags="name=clusterapi" \
	--bootstrap-type kind \
	-m ./examples/_out/controlplane.yaml \
	-c ./examples/_out/cluster.yaml \
	-p ./examples/_out/provider-components.yaml \
	-a ./examples/addons.yaml

# This is used in the get-kubeconfig call below in the create-cluster-management target. It may be overridden by the
# e2e-conformance.sh script, which is why we need it as a variable here.
CLUSTER_NAME ?= test1

.PHONY: create-cluster-management
create-cluster-management: $(CLUSTERCTL) ## Create a development Kubernetes cluster on DigitalOcean in a KIND management cluster.
	## Create kind management cluster.
	$(MAKE) kind-create

	@if [ ! -z "${LOAD_IMAGE}" ]; then \
		echo "loading ${LOAD_IMAGE} into kind cluster ..." && \
		kind --name="clusterapi" load docker-image "${LOAD_IMAGE}"; \
	fi
	# Apply core-components.
	kubectl \
		--kubeconfig=$$(kind get kubeconfig-path --name="clusterapi") \
		create -f examples/_out/core-components.yaml
	# Apply provider-components.
	kubectl \
		--kubeconfig=$$(kind get kubeconfig-path --name="clusterapi") \
		create -f examples/_out/provider-components.yaml
	# Create Cluster.
	kubectl \
		--kubeconfig=$$(kind get kubeconfig-path --name="clusterapi") \
		create -f examples/_out/cluster.yaml
	# Create control plane machine.
	kubectl \
		--kubeconfig=$$(kind get kubeconfig-path --name="clusterapi") \
		create -f examples/_out/controlplane.yaml
	# Get KubeConfig using clusterctl.
	$(CLUSTERCTL) \
		alpha phases get-kubeconfig -v=3 \
		--kubeconfig=$$(kind get kubeconfig-path --name="clusterapi") \
		--namespace=default \
		--cluster-name=$(CLUSTER_NAME)
	# Apply addons on the target cluster, waiting for the control-plane to become available.
	$(CLUSTERCTL) \
		alpha phases apply-addons -v=3 \
		--kubeconfig=./kubeconfig \
		-a examples/addons.yaml
	# Create a worker node with MachineDeployment.
	kubectl \
		--kubeconfig=$$(kind get kubeconfig-path --name="clusterapi") \
		create -f examples/_out/machinedeployment.yaml


.PHONY: delete-workload-cluster
delete-workload-cluster: $(CLUSTERCTL) ## Deletes the development Kubernetes Cluster "$CLUSTER_NAME"
	@echo 'Your DO resources will now be deleted, this can take some minutes'
	$(CLUSTERCTL) \
	delete cluster -v 4 \
	--bootstrap-type kind \
	--bootstrap-flags="name=clusterapi" \
	--cluster $(CLUSTER_NAME) \
	--kubeconfig ./kubeconfig \
	-p ./examples/_out/provider-components.yaml \

.PHONY: kind-create
kind-create: ## create capdo kind cluster if needed
	./scripts/kind-with-registry.sh

.PHONY: delete-cluster
delete-cluster: delete-workload-cluster  ## Deletes the example kind cluster "capdo"
	kind delete cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: kind-reset
kind-reset: ## Destroys the "capdo" kind cluster.
	kind delete cluster --name=$(KIND_CLUSTER_NAME) || true

## --------------------------------------
##@ Verification
## --------------------------------------

.PHONY: verify ## Verify all
verify: verify-boilerplate verify-modules verify-gen

.PHONY: verify-boilerplate
verify-boilerplate: ## Verify boilerplate
	./hack/verify-boilerplate.sh

.PHONY: verify-modules
verify-modules: modules ## Verify go module files
	@if !(git diff --quiet HEAD -- go.sum go.mod hack/tools/go.mod hack/tools/go.sum); then \
		echo "go module files are out of date"; exit 1; \
	fi

verify-gen: generate ## Verify generated files
	@if !(git diff --quiet HEAD); then \
		echo "generated files are out of date, run make generate"; exit 1; \
	fi

## --------------------------------------
##@ Misc
## --------------------------------------

.PHONY: vet
vet: ## Runs the Go vet command on this project
	go vet ./...

## --------------------------------------
##@ Cleanup
## --------------------------------------

.PHONY: clean
clean: ## Remove all generated files
	$(MAKE) clean-bin
	$(MAKE) clean-temporary

.PHONY: clean-bin
clean-bin: ## Remove all generated binaries
	rm -rf bin
	rm -rf hack/tools/bin

.PHONY: clean-temporary
clean-temporary: ## Remove all temporary files and folders
	rm -f minikube.kubeconfig
	rm -f kubeconfig

.PHONY: clean-release
clean-release: ## Remove the release folder
	rm -rf $(RELEASE_DIR)

.PHONY: clean-examples
clean-examples: ## Remove all the temporary files generated in the examples folder
	rm -rf examples/_out/
	rm -f examples/core-components/*-components.yaml
	rm -f examples/provider-components/provider-components-*.yaml
