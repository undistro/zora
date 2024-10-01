# Installation

Zora OSS is installed inside your Kubernetes clusters using [Helm](https://helm.sh/){:target="_blank"},
where the `zora-operator` deployment is created and scans are automatically scheduled for your cluster.

## Prerequisites

- Kubernetes cluster 1.21+
- Kubectl
- Helm 3.8+

## Install with Helm

First, [ensure that your current context of `kubectl` refer to the Kubernetes cluster you wish to install Zora into](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/){:target="_blank"}.

??? tip "Manage kubectl contexts"
    The following commands can help you to manage kubectl contexts:

    - List all contexts: `kubectl config get-contexts`

    - Display the current-context: `kubectl config current-context`

    - Use the context for the Kind cluster: `kubectl config use-context kind-kind`

Then, run the following command to install Zora [Helm chart](https://helm.sh/docs/topics/charts/){:target="_blank"}:

=== "HTTP chart repository"
    
    ```shell
    helm repo add undistro https://charts.undistro.io --force-update
    helm repo update undistro
    helm upgrade --install zora undistro/zora \
      -n zora-system \
      --version 0.10.1 \
      --create-namespace \
      --wait \
      --set clusterName="$(kubectl config current-context)"
    ```

=== "OCI registry"

    ```shell
    helm upgrade --install zora oci://ghcr.io/undistro/helm-charts/zora \
      -n zora-system \
      --version 0.10.1 \
      --create-namespace \
      --wait \
      --set clusterName="$(kubectl config current-context)"
    ```

This command will install Zora in `zora-system` namespace, creating the namespace if it doesn't already exist.

!!! info "Zora OSS + Zora Dashboard"
    To integrate your Zora OSS installation with Zora Dashboard, you need to authenticate with the authorization server and provide `saas.workspaceID` parameter in installation command.
    For more information, please refer to [this page](../dashboard.md#getting-started).

With the following commands, you can verify if Zora has been successfully installed and retrieve installation notes:

```shell
helm list -n zora-system
helm get notes zora -n zora-system
```

!!! info "Zora Helm Chart"
    To see the full list of available parameters in Zora Helm chart, please visit [this page](../helm-chart.md)

If everything is set up correctly, your cluster should have scheduled scans. Check it by running:

```shell
kubectl get cluster,scan -o wide -n zora-system
```

!!! tip "Customize scan schedule"
    To customize the scan schedule, please refer to the [Scan Schedule page](../configuration/scan-schedule.md).

Once the cluster is successfully scanned, you can check issues by running:

```shell
kubectl get misconfigurations -n zora-system
kubectl get vulnerabilities   -n zora-system
```

## Migrating to 0.8

### What's new in 0.8

#### Extended Vulnerability Reports Information

Now, `VulnerabilityReports` provide more in-depth information about the image, including `OS`, `architecture`, `distro`, and `digest`.
Additionally, details about vulnerabilities, such as `publishedDate` and `lastModifiedDate`, have been included 
to offer a clearer understanding of your cluster's security posture.

#### Full Integration with Zora Dashboard

Zora 0.8 introduces the integration of Vulnerability Reports with the Zora Dashboard.
Now, alongside misconfigurations, you can centrally explore images and vulnerabilities across your clusters.

### Migration guide

!!! warning "Version 0.7 or earlier"
    If you are currently using a version prior to 0.7, 
    please be aware that the 0.7 release brought about significant architectural changes. 
    Before upgrading to version 0.8, refer to [this page](/v0.7/getting-started/installation/#migrating-to-07) 
    for essential information and considerations to ensure a smooth transition.


The recommended way to migrate to version 0.8 is to reinstall Zora, including its CRDs.

## Uninstall

You can uninstall Zora and its components by uninstalling the Helm chart installed above.

```shell
helm uninstall zora -n zora-system
```

By design, [Helm doesn't upgrade or delete CRDs](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/#some-caveats-and-explanations){:target="_blank"}.
You can permanently delete Zora CRDs and any remaining associated resources from your cluster, using the following command.

```shell
kubectl get crd -o=name | grep --color=never 'zora.undistro.io' | xargs kubectl delete
```

You can also delete the `zora-system` namespace using the command below.

```shell
kubectl delete namespace zora-system
```
