package v1alpha1

import (
	"fmt"

	"github.com/getupio-undistro/snitch/pkg/apis"
	"github.com/getupio-undistro/snitch/pkg/discovery"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// KubeconfigRef is a reference to a secret that contains the kubeconfig data
	KubeconfigRef *corev1.SecretReference `json:"kubeconfigRef,omitempty"`

	Cloud *ClusterCloudSpec `json:"cloud,omitempty"`
}

type ClusterCloudSpec struct {
	EKS *ClusterEKSSpec `json:"eks,omitempty"`
	AKS *ClusterAKSSpec `json:"aks,omitempty"`
	GKE *ClusterGKESpec `json:"gke,omitempty"`
}

type ClusterEKSSpec struct {
	Name           string                 `json:"name"`
	Region         string                 `json:"region"`
	CredentialsRef corev1.SecretReference `json:"credentialsRef"`
}

type ClusterAKSSpec struct {
	Name           string                 `json:"name"`
	TenantID       string                 `json:"tenantID"`
	SubscriptionID string                 `json:"subscriptionID"`
	ResourceGroup  string                 `json:"resourceGroup"`
	CredentialsRef corev1.SecretReference `json:"credentialsRef"`
}

type ClusterGKESpec struct {
	Name           string                 `json:"name"`
	ServiceAccount string                 `json:"serviceAccount"`
	Zone           string                 `json:"zone,omitempty"`
	CredentialsRef corev1.SecretReference `json:"credentialsRef"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	apis.Status           `json:",inline"`
	discovery.ClusterInfo `json:",inline"`

	// Fields to be printed in columns

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
}

// SetClusterInfo fill ClusterInfo and temporary fields (TotalNodes, MemoryUsage and CPUUsage)
func (in *ClusterStatus) SetClusterInfo(c discovery.ClusterInfo) {
	in.ClusterInfo = c
	in.TotalNodes = len(c.Nodes)
	if m, found := in.ClusterInfo.Resources[corev1.ResourceMemory]; found {
		in.MemoryUsage = fmt.Sprintf("%vMi (%d%%)", m.Usage.Value()/(1024*1024), m.UsagePercentage)
		in.MemoryAvailable = fmt.Sprintf("%vMi", m.Available.Value()/(1024*1024))
	}
	if c, found := in.ClusterInfo.Resources[corev1.ResourceCPU]; found {
		in.CPUUsage = fmt.Sprintf("%vm (%d%%)", c.Usage.MilliValue(), c.UsagePercentage)
		in.CPUAvailable = fmt.Sprintf("%vm", c.Available.MilliValue())
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.kubernetesVersion",description=""
//+kubebuilder:printcolumn:name="MEM Available",type="string",JSONPath=".status.memoryAvailable",description=""
//+kubebuilder:printcolumn:name="MEM Usage (%)",type="string",JSONPath=".status.memoryUsage",description=""
//+kubebuilder:printcolumn:name="CPU Available",type="string",JSONPath=".status.cpuAvailable",description=""
//+kubebuilder:printcolumn:name="CPU Usage (%)",type="string",JSONPath=".status.cpuUsage",description=""
//+kubebuilder:printcolumn:name="Nodes",type="integer",JSONPath=".status.totalNodes",description=""

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

func (in *Cluster) GetKubeconfigSecretName() types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-kubeconfig", in.Name),
		Namespace: in.Namespace,
	}
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
