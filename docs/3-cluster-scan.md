# Configure a cluster scan

Since your clusters are connected it's possible configure a scan for them
creating a `ClusterScan` resource in the same namespace as `Cluster`.

## Create a `ClusterScan`

Here is a sample configuration that scan `mycluster` once an hour.
You can modify per your needs/wants.

```shell
cat << EOF | kubectl apply -f -
apiVersion: zora.undistro.io/v1alpha1
kind: ClusterScan
metadata:
  name: mycluster
spec:
  clusterRef:
    name: mycluster
  schedule: "0 */1 * * *"
EOF
```

## List cluster scans

Listing the `ClusterScans`, the information of the last scans are available:

```shell
kubectl get clusterscan -o wide
NAME        CLUSTER     SCHEDULE      SUSPEND   PLUGINS   LAST STATUS   LAST SCHEDULE   LAST SUCCESSFUL   ISSUES   READY   AGE   NEXT SCHEDULE
mycluster   mycluster   0 */1 * * *   false     popeye    Complete      12m             14m               21       True    32d   2022-06-27T23:00:00Z
```

The `LAST STATUS` column represents the status (Active, Complete or Failed) of the last **scan** 
that was scheduled at the time represented by `LAST SCHEDULE` column.

## List cluster issues

Once the cluster is successfully scanned,
the reported issues are available in `ClusterIssue` resources:

```shell
kubectl get clusterissues -l cluster=mycluster
NAME                          CLUSTER      ID         MESSAGE                                                                        SEVERITY   CATEGORY          AGE
mycluster-pop-102-27557035    mycluster    POP-102    No probes defined                                                              Medium     pods              4m8s
mycluster-pop-105-27557035    mycluster    POP-105    Liveness probe uses a port#, prefer a named port                               Low        pods              4m8s
mycluster-pop-106-27557035    mycluster    POP-106    No resources requests/limits defined                                           Medium     daemonsets        4m8s
mycluster-pop-1100-27557035   mycluster    POP-1100   No pods match service selector                                                 High       services          4m8s
mycluster-pop-306-27557035    mycluster    POP-306    Container could be running as root user. Check SecurityContext/Image           Medium     pods              4m8s
mycluster-pop-500-27557035    mycluster    POP-500    Zero scale detected                                                            Medium     deployments       4m8s
```

It's possible filter issues by cluster, issue ID, severity and category:

```shell
# issues from mycluster
kubectl get clusterissues -l cluster=mycluster

# clusters with issue POP-106
kubectl get clusterissues -l id=POP-106

# issues from mycluster with high severity
kubectl get clusterissues -l cluster=mycluster,severity=high

# only issues reported by the last scan from mycluster
kubectl get clusterissues -l cluster=mycluster,scanID=fa4e63cc-5236-40f3-aa7f-599e1c83208b
```
