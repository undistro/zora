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
helm repo add undistro https://registry.undistro.io/chartrepo/library
helm install inspect undistro/inspect \
  --set imageCredentials.username=<USERNAME> \
  --set imageCredentials.password=<PASSWORD> \
  -n undistro-inspect \
  --create-namespace
```

These commands deploy Snitch to the Kubernetes cluster. 
[This section](https://github.com/getupio-undistro/inspect/tree/main/charts/inspect) lists the parameters that can be configured during installation.

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
kubectl -n undistro-inspect create serviceaccount inspect-view
cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: inspect-view
  namespace: undistro-inspect
rules:
  - apiGroups: [ "" ]
    resources:
      - configmaps
      - endpoints
      - limitranges
      - namespaces
      - nodes
      - persistentvolumes
      - persistentvolumeclaims
      - pods
      - secrets
      - serviceaccounts
      - services
    verbs: [ "get", "list" ]
  - apiGroups: [ "apps" ]
    resources:
      - daemonsets
      - deployments
      - statefulsets
      - replicasets
    verbs: [ "get", "list" ]
  - apiGroups: [ "autoscaling" ]
    resources:
      - horizontalpodautoscalers
    verbs: [ "get", "list" ]
  - apiGroups: [ "networking.k8s.io" ]
    resources:
      - ingresses
      - networkpolicies
    verbs: [ "get", "list" ]
  - apiGroups: [ "policy" ]
    resources:
      - poddisruptionbudgets
      - podsecuritypolicies
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
kubectl -n undistro-inspect create clusterrolebinding inspect-view --clusterrole=inspect-view --serviceaccount=kube-system:inspect-view
```

2. Set up the following environment variables:
```shell
export TOKEN_NAME=$(kubectl -n undistro-inspect get serviceaccount inspect-view -o=jsonpath='{.secrets[0].name}')
export TOKEN_VALUE=$(kubectl -n undistro-inspect get secret ${TOKEN_NAME} -o=jsonpath='{.data.token}' | base64 --decode)
export CURRENT_CONTEXT=$(kubectl config current-context)
export CURRENT_CLUSTER=$(kubectl config view --raw -o=go-template='{{range .contexts}}{{if eq .name "'''${CURRENT_CONTEXT}'''"}}{{ index .context "cluster" }}{{end}}{{end}}')
export CLUSTER_CA=$(kubectl config view --raw -o=go-template='{{range .clusters}}{{if eq .name "'''${CURRENT_CLUSTER}'''"}}"{{with index .cluster "certificate-authority-data" }}{{.}}{{end}}"{{ end }}{{ end }}')
export CLUSTER_SERVER=$(kubectl config view --raw -o=go-template='{{range .clusters}}{{if eq .name "'''${CURRENT_CLUSTER}'''"}}{{ .cluster.server }}{{end}}{{ end }}')
```

3. Generate a kubeconfig file:
```shell
cat << EOF > inspect-view-kubeconfig.yml
apiVersion: v1
kind: Config
current-context: ${CURRENT_CONTEXT}
contexts:
- name: ${CURRENT_CONTEXT}
  context:
    cluster: ${CURRENT_CONTEXT}
    user: inspect-view
    namespace: undistro-inspect
clusters:
- name: ${CURRENT_CONTEXT}
  cluster:
    certificate-authority-data: ${CLUSTER_CA}
    server: ${CLUSTER_SERVER}
users:
- name: inspect-view
  user:
    token: ${TOKEN_VALUE}
EOF
```

#### Create a secret with your kubeconfig

```shell
kubectl create secret generic mycluster-kubeconfig \
  -n undistro-inspect \
  --from-file=value=inspect-view-kubeconfig.yml
```

#### Create a Cluster resource

```shell
cat << EOF | kubectl apply -f -
apiVersion: inspect.undistro.io/v1alpha1
kind: Cluster
metadata:
  name: mycluster
  namespace: undistro-inspect
  labels:
    inspect.undistro.io/environment: prod
spec:
  kubeconfigRef:
    name: mycluster-kubeconfig
EOF
```

> **Tip:**
> Clusters can be grouped by environment with the `inspect.undistro.io/environment` label.
> 
> You can list all clusters from `prod` environment using: `kubectl get clusters -l inspect.undistro.io/environment=prod`

## Uninstall

```shell
helm delete inspect -n undistro-inspect
kubectl delete namespace undistro-inspect
```

## Glossary

- **Management Cluster**: The only Kubernetes cluster where Snitch is installed.
