#!/bin/sh
set -o errexit

CLUSTER_NAME=${CLUSTER_NAME:-"inspected"}
CLUSTER_NS=${CLUSTER_NS:-"undistro-inspect"}
KCONFIG_NAME=${KCONFIG_NAME:-"inspect_view_kubeconfig.yaml"}
KCONFIG_SECRET_NAME=${KCONFIG_SECRET_NAME:-"$CLUSTER_NAME-kubeconfig"}

setup_namespaces() {
	if ! kubectl get namespace $CLUSTER_NS > /dev/null 2>&1; then
		kubectl create namespace $CLUSTER_NS 
	fi
}
setup_kubeconfig_secret() {
	if ! kubectl -n $CLUSTER_NS get secret $KCONFIG_SECRET_NAME > /dev/null 2>&1; then
		kubectl create secret generic $KCONFIG_SECRET_NAME \
      --namespace $CLUSTER_NS \
			--from-file=value=$KCONFIG_NAME
	fi
}

apply_cluster_crd() {
cat << EOF | kubectl apply -f -
apiVersion: inspect.undistro.io/v1alpha1
kind: Cluster
metadata:
  name: $CLUSTER_NAME
  namespace: $CLUSTER_NS
spec:
  kubeconfigRef:
    name: $KCONFIG_SECRET_NAME 
EOF
}


setup_namespaces
setup_kubeconfig_secret
apply_cluster_crd
