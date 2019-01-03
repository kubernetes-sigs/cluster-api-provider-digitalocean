# Image URL to use all building/pushing image targets
IMG ?= xmudrii/do-cluster-api-controller:latest

all: test manager clusterctl

# Run tests
test: generate fmt vet manifests
	go test -v -tags=integration ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager sigs.k8s.io/cluster-api-provider-digitalocean/cmd/manager

# Build manager binary
clusterctl: generate fmt vet
	go build -o bin/clusterctl sigs.k8s.io/cluster-api-provider-digitalocean/cmd/clusterctl

.PHONY: generate
generate:
	GOPATH=${GOPATH} go generate ./pkg/... ./cmd/...

compile: ## Compile project and create binaries for cluster-controller, machine-controller and clusterctl, in the ./bin directory
	mkdir -p ./bin
	# TODO: manager here
	go build -o ./bin/clusterctl ./cmd/clusterctl

install: ## Install cluster-controller, machine-controller and clusterctl
	# TODO: manager here
	CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' sigs.k8s.io/cluster-api-provider-digitalocean/cmd/clusterctl

test-unit: ## Run unit tests. Those tests will never communicate with cloud and cost you money
	go test -race -cover ./...

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build: generate fmt vet manifests
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/do_manager_image_patch.yaml

verify: depend vet gofmt
	hack/verify-boilerplate.sh

