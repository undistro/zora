# Install

Zora requires an existing Kubernetes cluster accessible via `kubectl`.
During the installation process the Kubernetes cluster will be your [management cluster](/glossary#management-cluster)
by installing the Zora components, so it is recommended to keep it separated from any application workload.

## Install with Helm

1. Install Zora using [Helm](https://helm.sh/docs/):

```shell
helm repo add undistro https://registry.undistro.io/chartrepo/library
helm upgrade --install zora undistro/zora \
  --set imageCredentials.username='<USERNAME>' \
  --set imageCredentials.password='<PASSWORD>' \
  -n zora-system \
  --create-namespace
```

!!! warning
    `<USERNAME>` and `<PASSWORD>` must be replaced with your credentials.

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
