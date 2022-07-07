# Connect the target cluster to Zora

Follow this guide to connect a [target cluster](/glossary#target-cluster) directly to Zora.

## Prerequisites

1. A kubeconfig file with an authentication `token` of the target cluster. 
   Follow [these instructions](/target-cluster) to generate it.
2. The [api-server](https://kubernetes.io/docs/concepts/overview/components/#kube-apiserver) 
   of the [target cluster](/glossary#target-cluster) 
   must be reachable by the [management cluster](/glossary#management-cluster). 


!!! warning "Metrics Server"
    If the target cluster has Metrics Server deployed, 
    information about the resource usage is collected
    and issues about potential resources over/under allocations can be reported.

    For more information about Metrics Server, visit the
    [official documentation](https://github.com/kubernetes-sigs/metrics-server/#readme).

!!! warning "Important"
    Ensure you are in the context of the **management cluster**.

    - Display list of contexts: `kubectl config get-contexts`
    - Display the current-context: `kubectl config current-context`
    - Set the default context to **my-management-cluster**: `kubectl config use-context my-management-cluster`

## Create a `Cluster` resource

First, create a `Secret` with the content of the kubeconfig file:

```shell
kubectl create secret generic mycluster-kubeconfig \
  -n zora-system \
  --from-file=value=zora-view-kubeconfig.yml
```

Now, you are able to create a `Cluster` resource referencing the kubeconfig Secret in the same namespace:

```yaml
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

!!! tip
    Clusters can be grouped by environment with the `zora.undistro.io/environment` label.
    
    You can list all clusters from `prod` environment using: `kubectl get clusters -l zora.undistro.io/environment=prod`


## List clusters

You can list the connected clusters with `kubectl` command:

```shell
kubectl get clusters -o wide
NAME        VERSION               MEM AVAILABLE   MEM USAGE (%)   CPU AVAILABLE   CPU USAGE (%)   NODES   READY   AGE   PROVIDER   REGION      ISSUES
mycluster   v1.21.5-eks-bc4871b   10033Mi         3226Mi (32%)    5790m           647m (11%)      3       True    40d   aws        us-east-1   22
```

!!! tip
    - Get clusters from all namespaces using `--all-namespaces` flag
    - Get clusters with additional information using `-o=wide` flag
    - Get the documentation for `clusters` manifests using `kubectl explain clusters`

The cluster list output has the following columns:

- `NAME`: Cluster name
- `VERSION`: Kubernetes version
- `MEM AVAILABLE`: Quantity of memory available (requires Metrics Server)
- `MEM USAGE (%)`: Usage of memory in quantity and percentage (requires Metrics Server)
- `CPU AVAILABLE`: Quantity of CPU available (requires Metrics Server)
- `CPU USAGE (%)`: Usage of CPU in quantity and percentage (requires Metrics Server)
- `NODES`: Total of nodes
- `READY`: Indicates whether the cluster is connected
- `AGE`: Age of the kube-system namespace in cluster
- `PROVIDER`: Cluster provider (with `-o=wide` flag)
- `REGION`: Cluster region (`multi-region` if nodes have different `topology.kubernetes.io/region` label)
  (with `-o=wide` flag)
- `ISSUES`: Total of issues reported in the last successful scan of this Cluster (with `-o=wide` flag)

!!! info
    - The quantity of available and in use resources, is a sum of all Nodes.
    - Only one provider is displayed in `PROVIDER` column. Different information can be displayed for multi-cloud clusters.
    - Show detailed description of a cluster, including **events**, running `kubectl describe cluster mycluster`.
