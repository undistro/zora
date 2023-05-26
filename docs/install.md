# Install

Zora requires an existing Kubernetes cluster accessible via `kubectl`.
After the installation process this cluster will be your [management cluster](../glossary#management-cluster)
with the Zora components installed. 
So it is recommended to keep it separated from any application workload.

## Setup Requirements

Zora's management cluster requires these programs in order to be installed and configured:

- Kubernetes >= 1.21.0
- [Helm](https://helm.sh/) >= 3.4.0
- [Kubectl](https://kubernetes.io/docs/reference/kubectl/)
- Awk
- Cat
- POSIX shell

## Install with Helm

!!! warning "Migrating to version 0.6"
    If you already have Zora installed and want to migrate to zora 0.6, 
    you need to follow some additional steps.
    
    Before running the `helm upgrade` command, it is necessary to apply the modified CRDs.
    ```shell
    kubectl apply -f https://raw.githubusercontent.com/undistro/zora/v0.6.0/charts/zora/crds/zora.undistro.io_clusterissues.yaml
    kubectl apply -f https://raw.githubusercontent.com/undistro/zora/v0.6.0/charts/zora/crds/zora.undistro.io_customchecks.yaml
    kubectl apply -f https://raw.githubusercontent.com/undistro/zora/v0.6.0/charts/zora/crds/zora.undistro.io_plugins.yaml
    ```
    These commands ensure that the modified Custom Resource Definitions (CRDs) are applied correctly.
    
    By default, Helm does not upgrade CRDs automatically, which is why this manual step is necessary.

1. To install Zora using [Helm](https://helm.sh/docs/) follow these commands:

=== "Zora + SaaS"

    In this option you have access to the powerful dashboard to see your clusters and issues.

    !!! warning
        The SaaS (`https://zora-dashboard.undistro.io/`) must be reachable by Zora.

    1.1 Sign in at [https://zora-dashboard.undistro.io/](https://zora-dashboard.undistro.io/) and select a workspace

    1.2 Get your workspace ID by clicking on :material-cloud-download: and provide it by the `saas.workspaceID` flag:

    ```shell
    helm repo add undistro https://charts.undistro.io --force-update
    helm repo update undistro
    helm upgrade --install zora undistro/zora \
      --set saas.workspaceID='<YOUR WORKSPACE ID>'
      -n zora-system \
      --version 0.6.1-rc1 \
      --create-namespace --wait
    ```

===  "Zora (kubectl)"

    In this option you can see your clusters and issues by `kubectl`.

    ```shell
    helm repo add undistro https://charts.undistro.io --force-update
    helm repo update undistro
    helm upgrade --install zora undistro/zora \
      -n zora-system \
      --version 0.6.1-rc1 \
      --create-namespace --wait
    ```

!!! info
    The Helm chart repository has been updated from `https://registry.undistro.io/chartrepo/library` to `https://charts.undistro.io`.

    The `--force-update` flag is needed to update the repository URL.

These commands deploy Zora to the Kubernetes cluster.
[This section](helm-chart.md) lists the parameters
that can be configured during installation.

## Access to the dashboard

If you installed Zora providing a workspace ID (Zora + SaaS), 
you have access to the powerful dashboard at [https://zora-dashboard.undistro.io/](https://zora-dashboard.undistro.io/)

The output of `helm install` and `helm upgrade` commands
contains the dashboard URL and you can get it anytime by running: 

```shell
helm get notes zora -n zora-system
```

## Uninstall

```shell
helm delete zora -n zora-system
kubectl delete namespace zora-system
```
