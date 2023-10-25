# Scan schedule

After successfully installing Zora, vulnerability and misconfiguration scans are 
automatically scheduled for your cluster, with each scan using different plugins.

Scan schedules are defined using Cron expressions. 
You can view the schedule for your cluster by listing `ClusterScan` resources:

```shell
kubectl get clusterscans -o wide -n zora-system
```

By default, the misconfiguration scan is scheduled to run every hour at the current minute plus 5, 
while the vulnerability scan is scheduled to run every day at the current hour and the current minute plus 5.

For example, if the installation occurred at 10:00 UTC, the scans will have the following schedules:

| Scan              | Cron         | Description            |
|:------------------|:-------------|:-----------------------|
| Misconfigurations | `5 * * * *`  | Every hour at minute 5 |
| Vulnerabilities   | `5 10 * * *` | Every day at 10:05     |

However, you can customize the schedule for each scan 
by directly editing the `ClusterScan` resource 
or by providing parameters in the `helm upgrade --install` command, as shown in the example below:

```shell 
# omitted command and parameters

--set scan.misconfiguration.schedule="0 * * * *" \
--set scan.vulnerability.schedule="0 0 * * *"
```

The recommended approach is to provide parameters through Helm.

!!! warning
    If you directly edit the `ClusterScan` resource, be cautious when running the next update via Helm, as the value may be overwritten.

## Cron schedule syntax

Cron expression has five fields separated by a space, and each field represents a time unit.

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of the month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday;
│ │ │ │ │                                   7 is also Sunday on some systems)
│ │ │ │ │                                   OR sun, mon, tue, wed, thu, fri, sat
│ │ │ │ │
* * * * *
```

| Operator | Descriptor           | Example                                                                                            |
|----------|----------------------|----------------------------------------------------------------------------------------------------|
| *        | Any value            | `15 * * * *` runs at every minute 15 of every hour of every day.                                   |
| ,        | Value list separator | `2,10 4,5 * * *` runs at minute 2 and 10 of the 4th and 5th hour of every day.                     |
| -        | Range of values      | `30 4-6 * * *` runs at minute 30 of the 4th, 5th, and 6th hour.                                    |
| /        | Step values          | `20/15 * * * *` runs every 15 minutes starting from minute 20 through 59 (minutes 20, 35, and 50). |

