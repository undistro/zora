# Snitch

Snitch denounces potential issues in your Kubernetes cluster
and provides multi cluster visibility.

- [Install](#install)
- [Usage](#usage)
    + [Connect a cluster](#connect-a-cluster)
        - [Generate a kubeconfig file](#generate-a-kubeconfig-file)
        - [Create a secret with your kubeconfig](#create-a-secret-with-your-kubeconfig)
        - [Create a Cluster resource](#create-a-cluster-resource)
- [Uninstall](#uninstall)
- [Glossary](#glossary)

## Install

1. Install Snitch using [Helm](https://helm.sh/docs/):
```shell
helm repo add snitch https://registry.undistro.io/chartrepo/snitch
helm install snitch snitch/snitch -n snitch-system --create-namespace
```

These commands deploy Snitch to the Kubernetes cluster. 
[This section](https://github.com/getupio-undistro/snitch/tree/main/charts/snitch) lists the parameters that can be configured during installation.

## Usage

### Connect a cluster

To connect a cluster, you must have a kubeconfig file with a `token`, and
the target cluster's api-server must be reachable by the management cluster.

If you already have a kubeconfig, 
skip the next step and go to the [Create a secret with your kubeconfig](#create-a-secret-with-your-kubeconfig) section. 

#### Generate a kubeconfig file

Most cloud providers have CLI tools, such as Amazon's `aws` and Google Cloud's
`gcloud`, which can be used to obtain an authentication token.

Snitch just needs a _serviceaccount_ token.

1. Create the service account with `view` permissions:
```shell
kubectl -n kube-system create serviceaccount snitch-view
cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: snitch-view
  namespace: snitch-system
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
kubectl -n snitch-system create clusterrolebinding snitch-view --clusterrole=snitch-view --serviceaccount=kube-system:snitch-view
```

2. Set up the following environment variables:
```shell
export TOKEN_NAME=$(kubectl -n snitch-system get serviceaccount snitch-view -o=jsonpath='{.secrets[0].name}')
export TOKEN_VALUE=$(kubectl -n snitch-system get secret ${TOKEN_NAME} -o=jsonpath='{.data.token}' | base64 --decode)
export CURRENT_CONTEXT=$(kubectl config current-context)
export CURRENT_CLUSTER=$(kubectl config view --raw -o=go-template='{{range .contexts}}{{if eq .name "'''${CURRENT_CONTEXT}'''"}}{{ index .context "cluster" }}{{end}}{{end}}')
export CLUSTER_CA=$(kubectl config view --raw -o=go-template='{{range .clusters}}{{if eq .name "'''${CURRENT_CLUSTER}'''"}}"{{with index .cluster "certificate-authority-data" }}{{.}}{{end}}"{{ end }}{{ end }}')
export CLUSTER_SERVER=$(kubectl config view --raw -o=go-template='{{range .clusters}}{{if eq .name "'''${CURRENT_CLUSTER}'''"}}{{ .cluster.server }}{{end}}{{ end }}')
```

3. Generate a kubeconfig file:
```shell
cat << EOF > snitch-view-kubeconfig.yml
apiVersion: v1
kind: Config
current-context: ${CURRENT_CONTEXT}
contexts:
- name: ${CURRENT_CONTEXT}
  context:
    cluster: ${CURRENT_CONTEXT}
    user: snitch-view
    namespace: snitch-system
clusters:
- name: ${CURRENT_CONTEXT}
  cluster:
    certificate-authority-data: ${CLUSTER_CA}
    server: ${CLUSTER_SERVER}
users:
- name: snitch-view
  user:
    token: ${TOKEN_VALUE}
EOF
```

#### Create a secret with your kubeconfig

```shell
kubectl create secret generic mycluster-kubeconfig \
  -n snitch-system \
  --from-file=value=snitch-view-kubeconfig.yml
```

#### Create a Cluster resource

```shell
cat << EOF | kubectl apply -f -
apiVersion: snitch.undistro.io/v1alpha1
kind: Cluster
metadata:
  name: mycluster
  namespace: snitch-system
spec:
  kubeconfigRef:
    name: mycluster-kubeconfig
EOF
```

## Uninstall

```shell
helm delete snitch -n snitch
kubectl delete namespace snitch-system
```

## Glossary

- **Management Cluster**: The only Kubernetes cluster where Snitch is installed.
