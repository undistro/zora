package discovery

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MeasuredResources = []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory}

type ClusterDiscoverer interface {
	Discover(context.Context) (*ClusterInfo, error)
	Version() (string, error)
	Nodes(context.Context) ([]NodeInfo, error)
	Provider(NodeInfo) string
	Region([]NodeInfo) (string, error)
}

// +k8s:deepcopy-gen=true
type ClusterInfo struct {
	// Info from cluster nodes
	Nodes []NodeInfo `json:"-"`

	// Average of usage and available resources
	Resources map[corev1.ResourceName]Resources `json:"resources,omitempty"`

	// CreationTimestamp is a timestamp representing the server time when the oldest Node was created.
	// It is represented in RFC3339 form and is in UTC.
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty"`

	// Provider stores the cluster's source.
	Provider string `json:"provider,omitempty"`

	// Region holds the geographic location with most nodes.
	Region string `json:"region,omitempty"`
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
