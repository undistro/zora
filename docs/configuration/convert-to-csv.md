# Convert results to CSV

After a successful scan, the results (vulnerabilities and misconfigurations) are available in your cluster via CRDs, 
and you can transform them into CSV files using [jq](https://github.com/jqlang/jq){:target="_blank"}.

## Vulnerabilities

Vulnerability scan results are stored as instances of `VulnerabilityReport` CRD within your cluster. 
You can export summaries or detailed reports of these vulnerabilities to CSV format for further analysis.

### Images summary

To generate a summary report of vulnerabilities by image, run the following command:

```shell
kubectl get vulnerabilityreports -n zora-system -o json | jq -r '
  ["Image", "Image digest", "OS", "Distro", "Distro version", "Total", "Critical", "High", "Medium", "Low", "Unknown", "Scanned at"],
  (.items[] | [
    .spec.image, .spec.digest, .spec.os, .spec.distro.name, .spec.distro.version, 
    .spec.summary.total, .spec.summary.critical, .spec.summary.high, .spec.summary.medium, .spec.summary.low, .spec.summary.unknown,
    .metadata.creationTimestamp
  ]) | @csv' > images.csv
```

This command will produce a CSV file, `images.csv`, with the following structure:

| Image                                               | Image digest                                                                                               | OS    | Distro | Distro version | Total | Critical | High | Medium | Low | Unknown | Scanned at           |
|-----------------------------------------------------|------------------------------------------------------------------------------------------------------------|-------|--------|----------------|-------|----------|------|--------|-----|---------|----------------------|
| docker.io/istio/examples-bookinfo-reviews-v1:1.20.1 | istio/examples-bookinfo-reviews-v1@sha256:5b3c8ec2cb877b7a3c93fc340bb91633c3e51a6bc43a2da3ae7d72727650ec07 | linux | ubuntu | 22.04          | 45    | 0        | 0    | 25     | 20  | 0       | 2024-10-31T12:56:51Z |
| nginx                                               | nginx@sha256:28402db69fec7c17e179ea87882667f1e054391138f77ffaf0c3eb388efc3ffb                              | linux | debian | 12.7           | 95    | 2        | 10   | 24     | 59  | 0       | 2024-10-31T12:56:51Z |

### Full report: images and vulnerabilities

To create a detailed report of each vulnerability affecting images, use the following command:

```shell
kubectl get vulnerabilityreports -n zora-system -o json | jq -r '
  ["Image", "Image digest", "OS", "Distro", "Distro version", "Vulnerability ID", "Severity", "Score", "Title", "Package", "Type", "Version", "Status", "Fix version", "Scanned at"],
  (.items[] | . as $i | $i.spec.vulnerabilities[] as $vuln | $vuln.packages[] | [
    $i.spec.image, $i.spec.digest, $i.spec.os, $i.spec.distro.name, $i.spec.distro.version,
    $vuln.id, $vuln.severity, $vuln.score, $vuln.title,
    .package, .type, .version, .status, .fixVersion,
    $i.metadata.creationTimestamp
  ]) | @csv' > vulnerabilities.csv
```

This will generate a `vulnerabilities.csv` file with details for each vulnerability:

!!! note
    A single vulnerability can affect multiple packages within the same image,
    so you may see repeated entries for the same vulnerability. 
    For instance, in the example below, `CVE-2024-7264` affects both `curl` and `libcurl4` packages in the same image.

| Image                                               | Image digest                                                                                               | OS    | Distro | Distro version | Vulnerability ID | Severity | Score | Title                                                                      | Package  | Type   | Version            | Status | Fix version        | Scanned at           |
|-----------------------------------------------------|------------------------------------------------------------------------------------------------------------|-------|--------|----------------|------------------|----------|-------|----------------------------------------------------------------------------|----------|--------|--------------------|--------|--------------------|----------------------|
| nginx                                               | nginx@sha256:28402db69fec7c17e179ea87882667f1e054391138f77ffaf0c3eb388efc3ffb                              | linux | debian | 12.7           | CVE-2023-49462   | HIGH     | 8.8   | libheif v1.17.5 was discovered to contain a segmentation violation via ... | libheif1 | debian | 1.15.1-1           | fixed  | 1.15.1-1+deb12u1   | 2024-10-31T12:56:51Z |
| docker.io/istio/examples-bookinfo-reviews-v1:1.20.1 | istio/examples-bookinfo-reviews-v1@sha256:5b3c8ec2cb877b7a3c93fc340bb91633c3e51a6bc43a2da3ae7d72727650ec07 | linux | ubuntu | 22.04          | CVE-2024-7264    | MEDIUM   | 6.5   | curl: libcurl: ASN.1 date parser overread                                  | curl     | ubuntu | 7.81.0-1ubuntu1.15 | fixed  | 7.81.0-1ubuntu1.17 | 2024-10-31T12:56:51Z |
| docker.io/istio/examples-bookinfo-reviews-v1:1.20.1 | istio/examples-bookinfo-reviews-v1@sha256:5b3c8ec2cb877b7a3c93fc340bb91633c3e51a6bc43a2da3ae7d72727650ec07 | linux | ubuntu | 22.04          | CVE-2024-7264    | MEDIUM   | 6.5   | curl: libcurl: ASN.1 date parser overread                                  | libcurl4 | ubuntu | 7.81.0-1ubuntu1.15 | fixed  | 7.81.0-1ubuntu1.17 | 2024-10-31T12:56:51Z |

## Misconfigurations

Misconfiguration scan results are represented as instances of `ClusterIssue` CRD within your cluster, 
and can also be parsed to CSV format.

### Misconfigurations summary

To generate a summary report of misconfigurations, you can run the following command:

```shell
kubectl get misconfigurations -n zora-system -o json | jq -r '
  ["ID", "Misconfiguration", "Severity", "Category", "Total resources", "Scanned at"],
  (.items[] | ([.spec.resources[] | length] | add) as $totalResources | [
    .spec.id, .spec.message, .spec.severity, .spec.category, $totalResources, .metadata.creationTimestamp
  ]) | @csv' > misconfigurations.csv
```

This command will create a `misconfigurations.csv` file with the following structure:

| ID    | Misconfiguration                                      | Severity | Category       | Total resources | Scanned at           |
|-------|-------------------------------------------------------|----------|----------------|-----------------|----------------------|
| M-102 | Privileged container                                  | High     | Security       | 2               | 2024-10-31T17:45:08Z |
| M-103 | Insecure capabilities                                 | High     | Security       | 2               | 2024-10-31T17:45:08Z |
| M-112 | Allowed privilege escalation                          | Medium   | Security       | 14              | 2024-10-31T17:45:08Z |
| M-113 | Container could be running as root user               | Medium   | Security       | 18              | 2024-10-31T17:45:08Z |
| M-201 | Application credentials stored in configuration files | High     | Security       | 6               | 2024-10-31T17:45:08Z |
| M-300 | Root filesystem write allowed                         | Low      | Security       | 29              | 2024-10-31T17:45:08Z |
| M-400 | Image tagged latest                                   | Medium   | Best Practices | 2               | 2024-10-31T17:45:08Z |
| M-403 | Liveness probe not configured                         | Medium   | Reliability    | 16              | 2024-10-31T17:45:08Z |
| M-406 | Memory not limited                                    | Medium   | Reliability    | 15              | 2024-10-31T17:45:08Z |

### Full report: misconfigurations and affected resources

A detailed CSV file containing the affected resources can be generated with the command below.

```shell
kubectl get misconfigurations -n zora-system -o json | jq -r '
  ["ID", "Misconfiguration", "Severity", "Category", "Resource Type", "Resource", "Scanned at"],
  (.items[] as $i | $i.spec.resources | to_entries[] as $resource | $resource.value[] as $affectedResource | [
    $i.spec.id, $i.spec.message, $i.spec.severity, $i.spec.category, $resource.key, $affectedResource, $i.metadata.creationTimestamp
  ]) | @csv' > misconfigurations_full.csv
```

This command will generate the `misconfigurations_full.csv` file with the following structure:

| ID    | Misconfiguration    | Severity | Category       | Resource Type | Resource      | Scanned at           |
|-------|---------------------|----------|----------------|---------------|---------------|----------------------|
| M-400 | Image tagged latest | Medium   | Best Practices | v1/pods       | default/test  | 2024-10-31T18:45:06Z |
| M-400 | Image tagged latest | Medium   | Best Practices | v1/pods       | default/nginx | 2024-10-31T18:45:06Z |

