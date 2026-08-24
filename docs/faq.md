---
title: FAQ 
---

# Frequently Asked Questions

Do you have any question about Zora?
We do our best to answer all of your questions on this page. 
If you can't find your question below, 
ask it on our [discussion board](https://github.com/undistro/zora/discussions/categories/q-a){:target="_blank"}!

## Is Zora open source?

There are two components documented here: Zora OSS and the now-discontinued Zora Dashboard.

[Zora OSS is open-source](https://github.com/undistro/zora){:target="_blank"}, available under Apache 2.0 license, 
and is used as a standalone tool. Zora Dashboard was discontinued on August 24, 2026.

## Can I use Zora OSS standalone?

Yes, you can use Zora OSS as a standalone tool and access scan results (misconfigurations and vulnerabilities) 
via `kubectl` one cluster at a time.

## Can I install Zora in a different namespace?

Yes, Zora can be installed in any namespace. 
Simply provide the namespace name using the `-n` flag in [Helm installation command](getting-started/installation.md).

The `Cluster`, `ClusterScan`, `Plugin`, `ClusterIssue`, and `VulnerabilityReport` objects 
will be created in the specified namespace.

If you already have Zora installed and want to change the namespace, you will need to reinstall it.

## Can I integrate my own plugins with Zora, and how?

Currently, integrating a new plugin into Zora requires modifying the source code of Worker, a Zora component.
The parsing of plugin results into `ClusterIssue` or `VulnerabilityReport` is directly handled by Worker, 
which is written in Go. A fully declarative approach is not yet supported.

Refer to [plugins page](plugins/index.md) to know more about how plugins work.

Feel free to [open an issue](https://github.com/undistro/zora/issues/new/choose){:target="_blank"} or 
[start a discussion](https://github.com/undistro/zora/discussions/categories/q-a){:target="_blank"} with any suggestions 
regarding this process.

## What happens to data collected by Zora OSS?

Scan results remain in the cluster where Zora OSS is installed and can be accessed through the Kubernetes API or
`kubectl`. For example, use `kubectl get vulnerabilities -n zora-system` and
`kubectl get misconfigurations -n zora-system`.

## Is there an on-premise replacement for Zora Dashboard?

No. Zora Dashboard was discontinued on August 24, 2026. Zora OSS remains available as open source and runs in
your own Kubernetes cluster.
