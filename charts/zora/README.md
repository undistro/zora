# Zora Helm Chart

![Version: 0.4.0](https://img.shields.io/badge/Version-0.4.0-informational?style=flat-square&color=3CA9DD) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square&color=3CA9DD) ![AppVersion: v0.4.0](https://img.shields.io/badge/AppVersion-v0.4.0-informational?style=flat-square&color=3CA9DD)

Zora scans multiple Kubernetes clusters and reports potential issues.

## Installing the Chart

To install the chart with the release name `zora`:

```console
helm repo add undistro https://charts.undistro.io --force-update
helm upgrade --install zora undistro/zora \
  -n zora-system \
  --version 0.4.0 \
  --create-namespace --wait
```

> The Helm chart repository has been updated from `https://registry.undistro.io/chartrepo/library` to `https://charts.undistro.io`.
>
> The `--force-update` flag is needed to update the repository URL.

These commands deploy Zora on the Kubernetes cluster in the default configuration.

The [Parameters](#parameters) section lists the parameters that can be configured during installation.

> **Tips:**
>
> - List all charts available in `undistro` repo using `helm search repo undistro`
>
> - Update `undistro` chart repository using `helm repo update undistro`
>
> - List all versions available of `undistro/zora` chart using `helm search repo undistro/zora --versions`
>
> - List all releases using `helm list`
>
> - Get the notes provided by `zora` release using `helm get notes zora -n zora-system`

## Uninstalling the Chart

To uninstall/delete the `zora` release:

```console
$ helm delete zora
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

The following table lists the configurable parameters of the Zora chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | String to partially override fullname template with a string (will prepend the release name) |
| fullnameOverride | string | `""` | String to fully override fullname template with a string |
| saas.workspaceID | string | `""` | Your SaaS workspace ID |
| saas.server | string | `"https://zora-dashboard.undistro.io"` | SaaS server URL |
| saas.hooks.image.repository | string | `"radial/busyboxplus"` | SaaS hooks image repository |
| saas.hooks.image.tag | string | `"curl"` | SaaS hooks image tag |
| saas.hooks.installURL | string | `"{{.Values.saas.server}}/zora/api/v1alpha1/workspaces/{{.Values.saas.workspaceID}}/helmreleases"` | SaaS install hook URL template |
| imageCredentials.create | bool | `false` | Specifies whether the secret should be created by providing credentials |
| imageCredentials.registry | string | `"ghcr.io"` | Docker registry host |
| imageCredentials.username | string | `""` | Docker registry username |
| imageCredentials.password | string | `""` | Docker registry password |
| imagePullSecrets | list | `[]` | Specify docker-registry secret names as an array to be used when `imageCredentials.create` is false |
| operator.replicaCount | int | `1` | Number of replicas desired of Zora operator |
| operator.image.repository | string | `"ghcr.io/undistro/zora/operator"` | Zora operator image repository |
| operator.image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion |
| operator.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| operator.rbac.create | bool | `true` | Specifies whether ClusterRoles and ClusterRoleBindings should be created |
| operator.rbac.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| operator.rbac.serviceAccount.annotations | object | `{}` | Annotations to be added to service account |
| operator.rbac.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| operator.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"manager"}` | Annotations to be added to pods |
| operator.podSecurityContext | object | `{"runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to the pod |
| operator.securityContext | object | `{"allowPrivilegeEscalation":false,"readOnlyRootFilesystem":true}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to `manager` container |
| operator.metricsService.type | string | `"ClusterIP"` | Type of metrics service |
| operator.metricsService.port | int | `8443` | Port of metrics service |
| operator.serviceMonitor.enabled | bool | `false` | Specifies whether a Prometheus `ServiceMonitor` should be enabled |
| operator.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"10m","memory":"64Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `manager` container |
| operator.rbacProxy.image.repository | string | `"gcr.io/kubebuilder/kube-rbac-proxy"` | `kube-rbac-proxy` image repository |
| operator.rbacProxy.image.tag | string | `"v0.8.0"` | `kube-rbac-proxy` image tag |
| operator.rbacProxy.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| operator.rbacProxy.securityContext | object | `{"allowPrivilegeEscalation":false,"readOnlyRootFilesystem":true}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to `kube-rbac-proxy` container |
| operator.rbacProxy.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"5m","memory":"64Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `kube-rbac-proxy` container |
| operator.nodeSelector | object | `{}` | [Node selection](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node) to constrain a Pod to only be able to run on particular Node(s) |
| operator.tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) for pod assignment |
| operator.affinity | object | `{}` | Map of node/pod [affinities](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) |
| operator.log.encoding | string | `"json"` | Log encoding (one of 'json' or 'console') |
| operator.log.level | string | `"info"` | Log level to configure the verbosity of logging. Can be one of 'debug', 'info', 'error', or any integer value > 0 which corresponds to custom debug levels of increasing verbosity |
| operator.log.stacktraceLevel | string | `"error"` | Log level at and above which stacktraces are captured (one of 'info', 'error' or 'panic') |
| operator.log.timeEncoding | string | `"rfc3339"` | Log time encoding (one of 'epoch', 'millis', 'nano', 'iso8601', 'rfc3339' or 'rfc3339nano') |
| scan.worker.image.repository | string | `"ghcr.io/undistro/zora/worker"` | worker image repository |
| scan.worker.image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion |
| scan.defaultPlugins | list | `["popeye"]` | Names of the default plugins |
| scan.plugins.popeye.enabled | bool | `true` |  |
| scan.plugins.popeye.image.repository | string | `"ghcr.io/undistro/popeye"` | popeye plugin image repository |
| scan.plugins.popeye.image.tag | string | `"v0.10.2"` | popeye plugin image tag |
| scan.plugins.kubescape.enabled | bool | `false` |  |
| scan.plugins.kubescape.image.repository | string | `"quay.io/armosec/kubescape"` | kubescape plugin image repository |
| scan.plugins.kubescape.image.tag | string | `"v2.0.163"` | kubescape plugin image tag |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install zora \
  --set server.service.port=8080 undistro/zora
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```console
$ helm install zora -f values.yaml undistro/zora
```

> **Tip**: You can use the default [values.yaml](values.yaml)
