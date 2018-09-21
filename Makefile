# Copyright 2018 The Kubernetes Authors.
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

all: generate build images

check: depend fmt vet gometalinter

depend: ## Sync vendor directory by running dep ensure
	dep version || go get -u github.com/golang/dep/cmd/dep
	dep ensure -v

depend-update: ## Update all dependencies
	dep ensure -update -v

.PHONY: generate
generate:
	go build -o $$GOPATH/bin/deepcopy-gen github.com/kubermatic/cluster-api-provider-digitalocean/vendor/k8s.io/code-generator/cmd/deepcopy-gen
	deepcopy-gen \
	  -i ./cloud/digitalocean/providerconfig,./cloud/digitalocean/providerconfig/v1alpha1 \
	  -O zz_generated.deepcopy \
	  -h boilerplate.go.txt

compile: ## Compile project and create binaries for cluster-controller, machine-controller and clusterctl, in the ./bin directory
	mkdir -p ./bin
	go build -o ./bin/cluster-controller ./cmd/cluster-controller
	go build -o ./bin/machine-controller ./cmd/machine-controller
	go build -o ./bin/clusterctl ./clusterctl

install: ## Install cluster-controller, machine-controller and clusterctl
	CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' github.com/kubermatic/cluster-api-provider-digitalocean/cmd/cluster-controller
	CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' github.com/kubermatic/cluster-api-provider-digitalocean/cmd/machine-controller
	CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' github.com/kubermatic/cluster-api-provider-digitalocean/clusterctl

test-unit: ## Run unit tests. Those tests will never communicate with cloud and cost you money
	go test -race -cover ./cmd/... ./cloud/...

clean: ## Remove compiled binaries
	rm -rf ./bin

images: ## Build images for cluster-controller and machine-controller
	$(MAKE) -C cmd/cluster-controller image
	$(MAKE) -C cmd/machine-controller image

push: ## Build and push images to repository for cluster-controller and machine-controller
	$(MAKE) -C cmd/cluster-controller push
	$(MAKE) -C cmd/machine-controller push

images-dev: ## Build development images for cluster-controller and machine-controller
	$(MAKE) -C cmd/cluster-controller image-dev
	$(MAKE) -C cmd/machine-controller image-dev

push-dev: ## Build and push development images to repository for cluster-controller and machine-controller
	$(MAKE) -C cmd/cluster-controller push-dev
	$(MAKE) -C cmd/machine-controller push-dev

gofmt: ## Go fmt your code
	hack/verify-gofmt.sh

vet: ## Apply go vet to all go files
	go vet ./...

gometalinter: ## Run gometalinter on all go files
	gometalinter --version || go get -u gopkg.in/alecthomas/gometalinter.v2
	gometalinter --install
	gometalinter --config gometalinter.json ./...

.PHONY: help
help:  ## Show help messages for make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'

