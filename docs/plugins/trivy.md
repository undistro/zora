---
title: Trivy Plugin 
---

# Trivy Plugin

Trivy is a versatile security scanner that can find **vulnerabilities**, misconfigurations, secrets, SBOM 
in different targets like containers, code repositories and **Kubernetes cluster**.

**Zora uses Trivy as a plugin exclusively to scan vulnerabilities in a Kubernetes cluster.**

:octicons-codescan-24: **Type**: `vulnerability`

:simple-docker: **Image**: `ghcr.io/undistro/trivy:0.49.1-3`

:simple-github: **GitHub repository**: [https://github.com/aquasecurity/trivy](https://github.com/aquasecurity/trivy){:target="_blank"}

## Large vulnerability reports

Vulnerability reports can be large.
If you encounter issues with etcd request payload limit, you can ignore unfixed vulnerabilities from reports 
by providing the following flag to `helm upgrade --install` command:

```
--set 'scan.plugins.trivy.ignoreUnfixed=true'
```

To identify this issue, check the logs of worker container in trivy pod.
The `ClusterScan` will have a `Failed` status. You will see a log entry similar to the following example:

```
2023-09-26T14:18:02Z	ERROR	worker	failed to run worker	{"error": "failed to create VulnerabilityReport \"kind-kind-usdockerpkgdevgooglesamplescontainersgkegbfrontendsha256dc8de8e0d569d2f828b187528c9317bd6b605c273ac5a282aebe471f630420fc-rzntw\": etcdserver: request is too large"}
```

## Scan timeout

Trivy's scan duration may vary depending on the total images in your cluster 
and the time to download the vulnerability database during each scan. 

By default, Zora sets a timeout of **10 minutes** for Trivy scan completion.

To adjust this timeout, use the following Helm parameter:

```shell
--set scan.plugins.trivy.timeout=15m
```

Once this parameter is updated, the next scan will use the specified value.
