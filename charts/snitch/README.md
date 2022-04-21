# Snitch Helm Chart

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

Snitch denounces potential issues in your Kubernetes cluster and provides multi cluster visibility.

## Installing the Chart

To install the chart with the release name `snitch`:

```console
helm repo add snitch https://registry.undistro.io/chartrepo/snitch
helm install snitch snitch/snitch -n snitch --create-namespace
```

These commands deploy Snitch on the Kubernetes cluster in the default configuration.

The [Parameters](#parameters) section lists the parameters that can be configured during installation.

> **Tips:**
> - List all charts available in `snitch` repo using `helm search repo snitch`
> - List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `snitch` release:

```console
$ helm delete snitch
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

The following table lists the configurable parameters of the Snitch chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | String to partially override fullname template with a string (will prepend the release name) |
| fullnameOverride | string | `""` | String to fully override fullname template with a string |
| imagePullSecrets | list | `[]` | Specify docker-registry secret names as an array |
| operator.replicaCount | int | `1` | Number of replicas desired of Snitch operator |
| operator.image | string | `"registry.undistro.io/snitch/operator:v0.1.0"` |  |
| operator.imagePullPolicy | string | `"IfNotPresent"` | Image pull policy |
| operator.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| operator.serviceAccount.annotations | object | `{}` | Annotations to be added to service account |
| operator.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| operator.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"manager"}` | Annotations to be added to pods |
| operator.podSecurityContext | object | `{"runAsNonRoot":true}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the pod |
| operator.securityContext | object | `{"allowPrivilegeEscalation":false}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to `manager` container |
| operator.metricsService.type | string | `"ClusterIP"` | Type of metrics service |
| operator.metricsService.port | int | `8443` | Port of metrics service |
| operator.monitor.enabled | bool | `false` | Specifies whether prometheus monitor should be enabled |
| operator.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"10m","memory":"64Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `manager` container |
| operator.rbacProxy.image | string | `"registry.undistro.io/gcr/kubebuilder/kube-rbac-proxy:v0.8.0"` |  |
| operator.rbacProxy.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"5m","memory":"64Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `kube-rbac-proxy` container |
| operator.nodeSelector | object | `{}` | [Node selection](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node) to constrain a Pod to only be able to run on particular Node(s) |
| operator.tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) for pod assignment |
| operator.affinity | object | `{}` | Map of node/pod [affinities](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) |
| server.replicaCount | int | `1` | Number of replicas desired of Snitch server |
| server.image | string | `"registry.undistro.io/snitch/server:v0.1.0"` |  |
| server.imagePullPolicy | string | `"IfNotPresent"` | Image pull policy |
| server.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| server.serviceAccount.annotations | object | `{}` | Annotations to be added to service account |
| server.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| server.podAnnotations | object | `{}` | Annotations to be added to pods |
| server.podSecurityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the pod |
| server.securityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the container |
| server.service.type | string | `"ClusterIP"` | Service type |
| server.service.port | int | `80` | Service port |
| server.resources | object | `{}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to the container |
| server.autoscaling.enabled | bool | `false` | Enable replica autoscaling settings |
| server.autoscaling.minReplicas | int | `1` | Minimum replicas for the pod autoscaling |
| server.autoscaling.maxReplicas | int | `100` | Maximum replicas for the pod autoscaling |
| server.autoscaling.targetCPUUtilizationPercentage | int | `80` | Percentage of CPU to consider when autoscaling |
| server.autoscaling.targetMemoryUtilizationPercentage | string | `""` | Percentage of Memory to consider when autoscaling |
| server.nodeSelector | object | `{}` | [Node selection](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node) to constrain a Pod to only be able to run on particular Node(s) |
| server.tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) for pod assignment |
| server.affinity | object | `{}` | Map of node/pod [affinities](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) |
| ui.replicaCount | int | `1` | Number of replicas desired of Snitch UI |
| ui.image | string | `"registry.undistro.io/snitch/ui:v0.1.0"` |  |
| ui.imagePullPolicy | string | `"IfNotPresent"` | Image pull policy |
| ui.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| ui.serviceAccount.annotations | object | `{}` | Annotations to be added to service account |
| ui.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| ui.podAnnotations | object | `{}` | Annotations to be added to pods |
| ui.podSecurityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the pod |
| ui.securityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the container |
| ui.service.type | string | `"ClusterIP"` | Service type |
| ui.service.port | int | `80` | Service port |
| ui.resources | object | `{}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to the container |
| ui.autoscaling.enabled | bool | `false` | Enable replica autoscaling settings |
| ui.autoscaling.minReplicas | int | `1` | Minimum replicas for the pod autoscaling |
| ui.autoscaling.maxReplicas | int | `100` | Maximum replicas for the pod autoscaling |
| ui.autoscaling.targetCPUUtilizationPercentage | int | `80` | Percentage of CPU to consider when autoscaling |
| ui.autoscaling.targetMemoryUtilizationPercentage | string | `""` | Percentage of Memory to consider when autoscaling |
| ui.nodeSelector | object | `{}` | [Node selection](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node) to constrain a Pod to only be able to run on particular Node(s) |
| ui.tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) for pod assignment |
| ui.affinity | object | `{}` | Map of node/pod [affinities](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) |
| nginx.replicaCount | int | `1` | Number of replicas desired of nginx |
| nginx.image | string | `"registry.undistro.io/dockerhub/library/nginx:1.20.2"` |  |
| nginx.imagePullPolicy | string | `"IfNotPresent"` | Image pull policy |
| nginx.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| nginx.serviceAccount.annotations | object | `{}` | Annotations to be added to service account |
| nginx.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| nginx.podAnnotations | object | `{}` | Annotations to be added to pods |
| nginx.podSecurityContext | object | `{"fsGroup":10000,"runAsUser":10000}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the pod |
| nginx.securityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the container |
| nginx.service.type | string | `"ClusterIP"` | Service type |
| nginx.service.port | int | `80` | Service port |
| nginx.resources | object | `{}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to the container |
| nginx.autoscaling.enabled | bool | `false` | Enable replica autoscaling settings |
| nginx.autoscaling.minReplicas | int | `1` | Minimum replicas for the pod autoscaling |
| nginx.autoscaling.maxReplicas | int | `100` | Maximum replicas for the pod autoscaling |
| nginx.autoscaling.targetCPUUtilizationPercentage | int | `80` | Percentage of CPU to consider when autoscaling |
| nginx.autoscaling.targetMemoryUtilizationPercentage | string | `""` | Percentage of Memory to consider when autoscaling |
| nginx.nodeSelector | object | `{}` | [Node selection](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node) to constrain a Pod to only be able to run on particular Node(s) |
| nginx.tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) for pod assignment |
| nginx.affinity | object | `{}` | Map of node/pod [affinities](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) |
| ingress.enabled | bool | `false` | Specifies whether the ingress should be created |
| ingress.className | string | `""` | Ingress class name |
| ingress.annotations | object | `{}` | Annotations to be added to Ingress |
| ingress.hosts | list | `[{"host":"chart-example.local","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}]` | Configure ingress hosts |
| ingress.tls | list | `[]` | Ingress TLS configuration |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install snitch \
  --set server.service.port=8080 snitch/snitch
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```console
$ helm install snitch -f values.yaml snitch/snitch
```

> **Tip**: You can use the default [values.yaml](values.yaml)
