# Snitch Helm Chart

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.1.0](https://img.shields.io/badge/AppVersion-v0.1.0-informational?style=flat-square)

Snitch denounces potential issues in your Kubernetes cluster and provides multi cluster visibility.

## Installing the Chart

To install the chart with the release name `inspect`:

```console
helm repo add undistro https://registry.undistro.io/chartrepo/library
helm install inspect undistro/inspect \
  --set imageCredentials.username=<USERNAME> \
  --set imageCredentials.password=<PASSWORD> \
  -n undistro-inspect \
  --create-namespace
```

These commands deploy Snitch on the Kubernetes cluster in the default configuration.

The [Parameters](#parameters) section lists the parameters that can be configured during installation.

> **Tips:**
> - List all charts available in `undistro` repo using `helm search repo undistro`
> - List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `inspect` release:

```console
$ helm delete inspect
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

The following table lists the configurable parameters of the Snitch chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | String to partially override fullname template with a string (will prepend the release name) |
| fullnameOverride | string | `""` | String to fully override fullname template with a string |
| imageCredentials.create | bool | `true` | Specifies whether the secret should be created by providing credentials |
| imageCredentials.registry | string | `"registry.undistro.io"` | Docker registry host |
| imageCredentials.username | string | `""` | Docker registry username |
| imageCredentials.password | string | `""` | Docker registry password |
| imagePullSecrets | list | `[]` | Specify docker-registry secret names as an array to be used when `imageCredentials.create` is false |
| ingress.enabled | bool | `false` | Specifies whether the ingress should be created |
| ingress.className | string | `""` | Ingress class name |
| ingress.annotations | object | `{}` | Annotations to be added to ingress |
| ingress.host | string | `"inspect.domain"` | The host of Snitch in ingress rule |
| ingress.server | object | `{"path":"/api","pathType":"ImplementationSpecific"}` | `path` and `pathType` of API in ingress rule. `path` pattern may vary according ingress controller (`/api/*` for GCE, `/api/.*` for NCP) |
| ingress.ui | object | `{"path":"/","pathType":"ImplementationSpecific"}` | `path` and `pathType` of UI in ingress rule. `path` pattern may vary according ingress controller (`/*` for GCE, `/.*` for NCP) |
| ingress.tlsSecretName | string | `""` | The name of secret which contains keys named: `tls.crt` - the certificate; `tls.key` - the private key |
| operator.replicaCount | int | `1` | Number of replicas desired of Snitch operator |
| operator.image.repository | string | `"registry.undistro.io/inspect/operator"` | Snitch operator image repository |
| operator.image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion |
| operator.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| operator.rbac.create | bool | `true` | Specifies whether ClusterRoles and ClusterRoleBindings should be created |
| operator.rbac.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| operator.rbac.serviceAccount.annotations | object | `{}` | Annotations to be added to service account |
| operator.rbac.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| operator.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"manager"}` | Annotations to be added to pods |
| operator.podSecurityContext | object | `{"runAsNonRoot":true}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the pod |
| operator.securityContext | object | `{"allowPrivilegeEscalation":false}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to `manager` container |
| operator.metricsService.type | string | `"ClusterIP"` | Type of metrics service |
| operator.metricsService.port | int | `8443` | Port of metrics service |
| operator.serviceMonitor.enabled | bool | `false` | Specifies whether a Prometheus `ServiceMonitor` should be enabled |
| operator.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"10m","memory":"64Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `manager` container |
| operator.rbacProxy.image.repository | string | `"registry.undistro.io/gcr/kubebuilder/kube-rbac-proxy"` | `kube-rbac-proxy` image repository |
| operator.rbacProxy.image.tag | string | `"v0.8.0"` | `kube-rbac-proxy` image tag |
| operator.rbacProxy.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| operator.rbacProxy.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"5m","memory":"64Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `kube-rbac-proxy` container |
| operator.nodeSelector | object | `{}` | [Node selection](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node) to constrain a Pod to only be able to run on particular Node(s) |
| operator.tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) for pod assignment |
| operator.affinity | object | `{}` | Map of node/pod [affinities](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) |
| server.replicaCount | int | `1` | Number of replicas desired of Snitch server |
| server.image.repository | string | `"registry.undistro.io/inspect/server"` | Snitch server image repository |
| server.image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion |
| server.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| server.rbac.create | bool | `true` | Specifies whether ClusterRole and ClusterRoleBinding should be created |
| server.rbac.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| server.rbac.serviceAccount.annotations | object | `{}` | Annotations to be added to service account |
| server.rbac.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| server.podAnnotations | object | `{}` | Annotations to be added to pods |
| server.podSecurityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the pod |
| server.securityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the container |
| server.service.type | string | `"ClusterIP"` | Service type |
| server.service.port | int | `8080` | Service port |
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
| ui.image.repository | string | `"registry.undistro.io/inspect/ui"` | Snitch UI image repository |
| ui.image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion |
| ui.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| ui.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| ui.serviceAccount.annotations | object | `{}` | Annotations to be added to service account |
| ui.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| ui.podAnnotations | object | `{}` | Annotations to be added to pods |
| ui.podSecurityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the pod |
| ui.securityContext | object | `{}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the container |
| ui.service.type | string | `"ClusterIP"` | Service type |
| ui.service.port | int | `8080` | Service port |
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
| nginx.image.repository | string | `"registry.undistro.io/dockerhub/library/nginx:1.20.2"` | NGINX image repository |
| nginx.image.tag | string | `"1.20.2"` | NGINX image tag |
| nginx.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
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

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install inspect \
  --set server.service.port=8080 undistro/inspect
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```console
$ helm install inspect -f values.yaml undistro/inspect
```

> **Tip**: You can use the default [values.yaml](values.yaml)
