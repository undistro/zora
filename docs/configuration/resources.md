# Compute resources

[Zora Helm Chart](../helm-chart.md) allows you to define resource requests and limits (memory and CPU) 
for `zora-operator` and plugins.
You can do this by setting specific parameters using `--set` argument as the example below.

```
--set operator.resources.limits.memory=256Mi
```

Alternatively, a YAML file can be specified using `-f myvalues.yaml` flag.

!!! tip
    Refer to the default [values.yaml](../values.yaml) file for more details

In a similar way, you can customize the resources for plugins.
The following example sets `1Gi` as memory limit for `marvin` plugin.

```
--set scan.plugins.marvin.resources.limits.memory=1Gi
```

Resources requests and limits can also be set for `worker` container ([See how plugins work](../plugins/#how-plugins-work)):
```
--set scan.plugins.trivy.workerResources.limits.memory=1Gi
```
