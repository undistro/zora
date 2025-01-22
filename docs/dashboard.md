---
title: Zora Dashboard
---
# Zora Dashboard

Zora Dashboard is a SaaS platform designed to seamlessly centralize the security posture management of all your
Kubernetes clusters, providing a full experience powered by Zora OSS.

It features a powerful UI that allows you to navigate, filter and explore details of issues and affected resources
across all your clusters. You can also invite users to your workspace.

![Zora Dashboard Screenshot](assets/zora-dashboard-screenshot.png)

<div align="center">
   <a href="https://zora-dashboard.undistro.io/" class="md-button">Try Zora Dashboard</a>
</div>

Zora Dashboard offers a ***starter plan for 14 days***, after which it will revert to the free plan which provides access for 2 clusters with up to 10 nodes per cluster.
Please [contact us](https://undistro.io/contact){:target="_blank"} if you need to discuss a tailored solution.

## Getting started

To integrate your Zora OSS installation with Zora Dashboard, you need to first authenticate with the authorization server and then provide your `saas.workspaceID` parameter in the Zora OSS installation command.

### Authenticating with the Authorization server
Authenticating with the authorization server is simplified through the use of a helm plugin, `zoraauth`, which can be installed by executing

```console
helm plugin install https://github.com/undistro/helm-zoraauth
```
and updated by executing
```console
helm plugin update zoraauth
```
The authentication process will occur when the plugin is executed, and you visit the authorization server to confirm the request. The instructions within the [Zora Dashboard](https://zora-dashboard.undistro.io/){:target="_blank"} console will include the appropriate parameters for the plugin, these can be obtained through the `Connect cluster` option once you have signed in to the [Zora Dashboard](https://zora-dashboard.undistro.io/){:target="_blank"}.

To authenticate with the authorization server, copy and run the `helm zoraauth` command and then follow the instructions within your terminal
```console
helm zoraauth --audience="zora_prod" \
  --client-id="<client id>" \
  --domain="login.undistro.io"
Initiating Device Authorization Flow...
Please visit https://login.undistro.io/activate and enter code: BFNS-NWFF, or visit: https://login.undistro.io/activate?user_code=BFNS-NWFF
```
Entering the login URL within your browser will present you with a screen similar to the following

<figure markdown="span">
  ![Zora Device Confirmation](assets/zora-device-confirmation.png){ width="300" }
</figure>

Once you have confirmed the request you should see the following message on your terminal

```console
Tokens saved to tokens.yaml
```

You can then install or upgrade Zora OSS by providing the `saas.workspaceID` parameter in the [Zora OSS installation command](getting-started/installation.md):

=== "HTTP chart repository"
    
    ```shell hl_lines="6 7"
    helm repo add undistro https://charts.undistro.io --force-update
    helm repo update undistro
    helm upgrade --install zora undistro/zora \
      -n zora-system --create-namespace --wait \
      --set clusterName="$(kubectl config current-context)" \
      --set saas.workspaceID=<YOUR WORKSPACE ID HERE> \
      --values tokens.yaml
    ```

=== "OCI registry"

    ```shell hl_lines="4 5"
    helm upgrade --install zora oci://ghcr.io/undistro/helm-charts/zora \
      -n zora-system --create-namespace --wait \
      --set clusterName="$(kubectl config current-context)" \
      --set saas.workspaceID=<YOUR WORKSPACE ID HERE> \
      --values tokens.yaml
    ```


## Architecture

Zora OSS acts as the engine of Zora Dashboard, meaning that once scans are completed,
**only the results are sent to Zora Dashboard**, where they are accessible by you
and those you have invited to your workspace.

![Zora Architecture Diagram](assets/dashboard-arch-light.png#only-light)
![Zora Architecture Diagram](assets/dashboard-arch-dark.png#only-dark)

Note that these results do not contain sensitive information or specific data about your cluster configuration.
