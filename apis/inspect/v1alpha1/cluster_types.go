package v1alpha1

import (
	"github.com/getupio-undistro/inspect/pkg/apis"
	"github.com/getupio-undistro/inspect/pkg/discovery"
	"github.com/getupio-undistro/inspect/pkg/formats"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	LabelEnvironment  = "inspect.undistro.io/environment"
	ClusterReady      = "Ready"
	ClusterDiscovered = "Discovered"
	ClusterScanned    = "Scanned"
)

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// KubeconfigRef is a reference to a secret in the same namespace that contains the kubeconfig data
	KubeconfigRef *corev1.LocalObjectReference `json:"kubeconfigRef,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	apis.Status           `json:",inline"`
	discovery.ClusterInfo `json:",inline"`

	// KubernetesVersion is the server's kubernetes version (git version).
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// Total of nodes
	TotalNodes int `json:"totalNodes,omitempty"`

	// Usage of memory in quantity and percentage
	MemoryUsage string `json:"memoryUsage,omitempty"`

	// Quantity of memory available in Mi
	MemoryAvailable string `json:"memoryAvailable,omitempty"`

	// Usage of CPU in quantity and percentage
	CPUUsage string `json:"cpuUsage,omitempty"`

	// Quantity of CPU available
	CPUAvailable string `json:"cpuAvailable,omitempty"`

	// Timestamp representing the server time of the last reconciliation
	LastRun metav1.Time `json:"lastRun,omitempty"`

	// Total of ClusterIssues reported by ClusterScan
	TotalIssues int `json:"totalIssues"`

	// List of last scan IDs
	LastScans []string `json:"lastScans,omitempty"`
}

// SetClusterInfo fill ClusterInfo and temporary fields (TotalNodes, MemoryUsage and CPUUsage)
func (in *ClusterStatus) SetClusterInfo(c discovery.ClusterInfo) {
	in.ClusterInfo = c
	in.TotalNodes = len(c.Nodes)
	if m, found := in.ClusterInfo.Resources[corev1.ResourceMemory]; found {
		in.MemoryAvailable = formats.Memory(m.Available)
		in.MemoryUsage = formats.MemoryUsage(m.Usage, m.UsagePercentage)
	}
	if c, found := in.ClusterInfo.Resources[corev1.ResourceCPU]; found {
		in.CPUAvailable = formats.CPU(c.Available)
		in.CPUUsage = formats.CPUUsage(c.Usage, c.UsagePercentage)
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Version",type="string",priority=0,JSONPath=".status.kubernetesVersion"
//+kubebuilder:printcolumn:name="MEM Available",type="string",priority=0,JSONPath=".status.memoryAvailable"
//+kubebuilder:printcolumn:name="MEM Usage (%)",type="string",priority=0,JSONPath=".status.memoryUsage"
//+kubebuilder:printcolumn:name="CPU Available",type="string",priority=0,JSONPath=".status.cpuAvailable"
//+kubebuilder:printcolumn:name="CPU Usage (%)",type="string",priority=0,JSONPath=".status.cpuUsage"
//+kubebuilder:printcolumn:name="Nodes",type="integer",priority=0,JSONPath=".status.totalNodes"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
//+kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".status.creationTimestamp"
//+kubebuilder:printcolumn:name="Provider",type="string",priority=1,JSONPath=".status.provider"
//+kubebuilder:printcolumn:name="Region",type="string",priority=1,JSONPath=".status.region"
//+kubebuilder:printcolumn:name="Issues",type="integer",priority=1,JSONPath=".status.totalIssues"

// Cluster is the Schema for the clusters API
//+genclient
//+genclient:onlyVerbs=list,get
//+genclient:noStatus
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

func (in *Cluster) KubeconfigRefKey() *types.NamespacedName {
	if in.Spec.KubeconfigRef == nil {
		return nil
	}
	return &types.NamespacedName{Name: in.Spec.KubeconfigRef.Name, Namespace: in.Namespace}
}

func (in *Cluster) SetStatus(statusType string, status bool, reason, msg string) {
	s := metav1.ConditionFalse
	if status {
		s = metav1.ConditionTrue
	}
	in.Status.SetCondition(metav1.Condition{
		Type:               statusType,
		Status:             s,
		ObservedGeneration: in.Generation,
		Reason:             reason,
		Message:            msg,
	})
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
