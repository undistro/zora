# Zora Helm Chart

![Version: 0.8.1-rc1](https://img.shields.io/badge/Version-0.8.1--rc1-informational?style=flat-square&color=3CA9DD) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square&color=3CA9DD) ![AppVersion: v0.8.1-rc1](https://img.shields.io/badge/AppVersion-v0.8.1--rc1-informational?style=flat-square&color=3CA9DD)

A multi-plugin solution that reports misconfigurations and vulnerabilities by scanning your cluster at scheduled times.

## Installing the Chart

To install the chart with the release name `zora` in `zora-system` namespace:

```console
helm repo add undistro https://charts.undistro.io --force-update
helm repo update undistro
helm upgrade --install zora undistro/zora \
  -n zora-system \
  --version 0.8.1-rc1 \
  --create-namespace \
  --wait \
  --set clusterName="$(kubectl config current-context)"
```

These commands deploy Zora on the Kubernetes cluster with the default configuration.

The [Parameters](#parameters) section lists the available parameters that can be configured during installation.

> **Tips:**
>
> - List all charts available in `undistro` repo using `helm search repo undistro`
>
> - Update `undistro` chart repository using `helm repo update undistro`
>
> - List all versions available of `undistro/zora` chart using `helm search repo undistro/zora --versions`
>
> - List all releases in a specific namespace using `helm list -n zora-system`
>
> - Get the notes provided by `zora` release using `helm get notes zora -n zora-system`

## Uninstalling the Chart

To uninstall/delete the `zora` release:

```console
helm uninstall zora -n zora-system
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

The following table lists the configurable parameters of the Zora chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | String to partially override fullname template with a string (will prepend the release name) |
| fullnameOverride | string | `""` | String to fully override fullname template with a string |
| clusterName | string | `""` | Cluster name. Should be set by `kubectl config current-context`. |
| saas.workspaceID | string | `""` | Your SaaS workspace ID |
| saas.server | string | `"https://zora-dashboard.undistro.io"` | SaaS server URL |
| saas.installURL | string | `"{{.Values.saas.server}}/zora/api/v1alpha1/workspaces/{{.Values.saas.workspaceID}}/helmreleases"` | SaaS URL template to notify installation |
| hooks.install.image.repository | string | `"curlimages/curl"` | Install hook image repository |
| hooks.install.image.tag | string | `"8.2.1"` | Install hook image tag |
| hooks.delete.image.repository | string | `"rancher/kubectl"` | Delete hook image repository |
| hooks.delete.image.tag | string | `"v1.28.2"` | Delete hook image tag |
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
| operator.rbacProxy.image.tag | string | `"v0.15.0"` | `kube-rbac-proxy` image tag |
| operator.rbacProxy.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| operator.rbacProxy.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true}` | [Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context) to add to `kube-rbac-proxy` container |
| operator.rbacProxy.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"5m","memory":"64Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `kube-rbac-proxy` container |
| operator.nodeSelector | object | `{}` | [Node selection](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node) to constrain a Pod to only be able to run on particular Node(s) |
| operator.tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) for pod assignment |
| operator.affinity | object | `{}` | Map of node/pod [affinities](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration) |
| operator.log.encoding | string | `"json"` | Log encoding (one of 'json' or 'console') |
| operator.log.level | string | `"info"` | Log level to configure the verbosity of logging. Can be one of 'debug', 'info', 'error', or any integer value > 0 which corresponds to custom debug levels of increasing verbosity |
| operator.log.stacktraceLevel | string | `"error"` | Log level at and above which stacktraces are captured (one of 'info', 'error' or 'panic') |
| operator.log.timeEncoding | string | `"rfc3339"` | Log time encoding (one of 'epoch', 'millis', 'nano', 'iso8601', 'rfc3339' or 'rfc3339nano') |
| scan.misconfiguration.enabled | bool | `true` | Specifies whether misconfiguration scan is enabled |
| scan.misconfiguration.schedule | string | Cron expression for every hour at the current minute + 5 minutes | Cluster scan schedule in Cron format for misconfiguration scan |
| scan.misconfiguration.successfulScansHistoryLimit | int | `1` | The number of successful finished scans and their issues to retain. |
| scan.misconfiguration.plugins | list | `["marvin","popeye"]` | Misconfiguration scanners plugins |
| scan.vulnerability.enabled | bool | `true` | Specifies whether vulnerability scan is enabled |
| scan.vulnerability.schedule | string | Cron expression for every day at the current hour and minute + 5 minutes | Cluster scan schedule in Cron format for vulnerability scan |
| scan.vulnerability.successfulScansHistoryLimit | int | `1` | The number of successful finished scans and their issues to retain. |
| scan.vulnerability.plugins | list | `["trivy"]` | Vulnerability scanners plugins |
| scan.worker.image.repository | string | `"ghcr.io/undistro/zora/worker"` | worker image repository |
| scan.worker.image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion |
| scan.plugins.annotations | object | `{}` | Annotations added to the plugin service account |
| scan.plugins.marvin.resources | object | `{"limits":{"cpu":"500m","memory":"500Mi"},"requests":{"cpu":"250m","memory":"256Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `marvin` container |
| scan.plugins.marvin.podAnnotations | object | `{}` | Annotations added to the marvin pods |
| scan.plugins.marvin.image.repository | string | `"ghcr.io/undistro/marvin"` | marvin plugin image repository |
| scan.plugins.marvin.image.tag | string | `"v0.2.1"` | marvin plugin image tag |
| scan.plugins.trivy.ignoreUnfixed | bool | `false` | Specifies whether only fixed vulnerabilities should be reported |
| scan.plugins.trivy.ignoreDescriptions | bool | `false` | Specifies whether vulnerability descriptions should be ignored |
| scan.plugins.trivy.resources | object | `{}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `trivy` container |
| scan.plugins.trivy.podAnnotations | object | `{}` | Annotations added to the trivy pods |
| scan.plugins.trivy.image.repository | string | `"ghcr.io/aquasecurity/trivy"` | trivy plugin image repository |
| scan.plugins.trivy.image.tag | string | `"0.48.2"` | trivy plugin image tag |
| scan.plugins.popeye.skipInternalResources | bool | `false` | Specifies whether the following resources should be skipped by `popeye` scans. 1. resources from `kube-system`, `kube-public` and `kube-node-lease` namespaces; 2. kubernetes system reserved RBAC (prefixed with `system:`); 3. `kube-root-ca.crt` configmaps; 4. `default` namespace; 5. `default` serviceaccounts; 6. Helm secrets (prefixed with `sh.helm.release`); 7. Zora components. See `popeye` configuration file that is used for this case: https://github.com/undistro/zora/blob/main/charts/zora/templates/plugins/popeye-config.yaml |
| scan.plugins.popeye.resources | object | `{"limits":{"cpu":"500m","memory":"500Mi"},"requests":{"cpu":"250m","memory":"256Mi"}}` | [Resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) to add to `popeye` container |
| scan.plugins.popeye.podAnnotations | object | `{}` | Annotations added to the popeye pods |
| scan.plugins.popeye.image.repository | string | `"ghcr.io/undistro/popeye"` | popeye plugin image repository |
| scan.plugins.popeye.image.tag | string | `"v0.11.3"` | popeye plugin image tag |
| kubexnsImage.repository | string | `"ghcr.io/undistro/kubexns"` | kubexns image repository |
| kubexnsImage.tag | string | `"v0.1.2"` | kubexns image tag |
| customChecksConfigMap | string | `"zora-custom-checks"` | Custom checks ConfigMap name |
| httpsProxy | string | `""` | HTTPS proxy URL |
| noProxy | string | `"kubernetes.default.svc.*,127.0.0.1,localhost"` | Comma-separated list of URL patterns to be excluded from going through the proxy |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
helm install zora \
  --set operator.resources.limits.memory=256Mi undistro/zora
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```console
helm install zora -f values.yaml undistro/zora
```

> **Tip**: You can use the default [values.yaml](values.yaml)
