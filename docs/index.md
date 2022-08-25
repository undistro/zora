# Introduction

## What is Zora?

Zora is a multi-cluster scan that helps you to identify potential issues and vulnerabilities 
in your Kubernetes clusters in a centralized way, ensuring that the recommended best practices are in place.

Throughout this documentation, we will use the following notation:
- "Management Cluster" to refer to the only Kubernetes cluster where Zora is installed;
- "Target Cluster" to refer to all clusters you will connect to Zora to be scanned. The Target clusters will be running on the Management Cluster.

Follow these steps to get started with Zora:

1. [Install Zora](/install) in a [Management Cluster](/glossary#management-cluster)

2. [Prepare the target cluster](/target-cluster) by creating a service account and generating a kubeconfig

3. [Connect the target cluster to Zora](/connect-cluster)

4. [Configure a scan for the target cluster](/cluster-scan)

5. After a successful scan [checkout the potential reported issues](/cluster-scan#list-cluster-issues)

All the information about these steps are detailed throughout this documentation.
