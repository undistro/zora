package discovery

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MeasuredResources = []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory}

type ClusterDiscovery interface {
	Discover(context.Context) (*ClusterInfo, error)
	DiscoverVersion(context.Context) (string, error)
	DiscoverNodes(context.Context) ([]NodeInfo, error)
}

// +k8s:deepcopy-gen=true
type ClusterInfo struct {
	// KubernetesVersion is the server's kubernetes version (git version).
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// Info from cluster nodes
	Nodes []NodeInfo `json:"nodes,omitempty"`

	// Average of usage and available resources
	Resources map[corev1.ResourceName]Resources `json:"resources,omitempty"`

	// CreationTimestamp is a timestamp representing the server time when the oldest Node was created.
	// It is represented in RFC3339 form and is in UTC.
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty"`
}

// +k8s:deepcopy-gen=true
type NodeInfo struct {
	// Node name
	Name string `json:"name,omitempty"`

	// Node labels
	Labels map[string]string `json:"labels,omitempty"`

	// Usage and available resources
	Resources map[corev1.ResourceName]Resources `json:"resources,omitempty"`

	// True if node is in ready condition
	Ready bool `json:"ready,omitempty"`

	// CreationTimestamp is a timestamp representing the server time when this object was created.
	// It is represented in RFC3339 form and is in UTC.
	CreationTimestamp metav1.Time `json:"-"`
}

// +k8s:deepcopy-gen=true
type Resources struct {
	// Quantity of resources available for scheduling
	Available resource.Quantity `json:"available,omitempty"`

	// Quantity of resources in use
	Usage resource.Quantity `json:"usage,omitempty"`

	// Percentage of resources in use
	UsagePercentage int32 `json:"usagePercentage,omitempty"`
}
