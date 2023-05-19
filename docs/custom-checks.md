# Custom checks

Zora offers a declarative way to create your own checks using the `CustomCheck` API.

Custom checks use the [Common Expression Language (CEL)](https://github.com/google/cel-spec) 
to declare the validation rules and are performed by the [Marvin](https://github.com/undistro/marvin) plugin, 
so it should be enabled in your cluster scans.

!!! info 
    Marvin is already a default plugin and enabled by default in cluster scans since Zora 0.5.0.

## `CustomCheck`

The example below represents a custom check that requires the labels `mycompany.com/squad` and `mycompany.com/component` 
to be present on `Pods`, `Deployments` and `Services`.

!!! example
    ```yaml
    apiVersion: zora.undistro.io/v1alpha1
    kind: CustomCheck
    metadata:
      name: mycheck
    spec:
      message: "Required labels"
      severity: Low
      category: Custom
      match:
        resources:
          - group: ""
            version: v1
            resource: pods
          - group: apps
            version: v1
            resource: deployments
          - group: ""
            version: v1
            resource: services
      params:
        requiredLabels:
          - mycompany.com/squad
          - mycompany.com/component
      validations:
        - expression: >
            has(object.metadata.labels) &&
            !object.metadata.labels.all(label,
              params.requiredLabels.all(
                req, req != label
              )
            )
          message: "Resource without required labels"
    ```

The `spec.match.resources` defines which resources will be checked by the expressions 
defined in `spec.validations.expression` as [Common Expression Language (CEL)](https://github.com/google/cel-spec).

If an expression evaluates to `false`, the check fails and a `ClusterIssue` is reported.

### Variables

The variables available in CEL expressions:

| Variable | Description                                   |
|:--------:|-----------------------------------------------|
| `object` | The object being scanned.                     |
| `params` | The parameter defined in `spec.params` field. |

If matches a `PodSpec`, the following useful variables are available:

| Variable        | Description                                                                     |
|-----------------|---------------------------------------------------------------------------------|
| `allContainers` | A list of all containers, including `initContainers` and `ephemeralContainers`. |
| `podMeta`       | The Pod `metadata`.                                                             |
| `podSpec`       | The Pod `spec`.                                                                 |

The following resources matches a `PodSpec`:

- `v1/pods`
- `v1/replicationcontrollers`
- `apps/v1/replicasets`
- `apps/v1/deployments`
- `apps/v1/statefulsets`
- `apps/v1/daemonsets`
- `batch/v1/jobs`
- `batch/v1/cronjobs`

### Apply a `CustomCheck`

Since you have a `CustomCheck` on a file, you can apply it with the command below.

```shell
kubectl apply -f check.yaml -n zora-system
```

### List custom checks

Once created, list the custom checks to see if it's ready.

```shell
kubectl get customchecks -n zora-system
```
```
NAME      MESSAGE           SEVERITY   READY
mycheck   Required labels   Low        True
```

The `READY` column indicates when the check has successfully compiled and is ready to be used in the next Marvin scan.

`ClusterIssues` reported by a custom check have are labeled `custom=true` and can be filtered by the following command:

```shell
kubectl get clusterissues -l custom=true
```
```
NAME                             CLUSTER     ID        MESSAGE           SEVERITY   CATEGORY   AGE
mycluster-mycheck-4edd75cb85a4   mycluster   mycheck   Required labels   Low        Custom     25s
```

### Examples

All Marvin checks are similar to the `CustomCheck` API. 
You can see them in the [`internal/builtins`](https://github.com/undistro/marvin/tree/main/internal/builtins) folder for examples.

Some examples of Marvin built-in checks expressions:

- [HostPath volumes must be forbidden](https://github.com/undistro/marvin/blob/main/internal/builtins/pss/baseline/M-104_host_path_volumes.yml)
  ```
  !has(podSpec.volumes) || podSpec.volumes.all(vol, !has(vol.hostPath))
  ```
- [Sharing the host namespaces must be disallowed](https://github.com/undistro/marvin/blob/main/internal/builtins/pss/baseline/M-101_host_namespaces.yml)
  ```
  (!has(podSpec.hostNetwork) || podSpec.hostNetwork == false) &&
  (!has(podSpec.hostPID) || podSpec.hostPID == false) &&
  (!has(podSpec.hostIPC) || podSpec.hostIPC == false)
  ```
- [Privileged Pods disable most security mechanisms and must be disallowed](https://github.com/undistro/marvin/blob/main/internal/builtins/pss/baseline/M-102_privileged_containers.yml)
  ```
  allContainers.all(container,
    !has(container.securityContext) ||
    !has(container.securityContext.privileged) ||
    container.securityContext.privileged == false)
  ```
- [HostPorts should be disallowed entirely (recommended) or restricted to a known list](https://github.com/undistro/marvin/blob/main/internal/builtins/pss/baseline/M-105_host_ports.yml)
  ```
  allContainers.all(container,
    !has(container.ports) ||
    container.ports.all(port,
      !has(port.hostPort) ||
      port.hostPort == 0 ||
      port.hostPort in params.allowedHostPorts
    )
  )
  ```

Marvin's checks and Zora's `CustomCheck` API are inspired in 
[Kubernetes ValidatingAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#validation-expression) API, 
introduced in version 1.26 as an alpha feature. 
Below, the table of [validation expression examples](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#validation-expression-examples) from Kubernetes documentation.

| Expression                                                                                   | Purpose                                                                                        |
|----------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------|
| `object.minReplicas <= object.replicas && object.replicas <= object.maxReplicas`             | Validate that the three fields defining replicas are ordered appropriately                     |
| `'Available' in object.stateCounts`                                                          | Validate that an entry with the 'Available' key exists in a map                                |
| `(size(object.list1) == 0) != (size(object.list2) == 0)`                                     | Validate that one of two lists is non-empty, but not both                                      |
| <code>!('MY_KEY' in object.map1) &#124;&#124; object['MY_KEY'].matches('^[a-zA-Z]*$')</code> | Validate the value of a map for a specific key, if it is in the map                            |
| `object.envars.filter(e, e.name == 'MY_ENV').all(e, e.value.matches('^[a-zA-Z]*$')`          | Validate the 'value' field of a listMap entry where key field 'name' is 'MY_ENV'               |
| `has(object.expired) && object.created + object.ttl < object.expired`                        | Validate that 'expired' date is after a 'create' date plus a 'ttl' duration                    |
| `object.health.startsWith('ok')`                                                             | Validate a 'health' string field has the prefix 'ok'                                           |
| `object.widgets.exists(w, w.key == 'x' && w.foo < 10)`                                       | Validate that the 'foo' property of a listMap item with a key 'x' is less than 10              |
| `type(object) == string ? object == '100%' : object == 1000`                                 | Validate an int-or-string field for both the int and string cases                              |
| `object.metadata.name.startsWith(object.prefix)`                                             | Validate that an object's name has the prefix of another field value                           |
| `object.set1.all(e, !(e in object.set2))`                                                    | Validate that two listSets are disjoint                                                        |
| `size(object.names) == size(object.details) && object.names.all(n, n in object.details)`     | Validate the 'details' map is keyed by the items in the 'names' listSet                        |
| `size(object.clusters.filter(c, c.name == object.primary)) == 1`                             | Validate that the 'primary' property has one and only one occurrence in the 'clusters' listMap |
