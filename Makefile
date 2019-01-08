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

QUAY_BUCKET = kubermatic
PREFIX = quay.io/$(QUAY_BUCKET)
NAME = cluster-api-do-controller
TAG = v1.0.0-alpha.1
DEV_TAG = v1.0.0-alpha.1

all: depend generate compile images

check: depend gofmt vet gometalinter

depend: ## Sync vendor directory by running dep ensure
	$$GOPATH/bin/dep version || go get -u github.com/golang/dep/cmd/dep
	$$GOPATH/bin/dep ensure -v

depend-update: ## Update all dependencies
	$$GOPATH/bin/dep version || go get -u github.com/golang/dep/cmd/dep
	$$GOPATH/bin/dep ensure -update -v

.PHONY: generate
generate:
	GOPATH=${GOPATH} go generate ./pkg/... ./cmd/...

compile: ## Compile project and create binaries for manager and clusterctl, in the ./bin directory
	mkdir -p ./bin
	go build -o ./bin/manager ./cmd/manager
	go build -o ./bin/clusterctl ./cmd/clusterctl

install: ## Install manager and clusterctl
	CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' sigs.k8s.io/cluster-api-provider-digitalocean/cmd/manager
	CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' sigs.k8s.io/cluster-api-provider-digitalocean/cmd/clusterctl

test-unit: ## Run unit tests. Those tests will never communicate with cloud and cost you money
	go test -race -cover ./...

clean: ## Remove compiled binaries
	rm -rf ./bin

images: ## Build image for manager
	docker build -t "$(PREFIX)/$(NAME):$(TAG)" -f ./Dockerfile .

push: images ## Build and push image to repository for manager
	docker push "$(PREFIX)/$(NAME):$(TAG)"

images-dev: ## Build development image for manager
	docker build -t "$(PREFIX)/$(NAME):$(DEV_TAG)" -f ./Dockerfile .

push-dev: images-dev ## Build and push development image to repository for manager
	docker push "$(PREFIX)/$(NAME):$(DEV_TAG)"

gofmt: ## Go fmt your code
	hack/verify-gofmt.sh

vet: ## Apply go vet to all go files
	go vet ./...

lint: ## Run gometalinter on all go files
	gometalinter --config gometalinter.json ./... --deadline 20m

.PHONY: help
help:  ## Show help messages for make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'

verify: depend vet gofmt
	hack/verify-boilerplate.sh

