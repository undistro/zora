---
title: FAQ 
---

# Frequently Asked Questions

Do you have any question about Zora?
We do our best to answer all of your questions on this page. 
If you can't find your question below, 
ask it on our [discussion board](https://github.com/undistro/zora/discussions/categories/q-a)!

## Is Zora open source?

There are two Zora tools: Zora OSS and Zora Dashboard.

[Zora OSS is open-source](https://github.com/undistro/zora), available under Apache 2.0 license, 
and can be used either as standalone tool or integrated with Zora Dashboard.

On the other hand, Zora Dashboard is a SaaS platform that provides a full experience, 
centralizing the security posture management of all your clusters.

It's free for up to 3 clusters. Visit the [Zora Dashboard page](dashboard.md) for more information.

## Can I use Zora OSS standalone without Zora Dashboard?

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

Feel free to [open an issue](https://github.com/undistro/zora/issues/new/choose) or 
[start a discussion](https://github.com/undistro/zora/discussions/categories/q-a) with any suggestions 
regarding this process.

## Which data is sent to Zora Dashboard (SaaS)?

When integrated with Zora Dashboard, **only scan results are sent to the SaaS** platform. 

**No sensitive information is collected or exposed**. 

Scans are performed in your cluster and the results are securely sent via HTTPS to Zora Dashboard, 
where only you and the users you've invited to your workspace will have access.

## Can I host Zora Dashboard on-premise?

Currently, Zora Dashboard is available as a SaaS platform. 
While we do not offer an on-premise version of Zora Dashboard at this time, we're continuously working to enhance and 
expand our offerings. If you have specific requirements or are interested in on-premise solutions, 
please [contact us](https://undistro.io/contact), and we'll be happy to discuss potential options and 
explore how we can meet your needs.
