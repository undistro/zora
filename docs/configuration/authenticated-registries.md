# Authenticated Registries

Trivy plugin is able to scan images from registries that require authentication.

It's necessary to create a secret containing authentication credentials as pairs, like the command below.

!!! note
    For [AWS ECR](private-registries/ecr.md) and [Azure ACR](private-registries/acr.md) registries, please refer to the specific pages.

```shell
kubectl create secret generic trivy-credentials -n zora-system \
  --from-literal=TRIVY_USERNAME="<username1>,<username2>" \
  --from-literal=TRIVY_PASSWORD="<password1>,<password2>"
```

!!! note
    Please note that the number of usernames and passwords must be the same.

Once the secret is created, it needs to be referenced in the Helm chart parameters as the following `values.yaml` file:

```yaml hl_lines="6"
scan:
  plugins:
    trivy:
      envFrom:
        - secretRef:
            name: trivy-credentials
            optional: true
```

Then provide this file in `helm upgrade --install` command with `-f values.yaml` flag.

This ensures that Trivy can authenticate with the private registries using the provided credentials.
