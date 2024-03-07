# Azure Container Registry (ACR)

If you are running within Azure, and making use of a private [Azure Container Registry (ACR)](https://learn.microsoft.com/en-us/azure/container-registry/){:target="_blank"}
to host your application images, then the Trivy plugin will be unable to scan those images unless access is granted to 
the registry through a service principal with `AcrPull` role assigned.

## Creating service principal

The following [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/){:target="_blank"} command creates a service principal 
with `AcrPull` role assigned, and stores the output including the credentials into `SP_DATA` environment variable.

!!! note
    Please replace `<SUBSCRIPTION_ID>`, `<RESOURCE_GROUP>`, and `<REGISTRY_NAME>` before running the command below.

```shell
export SP_DATA=$(az ad sp create-for-rbac --name ZoraTrivy --role AcrPull --scope "/subscriptions/<SUBSCRIPTION_ID>/resourceGroups/<RESOURCE_GROUP>/providers/Microsoft.ContainerRegistry/registries/<REGISTRY_NAME>")
```

## Usage

Once the service principal is created and the credentials are in `SP_DATA` environment variable,
create a Kubernetes secret to store these credentials by running:

```shell
kubectl create secret generic trivy-acr-credentials -n zora-system \
  --from-literal=AZURE_CLIENT_ID=$(echo $SP_DATA | jq -r '.appId') \
  --from-literal=AZURE_CLIENT_SECRET=$(echo $SP_DATA | jq -r '.password') \
  --from-literal=AZURE_TENANT_ID=$(echo $SP_DATA | jq -r '.tenant')
```

!!! note
    If you are running this command before a Zora installation, you may need to create the `zora-system` namespace.
    ```shell
    kubectl create namespace zora-system
    ```

Now set the secret name in a `values.yaml`

```yaml hl_lines="6"
scan:
  plugins:
    trivy:
      envFrom:
        - secretRef:
            name: trivy-acr-credentials
            optional: true
```

Then provide it in `helm upgrade --install` command

```shell
-f values.yaml
```

This will now allow the Trivy plugin to scan your internal images for vulnerabilities.
