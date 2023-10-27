# Retain issues

By default, both scans automatically scheduled by Zora upon installation 
are configured to retain issues/results only from the last scan.

To retain results from the last two scans, for example, 
you should set the `successfulScansHistoryLimit` field of `ClusterScan` to `2`.

This can be done by either directly editing the `ClusterScan` object 
or by providing a parameter in the Helm installation/upgrade command, 

```shell
# omitted "helm upgrade --install" command and parameters

--set scan.misconfiguration.successfulScansHistoryLimit=2
```

In this case, it may appear that there are duplicate issues when more than one scan completes successfully. 
However, these issues are actually related to different scans. 
The identifier of each scan can be found in the `scanID` label of each issue.

```shell
kubectl get issues -n zora-system --show-labels
```
```
NAME                    CLUSTER     ID      MESSAGE                SEVERITY   CATEGORY   AGE    LABELS
kind-kind-m-102-4wxvv   kind-kind   M-102   Privileged container   High       Security   43s    scanID=556cc35a-830e-45af-a31c-7130918de262,category=Security,cluster=kind-kind,custom=false,id=M-102,plugin=marvin,severity=High
kind-kind-m-102-nf5xq   kind-kind   M-102   Privileged container   High       Security   102s   scanID=8464411a-4b9c-456b-a11c-dd3a5ab905f5,category=Security,cluster=kind-kind,custom=false,id=M-102,plugin=marvin,severity=High
```

To list issues from a specific scan, you can use a label selector like this:

```shell
kubectl get issues -n zora-system -l scanID=556cc35a-830e-45af-a31c-7130918de262
```

This also applies to vulnerability scans and `VulnerabilityReport` results.

!!! warning
    Note that results are stored as CRDs in your Kubernetes cluster. 
    Be cautious not to set a high value that could potentially affect 
    the performance and storage capacity of your Kubernetes cluster

!!! note
    That applies only to Zora OSS. Zora Dashboard always shows results from the last scan.
