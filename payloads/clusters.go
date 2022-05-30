package payloads

import (
	"github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/pkg/formats"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Cluster struct {
	Name              string      `json:"name"`
	Namespace         string      `json:"namespace"`
	Environment       string      `json:"environment"`
	Provider          string      `json:"provider"`
	Region            string      `json:"region"`
	TotalNodes        int         `json:"totalNodes"`
	Version           string      `json:"version"`
	Ready             bool        `json:"ready"`
	TotalIssues       int         `json:"totalIssues"`
	Resources         *Resources  `json:"resources"`
	CreationTimestamp metav1.Time `json:"creationTimestamp"`
	Issues            []Issue     `json:"issues"`
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

func NewCluster(cluster v1alpha1.Cluster) Cluster {
	res := &Resources{}
	if cpu, ok := cluster.Status.Resources[corev1.ResourceCPU]; ok {
		res.CPU = &Resource{
			Available:       formats.CPU(cpu.Available),
			Usage:           formats.CPU(cpu.Usage),
			UsagePercentage: cpu.UsagePercentage,
		}
	}
	if mem, ok := cluster.Status.Resources[corev1.ResourceMemory]; ok {
		res.Memory = &Resource{
			Available:       formats.Memory(mem.Available),
			Usage:           formats.Memory(mem.Usage),
			UsagePercentage: mem.UsagePercentage,
		}
	}
	return Cluster{
		Name:              cluster.Name,
		Namespace:         cluster.Namespace,
		Environment:       cluster.Labels[v1alpha1.LabelEnvironment],
		Provider:          cluster.Status.Provider,
		Region:            cluster.Status.Region,
		TotalNodes:        cluster.Status.TotalNodes,
		Ready:             cluster.Status.ConditionIsTrue(v1alpha1.ClusterReady),
		Version:           cluster.Status.KubernetesVersion,
		CreationTimestamp: cluster.Status.CreationTimestamp,
		TotalIssues:       cluster.Status.TotalIssues,
		Resources:         res,
	}
}

func NewClusterWithIssues(cluster v1alpha1.Cluster, issues []v1alpha1.ClusterIssue) Cluster {
	c := NewCluster(cluster)
	for _, i := range issues {
		c.Issues = append(c.Issues, NewIssue(i))
	}
	c.TotalIssues = len(issues)
	return c
}
