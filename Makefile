include hack/make/*

##@ Tooling Download

controller-gen:  ## Download controller-gen locally if necessary.
	$(call go-install-tool,${CONTROLLER_GEN},sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)
kustomize:  ## Download kustomize locally if necessary.
	$(call go-install-tool,${KUSTOMIZE},sigs.k8s.io/kustomize/kustomize/v4@v4.5.2)
envtest:  ## Download envtest-setup locally if necessary.
	$(call go-install-tool,${ENVTEST},sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)
addlicense:  ## Download addlicense locally if necessary.
	$(call go-install-tool,${ADDLICENSE},github.com/google/addlicense@latest)


##@ Development

fmt:  ## Run go fmt against code.
	go fmt ./...
vet:  ## Run go vet against code.
	go vet ./...
test: manifests generate fmt vet envtest  ## Run tests.
	KUBEBUILDER_ASSETS="$(shell ${ENVTEST} use ${ENVTEST_K8S_VERSION} -p path)" go test ./... -coverprofile cover.out

charts/zora/templates/operator/rbac.yaml: config/rbac/service_account.yaml \
 config/rbac/leader_election_role.yaml \
 config/rbac/role.yaml \
 config/rbac/auth_proxy_client_clusterrole.yaml \
 config/rbac/auth_proxy_role.yaml \
 config/rbac/leader_election_role_binding.yaml \
 config/rbac/role_binding.yaml \
 config/rbac/auth_proxy_role_binding.yaml
	@ rm $@
	@ for f in $^; do \
		patch -Nfi "hack/patches/rbac/$$(basename -s '.yaml' $$f).patch" \
			--no-backup-if-mismatch \
			-p 1 -o - | sed '/#/{N; d}' >> $@; \
		echo "---" >> $@; \
	done

manifest-consitency: charts/zora/templates/operator/rbac.yaml

manifests: controller-gen  ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	${CONTROLLER_GEN} rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	@cp -r config/crd/bases/*.yaml charts/zora/crds/
	${MAKE} manifest-consitency license

hack/scripts/gen_zora_view_kubeconfig.sh docs/targetcluster.sh: hack/scripts/m4/*
	@ m4 -I hack/scripts/m4 $(shell basename $@.m4) > $@

script-consitency: hack/scripts/gen_zora_view_kubeconfig.sh docs/targetcluster.sh

generate: controller-gen script-consitency license  ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	${CONTROLLER_GEN} object:headerFile="hack/boilerplate.go.txt" paths="./..."

clientset-gen:  ## Generate clientset
	@rm -r pkg/clientset || echo -n
	@docker run -i --rm \
		-v ${PWD}:/go/src/${PROJECT_PACKAGE} \
		-e PROJECT_PACKAGE=${PROJECT_PACKAGE} \
		-e CLIENT_GENERATOR_OUT=${PROJECT_PACKAGE}/pkg \
		-e APIS_ROOT=${PROJECT_PACKAGE}/apis \
		-e GROUPS_VERSION="zora:v1alpha1" \
		-e GENERATION_TARGETS="client" \
		-e BOILERPLATE_PATH="hack/boilerplate.go.txt" \
		quay.io/slok/kube-code-generator:v1.23.0


##@ Build and Execution

build: generate fmt vet  ## Build manager binary.
	go build -o bin/manager main.go
	go build -o bin/worker worker/main.go

run: install manifests generate  ## Run a controller from your host.
	go run ./main.go -default-plugins-names ${PLUGINS} -worker-image ${WORKER_IMG}

docker-build: test  ## Build manager docker image.
	docker build -t ${IMG} -f ${DOCKERFILE} .
docker-build-worker: docker-build  ## Build Docker images for all components.
	${MAKE} IMG=${WORKER_IMG} DOCKERFILE=Dockerfile.worker docker-build


##@ Deployment

install: manifests kustomize  ## Install default configuration (RBAC for plugins) and CRDs on the current cluster.
	${KUSTOMIZE} build config/crd | kubectl apply -f -
	@kubectl apply -f config/rbac/clusterissue_editor_role.yaml
	@kubectl apply -f config/samples/zora_v1alpha1_plugin_popeye.yaml
	@kubectl create -f config/rbac/plugins_role_binding.yaml || true

uninstall: manifests kustomize  ## Uninstall CRDs from the current cluster.
	${KUSTOMIZE} build config/crd | kubectl delete --ignore-not-found=${IGNORE_NOT_FOUND} -f -

deploy: docker-build generate install  ## Deploy controller on the current cluster.
	cd config/manager && ${KUSTOMIZE} edit set image controller=${IMG}
	${KUSTOMIZE} build config/default | kubectl apply -f -
undeploy:  ## Undeploy controller from current cluster.
	${KUSTOMIZE} build config/default | kubectl delete --ignore-not-found=${IGNORE_NOT_FOUND} -f -

gen-zora-view-kubeconfig:  ## Create a service account and config RBAC for it.
	./hack/scripts/gen_zora_view_kubeconfig.sh
setup-zora-view: install  ## Create and apply view secret.
	./hack/scripts/setup_zora_view.sh

##@ Documentation

helm-docs:  ## Generate documentation for helm charts
	@docker run -it --rm \
		-v ${PWD}:/helm-docs \
		jnorwood/helm-docs:v1.8.1 \
		helm-docs -s=file --badge-style="flat-square&color=3CA9DD"

license: addlicense  ## Add license header to source files
	$(call addlicense-tool,)

check-license: addlicense  ## Check license header to source files
	$(call addlicense-tool,"-check")
