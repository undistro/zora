package payloads

import (
	"fmt"
	"sort"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/getupio-undistro/zora/pkg/formats"
)

type ScanStatusType string

const (
	Failed  ScanStatusType = "failed"
	Unknown ScanStatusType = "unknown"
	Scanned ScanStatusType = "scanned"
)

type Cluster struct {
	Name                   string            `json:"name"`
	Namespace              string            `json:"namespace"`
	Environment            string            `json:"environment"`
	Provider               string            `json:"provider"`
	Region                 string            `json:"region"`
	TotalNodes             *int              `json:"totalNodes"`
	Version                string            `json:"version"`
	Scan                   *ScanStatus       `json:"scan"`
	Connection             *ConnectionStatus `json:"connection"`
	TotalIssues            *int              `json:"totalIssues"`
	Resources              *Resources        `json:"resources"`
	CreationTimestamp      metav1.Time       `json:"creationTimestamp"`
	Issues                 []ResourcedIssue  `json:"issues"`
	LastSuccessfulScanTime *metav1.Time      `json:"lastSuccessfulScanTime"`
	NextScheduleScanTime   *metav1.Time      `json:"nextScheduleScanTime"`
}

type ResourcedIssue struct {
	Issue     `json:",inline"`
	Resources map[string][]string `json:"resources"`
}

type Resources struct {
	Discovered bool      `json:"discovered"`
	Message    string    `json:"message"`
	Memory     *Resource `json:"memory"`
	CPU        *Resource `json:"cpu"`
}

type Resource struct {
	Available       string `json:"available"`
	Usage           string `json:"usage"`
	UsagePercentage int32  `json:"usagePercentage"`
}

type ScanStatus struct {
	Status  ScanStatusType `json:"status"`
	Message string         `json:"message"`
}

type ConnectionStatus struct {
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
}

func NewCluster(cluster v1alpha1.Cluster, scans []v1alpha1.ClusterScan) Cluster {
	cl := Cluster{
		Name:              cluster.Name,
		Namespace:         cluster.Namespace,
		Environment:       cluster.Labels[v1alpha1.LabelEnvironment],
		Provider:          cluster.Status.Provider,
		Region:            cluster.Status.Region,
		TotalNodes:        cluster.Status.TotalNodes,
		Version:           cluster.Status.KubernetesVersion,
		CreationTimestamp: cluster.Status.CreationTimestamp,
		Resources:         &Resources{},
		Scan:              &ScanStatus{Status: Unknown},
		Connection:        &ConnectionStatus{},
	}

	for _, c := range cluster.Status.Conditions {
		switch c.Type {
		case v1alpha1.ClusterReady:
			cl.Connection.Connected = c.Status == metav1.ConditionTrue
			cl.Connection.Message = c.Message
		case v1alpha1.ClusterResourcesDiscovered:
			cl.Resources.Discovered = c.Status == metav1.ConditionTrue
			cl.Resources.Message = c.Message
		}
	}

	var notFinished []string
	var failedPlugins []string
	var failed string
	for _, cs := range scans {
		// total issues
		if cl.TotalIssues == nil {
			cl.TotalIssues = cs.Status.TotalIssues
		} else {
			sum := *cl.TotalIssues + *cs.Status.TotalIssues
			cl.TotalIssues = &sum
		}

		// last and next scan time
		if cl.LastSuccessfulScanTime == nil || cl.LastSuccessfulScanTime.Before(cs.Status.LastSuccessfulTime) {
			cl.LastSuccessfulScanTime = cs.Status.LastSuccessfulTime
		}
		if cl.NextScheduleScanTime == nil || cs.Status.NextScheduleTime.Before(cl.NextScheduleScanTime) {
			cl.NextScheduleScanTime = cs.Status.NextScheduleTime
		}

		// scan status
		if cs.Status.LastFinishedStatus == string(batchv1.JobFailed) {
			failed = cs.Name
			for n, p := range cs.Status.Plugins {
				if p.LastFinishedStatus == string(batchv1.JobFailed) {
					failedPlugins = append(failedPlugins, n)
				}
			}
			break
		} else if cs.Status.LastFinishedTime == nil {
			notFinished = append(notFinished, cs.Name)
		}
	}

	if len(scans) == 0 {
		cl.Scan.Message = "No scan configured for this cluster"
	} else if failed != "" {
		sort.Strings(failedPlugins)
		cl.Scan.Status = Failed
		cl.Scan.Message = failedPluginsMessage(failed, failedPlugins)
		cl.TotalIssues = nil
	} else if len(notFinished) == len(scans) {
		cl.Scan.Message = "No finished scan yet for this cluster"
	} else {
		cl.Scan.Status = Scanned
	}

	if cpu, ok := cluster.Status.Resources[corev1.ResourceCPU]; ok {
		cl.Resources.CPU = &Resource{
			Available:       formats.CPU(cpu.Available),
			Usage:           formats.CPU(cpu.Usage),
			UsagePercentage: cpu.UsagePercentage,
		}
	}
	if mem, ok := cluster.Status.Resources[corev1.ResourceMemory]; ok {
		cl.Resources.Memory = &Resource{
			Available:       formats.Memory(mem.Available),
			Usage:           formats.Memory(mem.Usage),
			UsagePercentage: mem.UsagePercentage,
		}
	}

	return cl
}

func failedPluginsMessage(cs string, p []string) string {
	css := fmt.Sprintf("in the last scan of ClusterScan <%s>", cs)
	plen := len(p)
	if plen == 1 {
		return fmt.Sprintf("Plugin <%s> failed %s", p[0], css)
	}
	bu := &strings.Builder{}
	bu.WriteString("Plugins ")
	for c := 0; c < plen; c++ {
		bu.WriteString("<")
		bu.WriteString(p[c])
		bu.WriteString(">")
		if c == plen-2 {
			bu.WriteString(" and ")
		} else if c != plen-1 {
			bu.WriteString(", ")
		}
	}
	bu.WriteString(" failed ")
	bu.WriteString(css)
	return bu.String()
}

func NewResourcedIssue(i v1alpha1.ClusterIssue) ResourcedIssue {
	ri := ResourcedIssue{}
	ri.Issue = NewIssue(i)
	ri.Resources = i.Spec.Resources
	return ri
}

func NewClusterWithIssues(cluster v1alpha1.Cluster, scans []v1alpha1.ClusterScan, issues []v1alpha1.ClusterIssue) Cluster {
	c := NewCluster(cluster, scans)
	if c.Scan.Status != Failed {
		for _, i := range issues {
			c.Issues = append(c.Issues, NewResourcedIssue(i))
		}
	}
	return c
}
