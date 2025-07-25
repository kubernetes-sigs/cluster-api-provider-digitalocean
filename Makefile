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

GOPATH  := $(shell go env GOPATH)
GOARCH  := $(shell go env GOARCH)
GOOS    := $(shell go env GOOS)
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
BIN_DIR := $(abspath $(ROOT_DIR)/bin)
TEST_E2E_DIR := test/e2e
REPO_ROOT := $(shell git rev-parse --show-toplevel)
ARTIFACTS ?= $(REPO_ROOT)/_artifacts
E2E_DATA_DIR ?= $(ROOT_DIR)/$(TEST_E2E_DIR)/data
KUBETEST_CONF_PATH ?= $(abspath $(E2E_DATA_DIR)/kubetest/conformance.yaml)
GO_INSTALL = ./scripts/go_install.sh

# curl retries
CURL_RETRIES=3

# Files
E2E_CONF_FILE ?= $(REPO_ROOT)/test/e2e/config/digitalocean-dev.yaml
E2E_CONF_FILE_ENVSUBST := $(ROOT_DIR)/test/e2e/config/digitalocean-dev-envsubst.yaml

# Define Docker related variables. Releases should modify and double check these vars.
export GCP_PROJECT ?= $(shell gcloud config get-value project)
REGISTRY ?= gcr.io/$(GCP_PROJECT)
STAGING_REGISTRY := gcr.io/k8s-staging-cluster-api-do
PROD_REGISTRY := registry.k8s.io/cluster-api-do
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
GINKGO_SKIP ?= API Version Upgrade
GINKGO_NODES ?= 3
GINKGO_NOCOLOR ?= false
GINKGO_ARGS ?=
GINKGO_TIMEOUT ?= 2h
GINKGO_POLL_PROGRESS_AFTER ?= 10m
GINKGO_POLL_PROGRESS_INTERVAL ?= 1m

# Allow overriding the imagePullPolicy
PULL_POLICY ?= Always

# Docker buildkit disabled for now until we figure out how to fix when
# building the image in the post stage
DOCKER_BUILDKIT ?= 0
export DOCKER_BUILDKIT

# Hosts running SELinux need :z added to volume mounts
SELINUX_ENABLED := $(shell cat /sys/fs/selinux/enforce 2> /dev/null || echo 0)

ifeq ($(SELINUX_ENABLED),1)
  DOCKER_VOL_OPTS?=:z
endif

KIND_CLUSTER_NAME ?= capdo

# set --output-base used for conversion-gen which needs to be different for in GOPATH and outside GOPATH dev
ifneq ($(abspath $(ROOT_DIR)),$(GOPATH)/src/sigs.k8s.io/cluster-api-provider-digitalocean)
  OUTPUT_BASE := --output-base=$(ROOT_DIR)
endif

# CI
CAPDO_WORKER_CLUSTER_KUBECONFIG ?= "/tmp/kubeconfig"

# Build time versioning details.
LDFLAGS := $(shell hack/version.sh)

# Other tools versions
CERT_MANAGER_VER := v1.14.4

## --------------------------------------
##@ Help
## --------------------------------------

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
##@ Binaries
## --------------------------------------

# Binaries.
CONTROLLER_GEN_VER := v0.17.1
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER)

CONVERSION_GEN_VER := v0.31.10
CONVERSION_GEN_BIN := conversion-gen
CONVERSION_GEN := $(TOOLS_BIN_DIR)/$(CONVERSION_GEN_BIN)-$(CONVERSION_GEN_VER)

ENVSUBST_VER := v1.4.3
ENVSUBST_BIN := envsubst
ENVSUBST := $(TOOLS_BIN_DIR)/$(ENVSUBST_BIN)

GOLANGCI_LINT_VER := v2.1.6
GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER)

KUSTOMIZE_VER := v4.5.7
KUSTOMIZE_BIN := kustomize
KUSTOMIZE := $(TOOLS_BIN_DIR)/$(KUSTOMIZE_BIN)-$(KUSTOMIZE_VER)

MOCKGEN_VER := v0.5.2
MOCKGEN_BIN := mockgen
MOCKGEN := $(TOOLS_BIN_DIR)/$(MOCKGEN_BIN)-$(MOCKGEN_VER)

RELEASE_NOTES_VER := v0.11.0
RELEASE_NOTES_BIN := release-notes
RELEASE_NOTES := $(TOOLS_BIN_DIR)/$(RELEASE_NOTES_BIN)-$(RELEASE_NOTES_VER)

GINKGO_VER := v2.23.4
GINKGO_BIN := ginkgo
GINKGO := $(TOOLS_BIN_DIR)/$(GINKGO_BIN)-$(GINKGO_VER)
GINKGO_PKG := github.com/onsi/ginkgo/v2/ginkgo

KUBECTL_VER := v1.31.3
KUBECTL_BIN := $(TOOLS_BIN_DIR)/kubectl
KUBECTL := $(KUBECTL_BIN)-$(KUBECTL_VER)

TIMEOUT := $(shell command -v timeout || command -v gtimeout)

## --------------------------------------
##@ Testing
## --------------------------------------

$(ARTIFACTS):
	mkdir -p $@

.PHONY: test
test: generate ## Run tests
	source ./scripts/fetch_ext_bins.sh; fetch_tools; setup_envs; go test -v -covermode=atomic -coverprofile=coverage.tmp.out ./api/... ./internal/controller/... ./cloud/...
	@cat coverage.tmp.out | grep -v "generated" > coverage.out
	@rm coverage.tmp.out

.PHONY: test-e2e ## Run e2e tests using clusterctl
test-e2e: e2e-image $(ENVSUBST) $(GINKGO) $(KIND) $(KUSTOMIZE)  ## Run e2e tests
	$(ENVSUBST) < $(E2E_CONF_FILE) > $(E2E_CONF_FILE_ENVSUBST) && \
	time $(GINKGO) -v --trace -poll-progress-after=$(GINKGO_POLL_PROGRESS_AFTER) -poll-progress-interval=$(GINKGO_POLL_PROGRESS_INTERVAL) \
	--tags=e2e --focus="$(GINKGO_FOCUS)" -skip="$(GINKGO_SKIP)" --nodes=$(GINKGO_NODES) --no-color=$(GINKGO_NOCOLOR) \
	--timeout=$(GINKGO_TIMEOUT) --output-dir="$(ARTIFACTS)" --junit-report="junit.e2e_suite.1.xml" $(GINKGO_ARGS) ./test/e2e -- \
			-e2e.config="$(E2E_CONF_FILE_ENVSUBST)" \
			-e2e.artifacts-folder="$(ARTIFACTS)" $(E2E_ARGS)

.PHONY: test-conformance
test-conformance: e2e-image $(ENVSUBST) $(GINKGO) $(KIND) $(KUSTOMIZE) ## Run conformance test on workload cluster
	$(ENVSUBST) < $(E2E_CONF_FILE) > $(E2E_CONF_FILE_ENVSUBST) && \
	time $(GINKGO) -v --trace -poll-progress-after=$(GINKGO_POLL_PROGRESS_AFTER) -poll-progress-interval=$(GINKGO_POLL_PROGRESS_INTERVAL) \
	--tags=e2e --focus="$(GINKGO_FOCUS)" -skip="$(GINKGO_SKIP)" --nodes=$(GINKGO_NODES) --no-color=$(GINKGO_NOCOLOR) \
	--timeout=$(GINKGO_TIMEOUT) --output-dir="$(ARTIFACTS)" --junit-report="junit.conformance_suite.1.xml" $(GINKGO_ARGS) ./test/e2e -- \
			-e2e.config="$(E2E_CONF_FILE_ENVSUBST)" \
			-kubetest.config-file=$(KUBETEST_CONF_PATH) \
			-e2e.artifacts-folder="$(ARTIFACTS)" $(E2E_ARGS)

.PHONY: e2e-image
e2e-image:
	docker build --build-arg ldflags="$(LDFLAGS)" --tag="gcr.io/k8s-staging-cluster-api-do/cluster-api-do-controller:e2e" .


.PHONY: binaries
binaries: manager ## Builds and installs all binaries

.PHONY: manager
manager: ## Build manager binary.
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/manager ./cmd

## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(CONTROLLER_GEN): ## Build controller-gen from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) sigs.k8s.io/controller-tools/cmd/controller-gen $(CONTROLLER_GEN_BIN) $(CONTROLLER_GEN_VER)

$(CONVERSION_GEN): ## Build conversion-gen.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) k8s.io/code-generator/cmd/conversion-gen $(CONVERSION_GEN_BIN) $(CONVERSION_GEN_VER)

$(ENVSUBST): ## Build envsubst from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/a8m/envsubst/cmd/envsubst $(ENVSUBST_BIN) $(ENVSUBST_VER)

$(GOLANGCI_LINT): ## Build golangci-lint from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint $(GOLANGCI_LINT_BIN) $(GOLANGCI_LINT_VER)

$(KUSTOMIZE): ## Build kustomize from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) sigs.k8s.io/kustomize/kustomize/v4 $(KUSTOMIZE_BIN) $(KUSTOMIZE_VER)

$(MOCKGEN): ## Build mockgen from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) go.uber.org/mock/mockgen $(MOCKGEN_BIN) $(MOCKGEN_VER)

$(RELEASE_NOTES): ## Build release notes.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) k8s.io/release/cmd/release-notes $(RELEASE_NOTES_BIN) $(RELEASE_NOTES_VER)

$(GINKGO): ## Build ginkgo.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(GINKGO_PKG) $(GINKGO_BIN) $(GINKGO_VER)

$(KUBECTL): ## Build kubectl
	mkdir -p $(TOOLS_BIN_DIR)
	rm -f "$(KUBECTL)*"
	curl --retry $(CURL_RETRIES) -fsL https://dl.k8s.io/release/$(KUBECTL_VER)/bin/$(GOOS)/$(GOARCH)/kubectl -o $(KUBECTL)
	ln -sf "$(KUBECTL)" "$(KUBECTL_BIN)"
	chmod +x "$(KUBECTL_BIN)" "$(KUBECTL)"

.PHONY: $(KUBECTL_BIN)
$(KUBECTL_BIN): $(KUBECTL) ## Building kubectl from the tools folder

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
generate-go: $(CONTROLLER_GEN) $(CONVERSION_GEN) $(MOCKGEN) ## Runs Go related generate targets
	$(CONTROLLER_GEN) \
		paths=./cmd \
		paths=./... \
		paths=./$(EXP_DIR)/api/... \
		object:headerFile=./hack/boilerplate/boilerplate.generatego.txt
	go generate ./...

.PHONY: generate-manifests
generate-manifests: $(CONTROLLER_GEN) ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./cmd \
		paths=./api/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=$(CRD_ROOT) \
		output:webhook:dir=$(WEBHOOK_ROOT) \
		webhook
	$(CONTROLLER_GEN) \
		paths=./cmd \
		paths=./internal/controller/... \
		output:rbac:dir=$(RBAC_ROOT) \
		rbac:roleName=manager-role

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: docker-pull-prerequisites
docker-pull-prerequisites:
	docker pull docker/dockerfile:1.1-experimental
	docker pull docker.io/library/golang:1.22.12
	docker pull gcr.io/distroless/static:latest

.PHONY: docker-build
docker-build: docker-pull-prerequisites ## Build the docker image for controller-manager
	DOCKER_BUILDKIT=1 docker build --build-arg goproxy=$(GOPROXY) --build-arg ARCH=$(ARCH) --build-arg ldflags="$(LDFLAGS)" . -t $(CONTROLLER_IMG)-$(ARCH):$(TAG)
	$(MAKE) set-manifest-image MANIFEST_IMG=$(CONTROLLER_IMG)-$(ARCH) MANIFEST_TAG=$(TAG) TARGET_RESOURCE="./config/default/manager_image_patch.yaml"
	$(MAKE) set-manifest-pull-policy TARGET_RESOURCE="./config/default/manager_pull_policy.yaml"

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(CONTROLLER_IMG)-$(ARCH):$(TAG)

## --------------------------------------
## Docker — All ARCH
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
	$(info Updating kustomize image patch file for default resource)
	sed -i'' -e 's@image: .*@image: '"${MANIFEST_IMG}:$(MANIFEST_TAG)"'@' ./config/default/manager_image_patch.yaml

.PHONY: set-manifest-pull-policy
set-manifest-pull-policy:
	$(info Updating kustomize pull policy file for default resource)
	sed -i'' -e 's@imagePullPolicy: .*@imagePullPolicy: '"$(PULL_POLICY)"'@' ./config/default/manager_pull_policy.yaml

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
	kustomize build config/default > $(RELEASE_DIR)/infrastructure-components.yaml

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
		golang:1.20.8 \
		go build -a -trimpath -ldflags '-extldflags "-static"' \
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

.PHONY: create-management-cluster
create-management-cluster: $(KUSTOMIZE) $(ENVSUBST)
	## Create kind management cluster.
	$(MAKE) kind-create

	# Install cert manager and wait for availability
	./hack/install-cert-manager.sh $(CERT_MANAGER_VER)

	# Deploy CAPI
	curl --retry $(CURL_RETRIES) -sSL https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.5.2/cluster-api-components.yaml | $(ENVSUBST) | kubectl apply -f -

	# Deploy CAPDO
	kind load docker-image $(CONTROLLER_IMG)-$(ARCH):$(TAG) --name=capdo
	$(KUSTOMIZE) build config/default | $(ENVSUBST) | kubectl apply -f -

	# Wait for CAPI pods
	kubectl wait --for=condition=Available --timeout=5m -n capi-system deployment -l cluster.x-k8s.io/provider=cluster-api
	kubectl wait --for=condition=Available --timeout=5m -n capi-kubeadm-bootstrap-system deployment -l cluster.x-k8s.io/provider=bootstrap-kubeadm
	kubectl wait --for=condition=Available --timeout=5m -n capi-kubeadm-control-plane-system deployment -l cluster.x-k8s.io/provider=control-plane-kubeadm

	# Wait for CAPDO pods
	kubectl wait --for=condition=Ready --timeout=5m -n capdo-system pod -l cluster.x-k8s.io/provider=infrastructure-digitalocean

	# required sleep for when creating management and workload cluster simultaneously
	sleep 10
	@echo 'Set kubectl context to the kind management cluster by running "kubectl config set-context kind-clusterapi"'

.PHONY: create-workload-cluster
create-workload-cluster: $(KUSTOMIZE) $(ENVSUBST)
	# Create workload Cluster.
	$(KUSTOMIZE) build templates | $(ENVSUBST) | kubectl apply -f -

	# Wait for the kubeconfig to become available.
	${TIMEOUT} 5m bash -c "while ! kubectl get secrets | grep $(CLUSTER_NAME)-kubeconfig; do sleep 1; done"
	# Get kubeconfig and store it locally.
	kubectl get secrets $(CLUSTER_NAME)-kubeconfig -o json | jq -r .data.value | base64 --decode > $(CAPDO_WORKER_CLUSTER_KUBECONFIG)
	${TIMEOUT} 15m bash -c "while ! kubectl --kubeconfig=$(CAPDO_WORKER_CLUSTER_KUBECONFIG) get nodes | grep master; do sleep 1; done"

	# Deploy calico
	kubectl --kubeconfig=$(CAPDO_WORKER_CLUSTER_KUBECONFIG) apply -f https://docs.projectcalico.org/manifests/calico.yaml

	@echo 'run "kubectl --kubeconfig=$(CAPDO_WORKER_CLUSTER_KUBECONFIG) ..." to work with the new target cluster'

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

## --------------------------------------
## Tilt / Kind
## --------------------------------------

.PHONY: kind-create
kind-create: $(KUBECTL) ## create capg kind cluster if needed
	./scripts/kind-with-registry.sh

.PHONY: kind-delete
kind-delete: ## Destroys the "capdo" kind cluster.
	kind delete cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: tilt-up
tilt-up: $(ENVSUBST) $(KUSTOMIZE) $(KUBECTL) kind-create ## start tilt and build kind cluster if needed
	EXP_CLUSTER_RESOURCE_SET=true tilt up

.PHONY: delete-cluster
delete-cluster: delete-workload-cluster  ## Deletes the example kind cluster "capdo"
	kind delete cluster --name=$(KIND_CLUSTER_NAME)

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

.PHONY: do-janitor
do-janitor: ## Cleanup old resources in the DO account
	go run hack/do-janitor/do-janitor.go
