package discovery

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func NewForConfig(c *rest.Config) (ClusterDiscoverer, error) {
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

func NewResources(available, usage resource.Quantity) Resources {
	fraction := float64(usage.MilliValue()) / float64(available.MilliValue()) * 100
	return Resources{Available: available, Usage: usage, UsagePercentage: int32(fraction)}
}

type clusterDiscovery struct {
	kubernetes *kubernetes.Clientset
	metrics    *versioned.Clientset
}

func (r *clusterDiscovery) Info(ctx context.Context) (*ClusterInfo, error) {
	nodeList, err := r.kubernetes.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Nodes: %w", err)
	}

	ns, err := r.kubernetes.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get <kube-system> namespace: %w", err)
	}

	totalNodes := len(nodeList.Items)
	return &ClusterInfo{
		TotalNodes:        &totalNodes,
		CreationTimestamp: ns.CreationTimestamp,
		Provider:          r.provider(nodeList.Items),
		Region:            r.region(nodeList.Items),
	}, nil
}

func (r *clusterDiscovery) Resources(ctx context.Context) (ClusterResources, error) {
	nodes, err := r.nodes(ctx)
	if err != nil {
		return nil, err
	}
	return sumNodeResources(nodes), nil
}

// Provider finds the cluster source by matching against provider specific
// labels on a node, returning the provider if the match succeeds and
// empty if it fails.
func (r *clusterDiscovery) provider(nodes []corev1.Node) string {
	for _, node := range nodes {
		for l := range node.Labels {
			for pref, p := range ClusterSourcePrefixes {
				if strings.HasPrefix(l, pref) {
					return p
				}
			}
		}
	}
	return ""
}

// region returns "multi-region" if the cluster nodes belong to distinct
// locations, otherwise it returns the region itself.
func (r *clusterDiscovery) region(nodes []corev1.Node) string {
	regs := map[string]bool{}
	for _, n := range nodes {
		for l, v := range n.Labels {
			if l == RegionLabel {
				regs[v] = true
				if len(regs) > 1 {
					return "multi-region"
				}
			}
		}
	}
	reg := ""
	for reg = range regs {
		continue
	}
	return reg
}

func (r *clusterDiscovery) Version() (string, error) {
	v, err := r.kubernetes.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to discover server version: %w", err)
	}
	return v.String(), nil
}

func (r *clusterDiscovery) nodes(ctx context.Context) ([]NodeInfo, error) {
	if err := r.checkMetricsAPI(); err != nil {
		return nil, err
	}
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

func (r *clusterDiscovery) checkMetricsAPI() error {
	apiGroups, err := r.kubernetes.Discovery().ServerGroups()
	if err != nil {
		return err
	}
	for _, group := range apiGroups.Groups {
		if group.Name != v1beta1.GroupName {
			continue
		}
		for _, version := range group.Versions {
			if version.Version == v1beta1.SchemeGroupVersion.Version {
				return nil
			}
		}
	}
	return errors.New("metrics API not available")
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
			Name:              n.Name,
			Labels:            n.Labels,
			Resources:         make(map[corev1.ResourceName]Resources),
			Ready:             nodeIsReady(n),
			CreationTimestamp: n.CreationTimestamp,
		}
		for _, res := range MeasuredResources {
			info.Resources[res] = NewResources(n.Status.Allocatable[res], usage[res])
		}
		infos = append(infos, info)
	}
	return infos
}

func sumNodeResources(nodes []NodeInfo) map[corev1.ResourceName]Resources {
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
