#!/bin/sh
set -o errexit

CLUSTER_NAME=${CLUSTER_NAME:-"zored"}
CLUSTER_NS=${CLUSTER_NS:-"zora-system"}
KCONFIG_PATH=${KCONFIG_PATH:-"zora_view_kubeconfig.yaml"}
KCONFIG_SECRET_NAME=${KCONFIG_SECRET_NAME:-"$CLUSTER_NAME-kubeconfig"}
ENABLE_CLUSTER_SCAN=${ENABLE_CLUSTER_SCAN:-0}
CLUSTER_SCAN_SCHEDULE=${CLUSTER_SCAN_SCHEDULE:-'*/2 * * * *'}

setup_namespaces() {
	if ! kubectl get namespace $CLUSTER_NS > /dev/null 2>&1; then
		kubectl create namespace $CLUSTER_NS 
	fi
}
setup_kubeconfig_secret() {
	if ! kubectl -n $CLUSTER_NS get secret $KCONFIG_SECRET_NAME > /dev/null 2>&1; then
		kubectl create secret generic $KCONFIG_SECRET_NAME \
		  --namespace $CLUSTER_NS \
		  --from-file=value=$KCONFIG_PATH
	fi
}

apply_cluster_crd() {
cat << EOF | kubectl apply -f -
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

apply_plugin_crds(){
	kubectl -n $CLUSTER_NS apply -f config/samples/zora_v1alpha1_plugin_popeye.yaml
	kubectl -n $CLUSTER_NS apply -f config/samples/zora_v1alpha1_plugin_kubescape.yaml
}

apply_clusterscan_crd(){
cat << EOF | kubectl apply -f -
apiVersion: zora.undistro.io/v1alpha1
kind: ClusterScan
metadata:
  name: $CLUSTER_NAME-scan
  namespace: $CLUSTER_NS
spec:
  clusterRef:
    name: $CLUSTER_NAME
  schedule: "$CLUSTER_SCAN_SCHEDULE"
EOF
}

setup_namespaces
setup_kubeconfig_secret
apply_cluster_crd
if test $ENABLE_CLUSTER_SCAN -eq 1; then
	apply_plugin_crds
	apply_clusterscan_crd
fi
