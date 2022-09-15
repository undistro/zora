include hack/make/*

##@ Tooling Download

controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)
kustomize: ## Download kustomize locally if necessary.
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.2)
envtest: ## Download envtest-setup locally if necessary.
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)
addlicense: ## Download addlicense locally if necessary.
	$(call go-install-tool,$(ADDLICENSE),github.com/google/addlicense@latest)

##@ Development

fmt: ## Run go fmt against code.
	go fmt ./...
vet: ## Run go vet against code.
	go vet ./...
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

charts/zora/templates/plugins/popeye.yaml: config/samples/zora_v1alpha1_plugin_popeye.yaml
	@ cp $< $@
	patch -Nf --no-backup-if-mismatch $@ hack/patches/popeye_plugin.patch
charts/zora/templates/plugins/kubescape.yaml: config/samples/zora_v1alpha1_plugin_kubescape.yaml
	@ cp $< $@
	patch -Nf --no-backup-if-mismatch $@ hack/patches/kubescape_plugin.patch

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
			-p 1 -o - >> $@; \
		echo "---" >> $@; \
	done

manifest-consitency: charts/zora/templates/operator/rbac.yaml \
 charts/zora/templates/plugins/popeye.yaml \
 charts/zora/templates/plugins/kubescape.yaml

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	@cp -r config/crd/bases/*.yaml charts/zora/crds/
	$(MAKE) manifest-consitency

generate: controller-gen ## Generate clientset and code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

clientset-gen: ## Generate clientset
	@rm -r pkg/clientset || echo -n
	@docker run -i --rm \
		-v $(PWD):/go/src/$(PROJECT_PACKAGE) \
		-e PROJECT_PACKAGE=$(PROJECT_PACKAGE) \
		-e CLIENT_GENERATOR_OUT=$(PROJECT_PACKAGE)/pkg \
		-e APIS_ROOT=$(PROJECT_PACKAGE)/apis \
		-e GROUPS_VERSION="zora:v1alpha1" \
		-e GENERATION_TARGETS="client" \
		-e BOILERPLATE_PATH="hack/boilerplate.go.txt" \
		registry.undistro.io/quay/slok/kube-code-generator:v1.23.0


##@ Build and Execution

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go
	go build -o bin/server cmd/server/main.go
	go build -o bin/worker worker/main.go

run: install manifests generate ## Run a controller from your host.
	go run ./main.go -default-plugins-names ${PLUGINS} -worker-image ${WORKER_IMG}
run-server: install manifests generate ## Run Zora's server locally.
	go run ./cmd/server/main.go

docker-build: test ## Build manager docker image.
	docker build -t ${IMG} -f ${DOCKERFILE} .
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

docker-build-operator-worker: docker-build ## Build Docker images for the operator and worker components.
	$(MAKE) IMG=${WORKER_IMG} DOCKERFILE=Dockerfile.worker docker-build
docker-push-operator-worker: docker-push ## Push Docker images for the operator and worker components.
	$(MAKE) IMG=${WORKER_IMG} DOCKERFILE=Dockerfile.worker docker-push


##@ Deployment

install: manifests kustomize ## Install default configuration (RBAC for plugins) and CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -
	@kubectl apply -f config/rbac/clusterissue_editor_role.yaml
	@kubectl apply -f config/samples/zora_v1alpha1_plugin_popeye.yaml
	@kubectl apply -f config/samples/zora_v1alpha1_plugin_kubescape.yaml
	@kubectl create -f config/rbac/plugins_role_binding.yaml || true

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f -

deploy: docker-build docker-push generate install ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f -

gen-zora-view-kubeconfig: ## Create a service account and config RBAC for it.
	./hack/scripts/gen_zora_view_kubeconfig.sh
setup-zora-view: install ## Create and apply view secret.
	./hack/scripts/setup_zora_view.sh

setup-region-label: ## Add label used by Zora to detect the cluster region.
	./hack/scripts/setup_region_label.sh
setup-local-registry: ## Create a local Docker registry.
	./hack/scripts/setup_local_registry.sh
setup-kind: setup-local-registry ## Start Kind and a local Docker registry.
	./hack/scripts/setup_kind.sh
	$(MAKE) setup-region-label
delete-kind: ## Delete Kind node.
	kind delete cluster
setup-minikube:  ## Start Minikube with an inner Docker registry.
	minikube start --addons="registry" \
		--driver=docker \
		--cni=kindnet \
		--container-runtime=containerd \
		--insecure-registry="${MINIK_ADDR}:${REG_PORT}" \
		--extra-config="kubelet.container-runtime-endpoint='http://${MINIK_ADDR}:${REG_PORT}"
	$(MAKE) setup-region-label
delete-minikube: ## Delete Minikube node.
	minikube delete


##@ Documentation

helm-docs: ## Generate documentation for helm charts
	@docker run -it --rm \
		-v $(PWD):/helm-docs \
		registry.undistro.io/dockerhub/jnorwood/helm-docs:v1.8.1 \
		helm-docs -s=file --badge-style="flat-square&color=38C794"

preview-docs: helm-docs ## Run a server to preview the documentation
	@docker run --name zora-docs-preview --rm -it \
		-p 8000:8000 \
		-v $(PWD)/mkdocs.yml:/docs/mkdocs.yml \
		-v $(PWD)/docs:/docs/docs \
		-v $(PWD)/charts/zora/README.md:/docs/docs/helm-chart.md \
		-v $(PWD)/charts/zora/values.yaml:/docs/docs/values.yaml \
		squidfunk/mkdocs-material:8.3.8

license: addlicense ## Add license header to source files
	$(call addlicense-tool,)

check-license: addlicense ## Check license header to source files
	$(call addlicense-tool,"-check")
