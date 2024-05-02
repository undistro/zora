# Image URL to use all building/pushing image targets
IMG ?= controller:latest
WORKER_IMG ?= worker:latest

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.3

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

NAMESPACE ?= zora-system
.PHONY: manifests
manifests: controller-gen addlicense yq ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	cd config/chart-crd && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build config/chart-crd | $(YQ) -s '"charts/zora/crds/" + .spec.group + "_" + .spec.names.plural + ".yaml"'
	$(ADDLICENSE) -c "Undistro Authors" -l "apache" -ignore ".github/**" -ignore ".idea/**" -ignore "dist/**" -ignore "site/**" -ignore "config/**" -ignore "docs/overrides/**" -ignore "docs/stylesheets/**" .

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: generate-client
generate-client:  ## Generate client
	@rm -r pkg/clientset pkg/informers pkg/listers || echo -n
	$(CONTAINER_TOOL) run -i --rm \
		-v $(PWD):/app \
		ghcr.io/slok/kube-code-generator:v0.1.0 \
		--apis-in ./api \
		--go-gen-out ./pkg \
		--debug

.PHONY: generate-helm-docs
generate-helm-docs: helm-docs ## Generate documentation for helm chart.
	$(HELM_DOCS) -s=file --badge-style="flat-square&color=3CA9DD"

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: check-license
check-license: ## Check license headers.
	$(ADDLICENSE) -c "Undistro Authors" -l "apache" -ignore ".github/**" -ignore ".idea/**" -ignore "dist/**" -ignore "site/**" -ignore "config/**" -ignore "docs/overrides/**" -ignore "docs/stylesheets/**" -check .

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager and worker binaries.
	go build -o bin/manager cmd/main.go
	go build -o bin/worker cmd/worker/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} -f cmd/Dockerfile .

.PHONY: docker-build-worker
docker-build-worker: test ## Build docker image with worker.
	$(CONTAINER_TOOL) build -t ${WORKER_IMG} -f cmd/worker/Dockerfile .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

.PHONY: docker-push-worker
docker-push-worker: ## Push docker image with worker.
	$(CONTAINER_TOOL) push ${WORKER_IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/myoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support.
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' cmd/Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install-crds
install-crds: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: install
install: install-crds ## Install CRs (plugins, custom checks, cluster, and scan) into the K8s cluster specified in ~/.kube/config.
	@$(KUBECTL) create namespace $(NAMESPACE) || true
	@$(KUBECTL) apply -f config/samples/zora_v1alpha1_plugin_popeye_all.yaml -n $(NAMESPACE)
	@$(KUBECTL) apply -f config/samples/zora_v1alpha1_plugin_marvin.yaml -n $(NAMESPACE)
	@$(KUBECTL) apply -f config/samples/zora_v1alpha1_plugin_trivy.yaml -n $(NAMESPACE)
	@$(KUBECTL) apply -f config/samples/zora_v1alpha1_customcheck_replicas.yaml -n $(NAMESPACE)
	@$(KUBECTL) apply -f config/samples/zora_v1alpha2_customcheck_labels.yaml -n $(NAMESPACE)
	@$(KUBECTL) apply -f config/rbac/zora_plugins_role.yaml
	@$(KUBECTL) create -f config/rbac/zora_plugins_role_binding.yaml || true
	@$(KUBECTL) apply -f config/samples/zora_v1alpha1_cluster.yaml -n $(NAMESPACE)
	@$(KUBECTL) apply -f config/samples/zora_v1alpha1_clusterscan_misconfig.yaml -n $(NAMESPACE)
	@$(KUBECTL) apply -f config/samples/zora_v1alpha1_clusterscan_vuln.yaml -n $(NAMESPACE)

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: template
template: manifests kustomize ## Build kustomize configurations.
	$(KUSTOMIZE) build config/default

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Kind

CLUSTER_NAME ?= kind
.PHONY: kind-create-cluster
kind-create-cluster: kind ## Create a local Kubernetes cluster with Kind
	$(KIND) create cluster --name $(CLUSTER_NAME)

.PHONY: kind-load-images
kind-load-images: kind docker-build docker-build-worker ## Build and load docker images into Kind nodes
	$(KIND) load docker-image ${IMG}
	$(KIND) load docker-image ${WORKER_IMG}

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize-$(KUSTOMIZE_VERSION)
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
ENVTEST ?= $(LOCALBIN)/setup-envtest-$(ENVTEST_VERSION)
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
ADDLICENSE ?= $(LOCALBIN)/addlicense-$(ADDLICENSE_VERSION)
HELM_DOCS ?= $(LOCALBIN)/helm-docs-$(HELM_DOCS_VERSION)
KIND ?= $(LOCALBIN)/kind-$(KIND_VERSION)
YQ ?= $(LOCALBIN)/yq-$(YQ_VERSION)

## Tool Versions
KUSTOMIZE_VERSION ?= v5.3.0
CONTROLLER_TOOLS_VERSION ?= v0.14.0
ENVTEST_VERSION ?= release-0.17
GOLANGCI_LINT_VERSION ?= v1.54.2
HELM_DOCS_VERSION ?= v1.13.1
ADDLICENSE_VERSION ?= v1.1.1
KIND_VERSION ?= v0.22.0
YQ_VERSION ?= v4.43.1

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

.PHONY: addlicense
addlicense: $(ADDLICENSE) ## Download addlicense locally if necessary
$(ADDLICENSE): $(LOCALBIN)
	$(call go-install-tool,$(ADDLICENSE),github.com/google/addlicense,${ADDLICENSE_VERSION})

.PHONY: helm-docs
helm-docs: $(HELM_DOCS) ## Download helm-docs locally if necessary
$(HELM_DOCS): $(LOCALBIN)
	$(call go-install-tool,$(HELM_DOCS),github.com/norwoodj/helm-docs/cmd/helm-docs,${HELM_DOCS_VERSION})

.PHONY: kind
kind: $(KIND) ## Download kind locally if necessary
$(KIND): $(LOCALBIN)
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind,${KIND_VERSION})

.PHONY: yq
yq: $(YQ) ## Download yq locally if necessary
$(YQ): $(LOCALBIN)
	$(call go-install-tool,$(YQ),github.com/mikefarah/yq/v4,${YQ_VERSION})

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef
