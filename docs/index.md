# Welcome to the Zora documentation

This documentation will help you install, explore, and configure Zora!

## What is Zora?

Zora is an open-source solution that helps you achieve compliance with Kubernetes best practices recommended by 
industry-leading frameworks.

By scanning your cluster with multiple plugins at scheduled times, 
Zora identifies potential issues, misconfigurations, and vulnerabilities.

## Zora OSS vs Zora Dashboard

[Zora OSS is open-source](https://github.com/undistro/zora), available under Apache 2.0 license,
and can be used either as standalone tool or integrated with [Zora Dashboard](dashboard.md), 
a SaaS platform which centralize all your clusters providing a full experience. 
Please refer to [Zora Dashboard page](dashboard.md) for more details.

## Key features

#### :octicons-plug-16: Multi-plugin architecture
  
Zora seamlessly integrates open-source tools like 
[Popeye](plugins/popeye.md), 
[Marvin](plugins/marvin.md), 
and [Trivy](plugins/trivy.md) as scanners. 
These tools' capabilities are combined to provide you with a unified view of your cluster's security posture, 
addressing potential issues, misconfigurations, and vulnerabilities.

#### :fontawesome-solid-list-check: Kubernetes compliance

Zora and its plugins provide actionable insights, guiding you to align your cluster with industry-recognized frameworks 
such as 
[NSA-CISA](https://media.defense.gov/2022/Aug/29/2003066362/-1/-1/0/CTR_KUBERNETES_HARDENING_GUIDANCE_1.2_20220829.PDF), 
[MITRE ATT&CK](https://microsoft.github.io/Threat-Matrix-for-Kubernetes), 
[CIS Benchmark](https://downloads.cisecurity.org), 
and [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards).

#### :octicons-sliders-16: Custom checks

Enabled by the [Marvin](https://github.com/undistro/marvin) plugin, Zora offers a declarative way to create your own 
checks by using [CEL](https://github.com/google/cel-spec) expressions to define validation rules.

#### :simple-kubernetes: Kubernetes-native

All scan configurations and plugin reports, including misconfigurations and vulnerabilities,
are securely stored as [CRDs (Custom Resource Definitions)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
within your Kubernetes cluster, making it easily accessible through the Kubernetes API and `kubectl` command.

## Architecture

Zora works as a [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/), 
where both scan and plugin configurations, as well as the results (misconfigurations and vulnerabilities), 
are managed in CRDs (Custom Resource Definitions).

![Zora architecture diagram](assets/oss-arch-light.png#only-light){ loading=lazy }
![Zora architecture diagram](assets/oss-arch-dark.png#only-dark){ loading=lazy }

## Zora origins

In the early days of the cloud native era, [Borg](https://intl.startrek.com/database_article/borg) 
dominated the container-oriented cluster management scene.
The origin of the name Borg refers to the cybernetic life form existing in the Star Trek series,
that worked as a collective of individuals with a single mind and the same purpose, as well as a "[cluster](https://pt.wikipedia.org/wiki/Cluster)".

As good nerds as we are and wishing to honor our Kubernetes' 
[predecessor](https://kubernetes.io/blog/2015/04/borg-predecessor-to-kubernetes/) (Borg) we named our project
[Zora](https://intl.startrek.com/node/15372).

In Star Trek, Zora is the Artificial Intelligence that controls the ship U.S.S Discovery.  
After being merged with a collective of other intelligences, Zora became sentient and became a member of the team, 
bringing insights and making the ship more efficient.

Like Star Trek's Zora, our goal is to help manage your Kubernetes environment by combining multiple plugin capabilities to
scan your clusters looking for misconfigurations and vulnerabilities.
