package discovery

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func NewForConfig(c *rest.Config) (ClusterDiscovery, error) {
	kclient, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	mclient, err := versioned.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &clusterDiscovery{kubernetes: kclient, metrics: mclient}, nil
}

func NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (ClusterDiscovery, error) {
	kclient, err := kubernetes.NewForConfigAndClient(c, httpClient)
	if err != nil {
		return nil, err
	}
	mclient, err := versioned.NewForConfigAndClient(c, httpClient)
	if err != nil {
		return nil, err
	}
	return &clusterDiscovery{kubernetes: kclient, metrics: mclient}, nil
}

func NewForConfigOrDie(c *rest.Config) ClusterDiscovery {
	return &clusterDiscovery{kubernetes: kubernetes.NewForConfigOrDie(c), metrics: versioned.NewForConfigOrDie(c)}
}

func New(c rest.Interface) ClusterDiscovery {
	return &clusterDiscovery{kubernetes: kubernetes.New(c), metrics: versioned.New(c)}
}

func NewResources(available, usage resource.Quantity) Resources {
	fraction := float64(usage.MilliValue()) / float64(available.MilliValue()) * 100
	return Resources{Available: available, Usage: usage, UsagePercentage: int32(fraction)}
}

type clusterDiscovery struct {
	kubernetes *kubernetes.Clientset
	metrics    *versioned.Clientset
}

func (r *clusterDiscovery) Discover(ctx context.Context) (*ClusterInfo, error) {
	v, err := r.DiscoverVersion(ctx)
	if err != nil {
		return nil, err
	}

	nodes, err := r.DiscoverNodes(ctx)
	if err != nil {
		return nil, err
	}

	return &ClusterInfo{KubernetesVersion: v, Nodes: nodes, Resources: avgNodeResources(nodes)}, nil
}

func (r *clusterDiscovery) DiscoverVersion(_ context.Context) (string, error) {
	v, err := r.kubernetes.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to discover server version: %w", err)
	}
	return v.String(), nil
}

func (r *clusterDiscovery) DiscoverNodes(ctx context.Context) ([]NodeInfo, error) {
	metricsList, err := r.metrics.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list NodeMetrics: %w", err)
	}
	nodeList, err := r.kubernetes.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Nodes: %w", err)
	}

	return nodeResources(nodeList.Items, metricsList.Items), nil
}

func nodeResources(nodes []corev1.Node, nodeMetrics []v1beta1.NodeMetrics) []NodeInfo {
	infos := make([]NodeInfo, 0, len(nodes))
	metrics := make(map[string]corev1.ResourceList)
	for _, m := range nodeMetrics {
		metrics[m.Name] = m.Usage
	}
	for _, n := range nodes {
		usage := metrics[n.Name]
		info := NodeInfo{
			Name:      n.Name,
			Labels:    n.Labels,
			Resources: make(map[corev1.ResourceName]Resources),
			Ready:     nodeIsReady(n),
		}
		for _, res := range MeasuredResources {
			info.Resources[res] = NewResources(n.Status.Allocatable[res], usage[res])
		}
		infos = append(infos, info)
	}
	return infos
}

func avgNodeResources(nodes []NodeInfo) map[corev1.ResourceName]Resources {
	totalAvailable := make(map[corev1.ResourceName]*resource.Quantity)
	totalUsage := make(map[corev1.ResourceName]*resource.Quantity)

	for _, node := range nodes {
		for _, res := range MeasuredResources {
			if r, found := node.Resources[res]; found {
				if _, ok := totalAvailable[res]; ok {
					totalAvailable[res].Add(r.Available)
					totalUsage[res].Add(r.Usage)
				} else {
					totalAvailable[res] = &r.Available
					totalUsage[res] = &r.Usage
				}
			}
		}
	}
	result := make(map[corev1.ResourceName]Resources)
	for _, res := range MeasuredResources {
		result[res] = NewResources(*totalAvailable[res], *totalUsage[res])
	}
	return result
}

func nodeIsReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
