# Zora Plugins

## Overview

Zora utilizes open-source CLI tools like
[Marvin](marvin.md),
[Popeye](popeye.md), 
and [Trivy](trivy.md) 
as plugins to perform scans on Kubernetes clusters.

The current available plugins of a Zora installation can be listed by running the following command:

```shell
kubectl get plugins -n zora-system
```
```
NAME     IMAGE                               TYPE               AGE
marvin   ghcr.io/undistro/marvin:v0.2.1      misconfiguration   14m
popeye   ghcr.io/undistro/popeye:0.21.2-1    misconfiguration   14m
trivy    ghcr.io/undistro/trivy:0.50.0-1     vulnerability      14m
```

Each item listed above is an instance of `Plugin` CRD and represents the execution configuration of a plugin.
More details can be seen by getting the YAML output of a plugin: 

```shell
kubectl get plugin marvin -o yaml -n zora-system
```

## Plugin types

Currently, Zora has two plugin types: `vulnerability` and `misconfiguration`, 
which determine the focus of plugin scans.

- `vulnerability` plugins scan cluster images for vulnerabilities, 
  and their results are stored as instances of `VulnerabilityReport` CRD.

- `misconfiguration` plugins scan cluster resources for potential configuration issues, 
  and their results are available as instances of the `ClusterIssue` CRD.

Both result types can be listed using `kubectl`, and some aliases are supported for your convenience, 
as shown in the following commands:

```shell

kubectl get vulnerabilityreports
kubectl get vuln
kubectl get vulns
kubectl get vulnerabilities
```
```shell
kubectl get clusterissues
kubectl get issue
kubectl get issues
kubectl get misconfig
kubectl get misconfigs
kubectl get misconfigurations
```

!!! note
    The results are only available after a successful scan, in the same namespace as the `ClusterScan` (default is `zora-system`).

## How plugins work

Starting from a `Plugin` and a `ClusterScan`, Zora manages and schedules scans by applying `CronJobs`, which
creates `Jobs` and `Pods`.

The `Pods` where the scans run, include a "sidecar" container called **worker** alongside the plugin container.

After the plugin completes its scan, it needs to signal to Zora (worker) by writing out the path of the results file 
into a "done file".

Worker container waits for the "done file" to be present, 
then transforms the results and creates `ClusterIssues` and `VulnerabilityReports` (depending on the plugin type).

!!! note
    This is the aspect that currently prevents the full declarative integration of new plugins. 
    The code responsible for transforming the output of each plugin into CRDs is written in Go within the worker.

    Any contributions or suggestions in this regard are greatly appreciated.

![Zora plugin diagram](../assets/plugin-arch-light.png#only-light)
![Zora plugin diagram](../assets/plugin-arch-dark.png#only-dark)

!!! note
    This architecture for supporting plugins is inspired by [Sonobuoy](https://sonobuoy.io/){:target="_blank"}, 
    a project used for CNCF conformance certification.
