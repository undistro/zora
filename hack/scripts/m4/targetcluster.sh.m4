include(common_opts_and_vars.sh.in)

include(common_get_funcs.sh.in)
include(common_create_funcs.sh.in)
include(common_setup_funcs.sh.in)

show_generated_kconfig_name() {
	echo "Kubeconfing file:
	$KCONFIG_NAME
	"
}

show_kconfig_creation_cmd() {
	echo "Create a Kubeconfig Secret on the management cluster by running:
	kubectl create secret generic $KCONFIG_SECRET_NAME \\
		--namespace $CLUSTER_NS \\
		--from-file=value=$KCONFIG_NAME
"
}

create_cluster_sample() {
	cat << EOF > $SAMPLE_MANIFEST_NAME
apiVersion: zora.undistro.io/v1alpha1
kind: Cluster
metadata:
  name: $CLUSTER_NAME
  namespace: $CLUSTER_NS
spec:
  kubeconfigRef:
	name: $KCONFIG_SECRET_NAME 
EOF
}

show_cluster_sample_name() {
	echo "Sample manifest:
	$SAMPLE_MANIFEST_NAME
	"
}


include(common_calls_and_vars.sh.in)
CLUSTER_NS=${CLUSTER_NS:-$SVC_ACCOUNT_NS}
KCONFIG_NAME=${KCONFIG_NAME:-"$CONTEXT-kubeconfig.yaml"}
KCONFIG_SECRET_NAME=${KCONFIG_SECRET_NAME:-"$CLUSTER_NAME-kubeconfig"}
SAMPLE_MANIFEST_NAME=${SAMPLE_MANIFEST_NAME:-"cluster_sample.yaml"}
include(common_calls.sh.in)

echo
show_generated_kconfig_name
show_kconfig_creation_cmd
create_cluster_sample
show_cluster_sample_name
