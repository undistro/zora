package payloads

import (
	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/getupio-undistro/zora/pkg/formats"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScanStatusType string

const (
	Failed  ScanStatusType = "failed"
	Unknown ScanStatusType = "unknown"
	Scanned ScanStatusType = "scanned"
)

type Cluster struct {
	Name                   string           `json:"name"`
	Namespace              string           `json:"namespace"`
	Environment            string           `json:"environment"`
	Provider               string           `json:"provider"`
	Region                 string           `json:"region"`
	TotalNodes             int              `json:"totalNodes"`
	Version                string           `json:"version"`
	Status                 ScanStatus       `json:"status"`
	TotalIssues            int              `json:"totalIssues"`
	Resources              *Resources       `json:"resources"`
	CreationTimestamp      metav1.Time      `json:"creationTimestamp"`
	Issues                 []ResourcedIssue `json:"issues"`
	LastSuccessfulScanTime metav1.Time      `json:"lastSuccessfulScanTime"`
	NextScheduleScanTime   metav1.Time      `json:"nextScheduleScanTime"`
}

type ResourcedIssue struct {
	Issue     `json:",inline"`
	Resources map[string][]string `json:"resources"`
}

type Resources struct {
	Memory *Resource `json:"memory"`
	CPU    *Resource `json:"cpu"`
}

type Resource struct {
	Available       string `json:"available"`
	Usage           string `json:"usage"`
	UsagePercentage int32  `json:"usagePercentage"`
}

type ScanStatus struct {
	Type    ScanStatusType `json:"scanStatusType"`
	Message string         `json:"message"`
}

func deriveStatus(conds []metav1.Condition) ScanStatus {
	for _, c := range conds {
		if c.Type == v1alpha1.ClusterReady && c.Status == metav1.ConditionFalse ||
			c.Type == v1alpha1.ClusterDiscovered && c.Status == metav1.ConditionFalse {
			return ScanStatus{Type: Failed, Message: c.Message}
		}
		if c.Type == v1alpha1.ClusterScanned && c.Status == metav1.ConditionFalse {
			if c.Reason == v1alpha1.ClusterNotScanned || c.Reason == v1alpha1.ClusterScanNotConfigured {
				return ScanStatus{Type: Unknown, Message: c.Message}
			}
			return ScanStatus{Type: Failed, Message: c.Message}
		}
	}
	return ScanStatus{Type: Scanned}
}

func NewCluster(cluster v1alpha1.Cluster) Cluster {
	cl := Cluster{
		Name:              cluster.Name,
		Namespace:         cluster.Namespace,
		Environment:       cluster.Labels[v1alpha1.LabelEnvironment],
		Provider:          cluster.Status.Provider,
		Region:            cluster.Status.Region,
		TotalNodes:        cluster.Status.TotalNodes,
		Version:           cluster.Status.KubernetesVersion,
		CreationTimestamp: cluster.Status.CreationTimestamp,
		TotalIssues:       cluster.Status.TotalIssues,
		Resources:         &Resources{},
	}
	if cluster.Status.LastSuccessfulScanTime != nil {
		cl.LastSuccessfulScanTime = *cluster.Status.LastSuccessfulScanTime
	}
	if cluster.Status.NextScheduleScanTime != nil {
		cl.NextScheduleScanTime = *cluster.Status.NextScheduleScanTime
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
	cl.Status = deriveStatus(cluster.Status.Conditions)

	return cl
}

func NewResourcedIssue(i v1alpha1.ClusterIssue) ResourcedIssue {
	ri := ResourcedIssue{}
	ri.Issue = NewIssue(i)
	ri.Resources = i.Spec.Resources
	return ri
}

func NewClusterWithIssues(cluster v1alpha1.Cluster, issues []v1alpha1.ClusterIssue) Cluster {
	c := NewCluster(cluster)
	for _, i := range issues {
		c.Issues = append(c.Issues, NewResourcedIssue(i))
	}
	c.TotalIssues = len(issues)
	return c
}
