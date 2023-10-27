# Suspending scans

The cluster scans, which are automatically scheduled upon installation, 
can be suspended by setting `spec.suspend` to `true` in a `ClusterScan` object. 
This action will suspend subsequent scans, it does not apply to already started scans.

The command below suspends the `mycluster-vuln` scan.

```shell
kubectl patch scan mycluster-vuln --type='merge' -p '{"spec":{"suspend":true}}' -n zora-system
```

!!! note
    This way, the scan results remain available, 
    unlike if the `ClusterScan` had been deleted, in which case the results would also be removed.

Setting `spec.suspend` back to `false`, the scans are resume:

```shell
kubectl patch scan mycluster-vuln --type='merge' -p '{"spec":{"suspend":false}}' -n zora-system
```
