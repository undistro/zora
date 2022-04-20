#!/bin/sh
set -o errexit

KCONFIG_NAME=${KCONFIG_NAME:-"snitch-view-kubeconfig.yaml"}
CLUSTER_ROLE_NAME=${CLUSTER_ROLE_NAME:-"snitch-view"}
SVC_ACCOUNT_NS=${SVC_ACCOUNT_NS:-"kube-system"}
SVC_ACCOUNT_NAME=${SVC_ACCOUNT_NAME:-"snitch-view"}

get_token_name() {
	echo $(kubectl -n $SVC_ACCOUNT_NS \
		get serviceaccount $SVC_ACCOUNT_NAME \
		-o=jsonpath='{.secrets[0].name}'
	)
}
get_token_value() {
	echo $(kubectl -n $SVC_ACCOUNT_NS \
		get secret $TOKEN_NAME \
		-o=jsonpath='{.data.token}' | base64 --decode
	)
}
get_current_context() {
	echo $(kubectl config current-context)
}
get_current_cluster() {
	echo $(kubectl config view \
		--raw -o=go-template='
			{{range .contexts}}
				{{if eq .name "'$CURRENT_CONTEXT'"}}
					{{index .context "cluster"}}
				{{end}}
			{{end}}
		'
	)
}
get_cluster_ca() {
	echo $(kubectl config view \
		--raw -o=go-template='
			{{range .clusters}}
				{{if eq .name "'$CURRENT_CLUSTER'"}}
					{{with index .cluster "certificate-authority-data"}}
						{{.}}
					{{end}}
				{{end}}
			{{end}}
		'
	)
}
get_cluster_server() {
	echo $(kubectl config view \
		--raw -o=go-template='
			{{range .clusters}}
				{{if eq .name "'$CURRENT_CLUSTER'"}}
					{{ .cluster.server }}
				{{end}}
			{{end}}
		'
	)
}

create_svc_account() {
	kubectl -n $SVC_ACCOUNT_NS create serviceaccount $SVC_ACCOUNT_NAME
}

create_cluster_role() {
cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: $CLUSTER_ROLE_NAME
rules:
  - apiGroups: [ "" ]
    resources:
      - configmaps
      - deployments
      - endpoints
      - horizontalpodautoscalers
      - namespaces
      - nodes
      - persistentvolumes
      - persistentvolumeclaims
      - pods
      - secrets
      - serviceaccounts
      - services
      - statefulsets
    verbs: [ "get", "list" ]
  - apiGroups: [ "rbac.authorization.k8s.io" ]
    resources:
      - clusterroles
      - clusterrolebindings
      - roles
      - rolebindings
    verbs: [ "get", "list" ]
  - apiGroups: [ "metrics.k8s.io" ]
    resources:
      - pods
      - nodes
    verbs: [ "get", "list" ]
EOF
}

create_cluster_role_binding() {
	kubectl create clusterrolebinding $SVC_ACCOUNT_NAME \
		--clusterrole=snitch-view \
		--serviceaccount=$SVC_ACCOUNT_NS:$SVC_ACCOUNT_NAME
}

create_kubeconfig() {
cat << EOF > $KCONFIG_NAME
apiVersion: v1
kind: Config
current-context: $CURRENT_CONTEXT
contexts:
  - name: $CURRENT_CONTEXT
    context:
      cluster: $CURRENT_CONTEXT
      user: snitch-view
      namespace: kube-system
clusters:
  - name: $CURRENT_CONTEXT
    cluster:
      certificate-authority-data: $CLUSTER_CA
      server: $CLUSTER_SERVER
users:
  - name: snitch-view
    user:
      token: $TOKEN_VALUE
EOF
}


setup_namespaces() {
	if ! kubectl get namespace $SVC_ACCOUNT_NS > /dev/null 2>&1; then
		kubectl create namespace $SVC_ACCOUNT_NS
	fi
}
setup_svc_account() {
	if ! kubectl -n $SVC_ACCOUNT_NS get serviceaccount $SVC_ACCOUNT_NAME > /dev/null 2>&1; then
		create_svc_account
	fi
}
setup_cluster_role() {
	if ! kubectl get clusterrole $CLUSTER_ROLE_NAME > /dev/null 2>&1; then
		create_cluster_role
	fi
}
setup_cluster_role_binding() {
	if ! kubectl get -n $SVC_ACCOUNT_NS clusterrolebinding $SVC_ACCOUNT_NAME > /dev/null 2>&1; then
		create_cluster_role_binding
	fi
}


setup_namespaces
setup_svc_account

TOKEN_NAME=${TOKEN_NAME:-"$(get_token_name)"}
TOKEN_VALUE=${TOKEN_VALUE:-"$(get_token_value)"}
CURRENT_CONTEXT=${CURRENT_CONTEXT:-"$(get_current_context)"}
CURRENT_CLUSTER=${CURRENT_CLUSTER:-"$(get_current_cluster)"}
CLUSTER_CA=${CLUSTER_CA:-"$(get_cluster_ca)"}
CLUSTER_SERVER=${CLUSTER_SERVER:-"$(get_cluster_server)"}

setup_cluster_role_binding
setup_cluster_role_binding
create_kubeconfig
