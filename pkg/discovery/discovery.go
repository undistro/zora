package discovery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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

func NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (ClusterDiscoverer, error) {
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

func NewForConfigOrDie(c *rest.Config) ClusterDiscoverer {
	return &clusterDiscovery{kubernetes: kubernetes.NewForConfigOrDie(c), metrics: versioned.NewForConfigOrDie(c)}
}

func New(c rest.Interface) ClusterDiscoverer {
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
	v, err := r.Version(ctx)
	if err != nil {
		return nil, err
	}

	nodes, err := r.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, errors.New("cluster has no nodes")
	}

	clsource, err := r.ClusterSource(ctx, nodes[0])
	if err != nil {
		return nil, err
	}
	reg, err := r.Region(ctx, nodes)
	if err != nil {
		return nil, err
	}

	return &ClusterInfo{
		KubernetesVersion: v,
		Nodes:             nodes,
		Resources:         avgNodeResources(nodes),
		CreationTimestamp: oldestNodeTimestamp(nodes),
		Provider:          clsource["provider"],
		Flavor:            clsource["flavor"],
		Region:            reg,
	}, nil
}

func (r *clusterDiscovery) ClusterSource(_ context.Context, node NodeInfo) (map[string]string, error) {
	match := false
	cls := map[string]string{}
	for l, _ := range node.Labels {
		for pr, cs := range ClusterSourcePrefixes {
			match = strings.HasPrefix(l, pr)
			if match {
				cls = cs
				break
			}
		}
		if match {
			break
		}
	}
	if !match {
		return nil, errors.New("no labels match cluster flavor")
	}
	return cls, nil
}

func (r *clusterDiscovery) Provider(_ context.Context, node NodeInfo) (string, error) {
	clsource, err := r.ClusterSource(nil, node)
	if err != nil {
		return "", fmt.Errorf("failed to discover provider: %w", err)
	}
	return clsource["provider"], nil
}

func (r *clusterDiscovery) Flavor(_ context.Context, node NodeInfo) (string, error) {
	clsource, err := r.ClusterSource(nil, node)
	if err != nil {
		return "", fmt.Errorf("failed to discover flavor: %w", err)
	}
	return clsource["flavor"], nil
}

func (r *clusterDiscovery) Region(_ context.Context, nodes []NodeInfo) (string, error) {
	regc := map[string]int{}
	haslabel := false
	for c := 0; c < len(nodes); c++ {
		for l, reg := range nodes[c].Labels {
			if l == RegionLabel {
				if !haslabel {
					haslabel = true
				}
				regc[reg]++
			}
		}
	}
	if !haslabel {
		return "", fmt.Errorf("unable to discover region: %w",
			fmt.Errorf("no node has the label <%s>", RegionLabel))
	}
	maxc := 0
	maxcreg := ""
	for reg, c := range regc {
		if maxc < c {
			maxc = c
			maxcreg = reg
		}
	}
	return maxcreg, nil
}

func (r *clusterDiscovery) Version(_ context.Context) (string, error) {
	v, err := r.kubernetes.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to discover server version: %w", err)
	}
	return v.String(), nil
}

func (r *clusterDiscovery) Nodes(ctx context.Context) ([]NodeInfo, error) {
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

func oldestNodeTimestamp(nodes []NodeInfo) metav1.Time {
	oldest := metav1.NewTime(time.Now().UTC())
	for _, node := range nodes {
		if node.CreationTimestamp.Before(&oldest) {
			oldest = node.CreationTimestamp
		}
	}
	return oldest
}
