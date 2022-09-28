# Install

Zora requires an existing Kubernetes cluster accessible via `kubectl`.
After the installation process this cluster will be your [management cluster](/glossary#management-cluster)
with the Zora components installed. 
So it is recommended to keep it separated from any application workload.

## Install with Helm

1. To install Zora using [Helm](https://helm.sh/docs/) follow these commands:

```shell
helm repo add undistro https://registry.undistro.io/chartrepo/library
helm repo update undistro
helm upgrade --install zora undistro/zora \
  -n zora-system \
  --create-namespace --wait
```

These commands deploy Zora to the Kubernetes cluster.
[This section](helm-chart.md) lists the parameters
that can be configured during installation.

## Access to the UI

The output of `helm install` and `helm upgrade` commands
contains instructions to access Zora UI based on the provided parameters.

You can get the instructions anytime by running: 

```shell
helm get notes zora -n zora-system
```

## Uninstall

```shell
helm delete zora -n zora-system
kubectl delete namespace zora-system
```
