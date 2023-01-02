# Install

Zora requires an existing Kubernetes cluster accessible via `kubectl`.
After the installation process this cluster will be your [management cluster](/glossary#management-cluster)
with the Zora components installed. 
So it is recommended to keep it separated from any application workload.

## Setup Requirements

Zora's management cluster requires these programs in order to be installed and configured:

- Kubernetes >= 1.21.0
- Helm >= 3.4.0
- Kubectl
- Awk
- Cat
- POSIX shell


## Install with Helm

1. To install Zora using [Helm](https://helm.sh/docs/) follow these commands:

=== "Zora + SaaS"

    In this option you have access to the powerful dashboard to see your clusters and issues.

    !!! warning
        The SaaS (`https://saas-hml.undistro.io/`) must be reachable by Zora.

    1.1 Sign in at [https://saas-hml.undistro.io/](https://saas-hml.undistro.io/) and select a workspace

    1.2 Get your workspace ID by clicking on :material-cloud-download: and provide it by the `saas.workspaceID` flag:

    ```shell
    helm repo add undistro https://charts.undistro.io --force-update
    helm repo update undistro
    helm upgrade --install zora undistro/zora \
      --set saas.workspaceID='<YOUR WORKSPACE ID>'
      -n zora-system \
      --create-namespace --wait
    ```

===  "Zora (kubectl)"

    In this option you can see your clusters and issues by `kubectl`.

    ```shell
    helm repo add undistro https://charts.undistro.io --force-update
    helm repo update undistro
    helm upgrade --install zora undistro/zora \
      -n zora-system \
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
you have access to the powerful dashboard at [https://saas-hml.undistro.io/](https://saas-hml.undistro.io/)

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
