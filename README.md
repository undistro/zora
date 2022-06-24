# Zora

Zora denounces potential issues in your Kubernetes cluster and
provides multi cluster visibility.

- [Installation](#installation)
- [Usage](#usage)
  + [Connect to a cluster](#connect-to-a-cluster)
      - [Generate a kubeconfig file](#generate-a-kubeconfig-file)
      - [Create a Secret with your kubeconfig](#create-a-secret-with-your-kubeconfig)
      - [Create a Cluster resource](#create-a-cluster-resource)
  + [List clusters](#list-clusters)
  + [Configure a cluster scan](#configure-a-cluster-scan)
- [Uninstall](#uninstall)
- [Glossary](#glossary)

## Installation

1. Install Zora using [Helm](https://helm.sh/docs/):
```shell
helm repo add undistro https://registry.undistro.io/chartrepo/library
helm install zora undistro/zora \
  --set imageCredentials.username=<USERNAME> \
  --set imageCredentials.password=<PASSWORD> \
  -n zora-system \
  --create-namespace
```

These commands deploy Zora to the Kubernetes cluster. [This
section](https://github.com/getupio-undistro/zora/tree/main/charts/zora)
lists the parameters that can be configured during installation.

## Usage

### Connect to a cluster

**Before you begin**

- You must have a kubeconfig file with an authentication `token` of the target cluster.
- The api-server of target cluster must be reachable by the management cluster.
- The target cluster must have Metrics Server deployed. For more information, visit the 
[official documentation](https://github.com/kubernetes-sigs/metrics-server/#readme).

If you already have a kubeconfig, 
skip to the [Create a Secret with your kubeconfig](#create-a-secret-with-your-kubeconfig) section. 

#### Generate a kubeconfig file

Most cloud providers have CLI tools, such as `aws` and `gcloud`, which can be used to obtain an authentication token.

Zora needs a _ServiceAccount_ token.

> **Important:**
> Ensure you are in the context of the cluster that you want to connect.
>
> - Display list of contexts: `kubectl config get-contexts`
> - Display the current-context: `kubectl config current-context`
> - Set the default context to my-cluster-name: `kubectl config use-context my-cluster-name`

1. Create the service account with `view` permissions:
```shell
kubectl -n zora-system create serviceaccount zora-view
cat << EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: zora-view
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
kubectl create clusterrolebinding zora-view --clusterrole=zora-view --serviceaccount=zora-system:zora-view
```

2. Check which version of Kubernetes your cluster is running, then proceed to section 2.2 for version 1.24.0 or later, otherwise, follow up on section 2.1.

2.1. For versions prior to 1.24.0, set the `TOKEN_NAME` variable as follows:

```shell
export TOKEN_NAME=$(kubectl -n zora-system get serviceaccount zora-view -o=jsonpath='{.secrets[0].name}')
```

2.2. For clusters running Kubernetes 1.24.0 or later, create a Secret to generate a ServiceAccount token and set the `TOKEN_NAME` variable with the Secret name as follows:

```shell
export TOKEN_NAME="zora-view-token"

cat << EOF | kubectl apply -f - 
apiVersion: v1
kind: Secret
metadata:
  name: "$TOKEN_NAME"
  namespace: "zora-system"
  annotations:
    kubernetes.io/service-account.name: "zora-view"
type: kubernetes.io/service-account-token
EOF
```


3. Set up the remaining environment variables:

```shell
export TOKEN_VALUE=$(kubectl -n zora-system get secret ${TOKEN_NAME} -o=jsonpath='{.data.token}' | base64 --decode)
export CURRENT_CONTEXT=$(kubectl config current-context)
export CURRENT_CLUSTER=$(kubectl config view --raw -o=go-template='{{range .contexts}}{{if eq .name "'''${CURRENT_CONTEXT}'''"}}{{ index .context "cluster" }}{{end}}{{end}}')
export CLUSTER_CA=$(kubectl config view --raw -o=go-template='{{range .clusters}}{{if eq .name "'''${CURRENT_CLUSTER}'''"}}"{{with index .cluster "certificate-authority-data" }}{{.}}{{end}}"{{ end }}{{ end }}')
export CLUSTER_SERVER=$(kubectl config view --raw -o=go-template='{{range .clusters}}{{if eq .name "'''${CURRENT_CLUSTER}'''"}}{{ .cluster.server }}{{end}}{{ end }}')
```

4. Generate a kubeconfig file:
```shell
cat << EOF > zora-view-kubeconfig.yml
apiVersion: v1
kind: Config
current-context: ${CURRENT_CONTEXT}
contexts:
- name: ${CURRENT_CONTEXT}
  context:
    cluster: ${CURRENT_CONTEXT}
    user: zora-view
    namespace: zora-system
clusters:
- name: ${CURRENT_CONTEXT}
  cluster:
    certificate-authority-data: ${CLUSTER_CA}
    server: ${CLUSTER_SERVER}
users:
- name: zora-view
  user:
    token: ${TOKEN_VALUE}
EOF
```

#### Create a Secret with your kubeconfig

> **Important:**
> Ensure you are in the context of the management cluster.

```shell
kubectl create secret generic mycluster-kubeconfig \
  -n zora-system \
  --from-file=value=zora-view-kubeconfig.yml
```

#### Create a Cluster resource

Create a `Cluster` resource referencing the kubeconfig Secret in the same namespace:

```shell
cat << EOF | kubectl apply -f -
apiVersion: zora.undistro.io/v1alpha1
kind: Cluster
metadata:
  name: mycluster
  namespace: zora-system
  labels:
    zora.undistro.io/environment: prod
spec:
  kubeconfigRef:
    name: mycluster-kubeconfig
EOF
```

> **Tip:**
> 
> Clusters can be grouped by environment with the `zora.undistro.io/environment` label.
> 
> You can list all clusters from `prod` environment using: `kubectl get clusters -l zora.undistro.io/environment=prod`

### List clusters

You can see the connected clusters with `kubectl` command:

```shell
kubectl get clusters -o wide
NAME        VERSION               MEM AVAILABLE   MEM USAGE (%)   CPU AVAILABLE   CPU USAGE (%)   NODES   READY   AGE   PROVIDER   REGION      ISSUES
mycluster   v1.21.5-eks-bc4871b   10033Mi         3226Mi (32%)    5790m           647m (11%)      3       True    40d   aws        us-east-1   22
```

> **Tips:**
>
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
- `READY`: Indicates whether the cluster is connected
- `AGE`: Age of the kube-system namespace in cluster
- `PROVIDER`: Cluster provider (with `-o=wide` flag)
- `REGION`: Cluster region (`multi-region` if nodes have different `topology.kubernetes.io/region` label) (with `-o=wide` flag)
- `ISSUES`: Total of issues reported in this Cluster (with `-o=wide` flag)

> **Info:**
>
> - The quantity of available and in use resources, is a sum of all Nodes.
> - Only one provider is displayed in `PROVIDER` column. Different information can be displayed for multi-cloud clusters.
> - Show detailed description of a cluster, including **events**, running `kubectl describe cluster mycluster`.

### Configure a cluster scan

Since your clusters are connected it's possible configure a scan for them
creating a `ClusterScan` resource in the same namespace as `Cluster`:

Here is a sample configuration that scan `mycluster` once an hour.
You can modify per your needs/wants.

```shell
cat << EOF | kubectl apply -f -
apiVersion: zora.undistro.io/v1alpha1
kind: ClusterScan
metadata:
  name: mycluster
spec:
  clusterRef:
    name: mycluster
  schedule: "0 */1 * * *"
EOF
```

Once the cluster is successfully scanned, 
the reported issues are available in `ClusterIssue` resources:

```shell
kubectl get clusterissues -l cluster=mycluster
NAME                          CLUSTER      ID         MESSAGE                                                                        SEVERITY   CATEGORY          AGE
mycluster-pop-102-27557035    mycluster    POP-102    No probes defined                                                              Medium     pods              4m8s
mycluster-pop-105-27557035    mycluster    POP-105    Liveness probe uses a port#, prefer a named port                               Low        pods              4m8s
mycluster-pop-106-27557035    mycluster    POP-106    No resources requests/limits defined                                           Medium     daemonsets        4m8s
mycluster-pop-1100-27557035   mycluster    POP-1100   No pods match service selector                                                 High       services          4m8s
mycluster-pop-306-27557035    mycluster    POP-306    Container could be running as root user. Check SecurityContext/Image           Medium     pods              4m8s
mycluster-pop-500-27557035    mycluster    POP-500    Zero scale detected                                                            Medium     deployments       4m8s
```

It's possible filter issues by cluster, issue ID, severity and category:
```shell
# issues from mycluster
kubectl get clusterissues -l cluster=mycluster

# clusters with issue POP-106
kubectl get clusterissues -l id=POP-106

# issues from mycluster with high severity
kubectl get clusterissues -l cluster=mycluster,severity=high

# only issues reported by the last scan from mycluster
kubectl get clusterissues -l cluster=mycluster,scanID=fa4e63cc-5236-40f3-aa7f-599e1c83208b
```

## Uninstall

```shell
helm delete zora -n zora-system
kubectl delete namespace zora-system
```

## Glossary

- **Management Cluster**: The only Kubernetes cluster where Zora is installed.
- **Target Cluster**: The Kubernetes cluster that you connect to Zora (which is running on management cluster).
