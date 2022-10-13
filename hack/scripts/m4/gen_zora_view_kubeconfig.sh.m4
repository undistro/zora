include(common_opts_and_vars.sh.in)
KCONFIG_NAME=${KCONFIG_NAME:-"zora_view_kubeconfig.yaml"}
METRICS_SERVER_VERSION=${METRICS_SERVER_VERSION:-"latest"}
METRICS_SERVER_DEPLOYMENT_NAME=${METRICS_SERVER_DEPLOYMENT_NAME:-"metrics-server"}
METRICS_SERVER_DEPLOYMENT=${METRICS_SERVER_DEPLOYMENT:-"https://github.com/kubernetes-sigs/metrics-server/releases/$METRICS_SERVER_VERSION/download/components.yaml"}

include(common_get_funcs.sh.in)
include(common_create_funcs.sh.in)

setup_metrics_server() {
	if ! kubectl get pods -A 2> /dev/null | grep -q $METRICS_SERVER_DEPLOYMENT_NAME; then
		kubectl apply -f "$METRICS_SERVER_DEPLOYMENT"
	fi
}

include(common_setup_funcs.sh.in)

include(common_calls_and_vars.sh.in)

setup_metrics_server
include(common_calls.sh.in)
