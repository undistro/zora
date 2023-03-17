.DEFAULT_GOAL = build

.PHONY: default \
 controller-gen \
 kustomize \
 envtest \
 fmt \
 vet \
 test \
 manifest-consistency \
 manifests \
 script-consistency \
 generate \
 clientset-gen \
 build \
 run \
 docker-build \
 docker-build-worker \
 docker-push \
 install \
 uninstall \
 deploy \
 undeploy \
 gen-zora-view-kubeconfig \
 setup-zora-view \
 helm-docs


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
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


# go-install-tool will run "go install" for packages passed as arg <$2> and
# install them to the location passed on arg <$1>.
define go-install-tool
  @ test -f $(1) || { \
  set -e ;\
  TMP_DIR=$$(mktemp -d) ;\
  cd $$TMP_DIR ;\
  go mod init tmp ;\
  echo "Downloading $(2)" ;\
  GOBIN=$(PROJECT_ROOT)/bin go install $(2) ;\
  rm -rf $$TMP_DIR ;\
 }
endef

define addlicense-tool
  $(ADDLICENSE) \
  -c "Undistro Authors" \
  -l "apache" \
  -ignore "**/*.png" \
  -ignore "**/*.md" \
  -ignore "**/*.css" \
  -ignore "**/*.html" \
  -ignore ".github/**" \
  -ignore ".idea/**" \
  -v \
  $(1) \
  .
endef
