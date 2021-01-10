IMAGE ?= controller:latest
PUSH_IMAGE ?= false
PLATFORM ?= linux/amd64,linux/arm64

CRD_OPTIONS ?= "crd:trivialVersions=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	DEV_MODE=true HOSTNAME=local NAMESPACE=k3os-config-operator-system go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
dev: manifests tools
	PLATFORM=linux/arm64 $(SKAFFOLD) run -p dev --tail

dev-delete: tools
	@ $(SKAFFOLD) delete -p dev

deploy: manifests tools
	$(SKAFFOLD) run -p release

# This is used to update the manifests into deploy/operator.yaml
render-static-manifests:
	@ $(SKAFFOLD) build -q -p release
	@ $(KUSTOMIZE) build config/release > deploy/operator.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the Docker image
docker-build: test
	docker buildx build . -t ${IMAGE} --platform ${PLATFORM} --push=${PUSH_IMAGE}

# Build dev Docker image
docker-build-dev:
	docker buildx build . -t ${IMAGE} --platform ${PLATFORM} -f Dockerfile.dev --push=${PUSH_IMAGE}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

skaffold:
ifeq (, $(shell which skaffold))
	@{ \
	set -e ;\
	SKAFFOLD_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$SKAFFOLD_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get github.com/GoogleContainerTools/skaffold/cmd/skaffold@v1.17.2 ;\
	rm -rf $$SKAFFOLD_GEN_TMP_DIR ;\
	}
SKAFFOLD=$(GOBIN)/skaffold
else
SKAFFOLD=$(shell which skaffold)
endif

tools: kustomize skaffold
