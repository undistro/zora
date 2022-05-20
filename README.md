# Undistro Inspect

Undistro Inspect denounces potential issues in your Kubernetes cluster and
provides multi cluster visibility.

- [Installation](#installation)
- [Usage](#usage)
  + [Connect to a cluster](#connect-to-a-cluster)
      - [Generate a kubeconfig file](#generate-a-kubeconfig-file)
      - [Create a secret with your kubeconfig](#create-a-secret-with-your-kubeconfig)
      - [Create a Cluster resource](#create-a-cluster-resource)
  + [List clusters](#list-clusters)
  + [Configure a cluster scan](#configure-a-cluster-scan)
- [Uninstall](#uninstall)
- [Glossary](#glossary)

## Installation

1. Install Undistro Inspect using [Helm](https://helm.sh/docs/):
```shell
helm repo add undistro https://registry.undistro.io/chartrepo/library
helm install undistro-inspect undistro/inspect \
  --set imageCredentials.username=<USERNAME> \
  --set imageCredentials.password=<PASSWORD> \
  -n undistro-inspect \
  --create-namespace
```

These commands deploy Undistro Inspect to the Kubernetes cluster. [This
section](https://github.com/getupio-undistro/inspect/tree/main/charts/inspect)
lists the parameters that can be configured during installation.

## Usage

### Connect to a cluster

To connect a cluster, you must have a kubeconfig file with an authentication
`token`, the target cluster's api-server must be reachable by the management
cluster, and Metrics Server must be deployed on the target cluster.

If you already have Metrics Server deployed and a kubeconfig, 
skip to the [Create a secret with your kubeconfig](#create-a-secret-with-your-kubeconfig) section. 

#### Installing Metrics Server

Metrics Server can be installed through the deployment made available on its
official repository. The following command will install its latest version on
the namespace `kube-system`:

```shell
kubectl apply -f "https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml"
```

For more information, visit the [Metrics Server
documentation](https://github.com/kubernetes-sigs/metrics-server/blob/master/README.md).


#### Generate a kubeconfig file

Most cloud providers have CLI tools, such as Amazon's `aws` and Google Cloud's
`gcloud`, which can be used to obtain an authentication token.

Undistro Inspect needs a _serviceaccount_ token.

> **Important:**
> Ensure you are in the context of the cluster that you want to connect.
> - Display list of contexts: `kubectl config get-contexts`
> - Display the current-context: `kubectl config current-context`
> - Set the default context to my-cluster-name: `kubectl config use-context my-cluster-name`

1. Create the service account with `view` permissions:
```shell
kubectl -n undistro-inspect create serviceaccount inspect-view
cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: inspect-view
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
kubectl create clusterrolebinding inspect-view --clusterrole=inspect-view --serviceaccount=undistro-inspect:inspect-view
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

> **Important:**
> Ensure you are in the context of the management cluster.

```shell
kubectl create secret generic mycluster-kubeconfig \
  -n undistro-inspect \
  --from-file=value=inspect-view-kubeconfig.yml
```

#### Create a Cluster resource

Create a `Cluster` resource referencing the kubeconfig Secret in the same namespace:

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
> 
> Clusters can be grouped by environment with the `inspect.undistro.io/environment` label.
> 
> You can list all clusters from `prod` environment using: `kubectl get clusters -l inspect.undistro.io/environment=prod`

### List clusters

You can see the connected clusters with `kubectl` command:

```shell
$ kubectl get clusters
NAME        VERSION               MEM AVAILABLE   MEM USAGE (%)   CPU AVAILABLE   CPU USAGE (%)   NODES   AGE   READY
mycluster   v1.21.5-eks-bc4871b   10033Mi         3226Mi (32%)    5790m           647m (11%)      3       40d   True
```

> **Tips:**
> - Get clusters from all namespaces using `--all-namespaces` flag
> - Get clusters with additional information using `-o=wide` flag
> - Get the documentation for `clusters` manifests using `kubectl explain clusters`

The cluster list output has the following columns:

- `NAME`: Cluster name
- `VERSION`: Kubernetes version
- `MEM AVAILABLE`: Quantity of memory available
- `MEM USAGE (%)`: Usage of memory in quantity and percentage
- `CPU AVAILABLE`: Quantity of CPU available
- `CPU USAGE (%)`: Usage of CPU in quantity and percentage
- `NODES`: Total of nodes
- `AGE`: Age of the oldest Node in cluster
- `READY`: Indicates whether the cluster is connected
- `PROVIDER`: Cluster's provider (with `-o=wide` flag)
- `REGION`: Cluster's region (`multi-region` if nodes have different `topology.kubernetes.io/region` label) (with `-o=wide` flag)

> **Info:**
> - The quantity of available and in use resources, is an average of all Nodes.
> - Only one provider is displayed in `PROVIDER` column. Different information can be displayed for multi-cloud clusters.
> - Show detailed description of a cluster, including **events**, running `kubectl describe cluster mycluster`.

### Configure a cluster scan

Since your clusters are connected it's possible configure a scan for them
creating a `ClusterScan` resource in the same namespace as `Cluster`:

Here is a sample configuration that scan `mycluster` once an hour.
You can modify per your needs/wants.

```shell
cat << EOF | kubectl apply -f -
apiVersion: inspect.undistro.io/v1alpha1
kind: ClusterScan
metadata:
  name: mycluster
spec:
  clusterRef:
    name: mycluster
  schedule: "* */1 * * *"
EOF
```

Once the cluster is successfully scanned, 
the reported issues are available in `ClusterIssue` resources:

```shell
kubectl get clusterissues -l cluster=mycluster
NAME                CLUSTER     ID        MESSAGE                                SEVERITY   CATEGORY    AGE   TOTAL
mycluster-pop-106   mycluster   POP-106   No resources requests/limits defined   Medium     Container   17s   10
```

It's possible filter issues by cluster, issue ID, severity and category:
```shell
# issues from mycluster
kubectl get clusterissues -l cluster=mycluster

# clusters with issue POP-106
kubectl get clusterissues -l id=POP-106

# issues from mycluster with high severity
kubectl get clusterissues -l cluster=mycluster,severity=high
```

## Uninstall

```shell
helm delete undistro-inspect -n undistro-inspect
kubectl delete namespace undistro-inspect
```

## Glossary

- **Management Cluster**: The only Kubernetes cluster where Undistro Inspect is installed.
- **Target Cluster**: The Kubernetes cluster that you connect to Undistro Inspect (which is running on management cluster).
