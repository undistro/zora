---
title: Trivy Plugin 
---

# Trivy Plugin

Trivy is a versatile security scanner that can find **vulnerabilities**, misconfigurations, secrets, SBOM 
in different targets like containers, code repositories and **Kubernetes cluster**.

**Zora uses Trivy as a plugin exclusively to scan vulnerabilities in a Kubernetes cluster.**

:octicons-codescan-24: **Type**: `vulnerability`

:simple-docker: **Image**: `ghcr.io/undistro/trivy:0.67`

:simple-github: **GitHub repository**: [https://github.com/aquasecurity/trivy](https://github.com/aquasecurity/trivy){:target="_blank"}

## Vulnerability Database Persistence

Trivy utilizes a database containing vulnerability information. 
This database is updated every **6 hours** and persisted by default for caching purposes between the schedule scans.

Please refer to [this page](../configuration/vulnerability-database-persistence.md) for further details and 
configuration options regarding vulnerability database persistence.

## Large vulnerability reports

Vulnerability reports can be large depending on the scanned image.

Zora automatically handles oversized reports:

- Zora automatically splits vulnerability reports recursively when errors such as `etcdserver: request is too large` or `Request entity too large` occur. This is an example of how reports are split for an image with 4499 vulnerabilities:
  ```shell
  kubectl get vulns -l zora.undistro.io/name=kind-python37-b8z6h -n zora-system
  NAME                      CLUSTER   IMAGE        TOTAL   CRITICAL   HIGH   AGE
  kind-python37-b8z6h-1-1   kind      python:3.7   1124    4          256    20h
  kind-python37-b8z6h-1-2   kind      python:3.7   1125    3          287    20h
  kind-python37-b8z6h-2-1   kind      python:3.7   1125    7          279    20h
  kind-python37-b8z6h-2-2   kind      python:3.7   1125    8          262    20h
  ```
- Vulnerability descriptions are automatically truncated to 300 characters to reduce the payload size.


You can also further reduce report size using the following configurations:

| Helm Parameter                                                                | Description                                           |
|-------------------------------------------------------------------------------|-------------------------------------------------------|
| `--set scan.plugins.trivy.ignoreUnfixed=true`                                 | Ignore unfixed vulnerabilities                        |
| `--set scan.plugins.trivy.ignoreDescriptions=true`                            | Do not store vulnerability descriptions (only titles) |
| `--set scan.plugins.trivy.args="--exclude-namespaces kube-system\,openshift"` | Indicate namespaces excluded from scanning            |


## Scan timeout

Trivy's scan duration may vary depending on the total images in your cluster 
and the time to download the vulnerability database when needed. 

By default, Zora sets a timeout of **40 minutes** for Trivy scan completion.

To adjust this timeout, use the following Helm parameter:

```shell
--set scan.plugins.trivy.timeout=60m
```

Once this parameter is updated, the next scan will use the specified value.
