package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterScanSpec defines the desired state of ClusterScan
type ClusterScanSpec struct {
	ClusterRef *ClusterReference `json:"clusterRef"`
	Suspend    bool              `json:"suspend,omitempty"`
	Schedule   string            `json:"schedule"`
	Plugins    []PluginReference `json:"plugins,omitempty"`
}

// ClusterReference represents a Cluster Reference. It has enough information to retrieve cluster
// in any namespace
type ClusterReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type PluginReference struct {
	Name      string          `json:"name"`
	Namespace string          `json:"namespace,omitempty"`
	Schedule  string          `json:"schedule,omitempty"`
	Env       []corev1.EnvVar `json:"env,omitempty"`
}

// ClusterScanStatus defines the observed state of ClusterScan
type ClusterScanStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterScan is the Schema for the clusterscans API
type ClusterScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterScanSpec   `json:"spec,omitempty"`
	Status ClusterScanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterScanList contains a list of ClusterScan
type ClusterScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterScan{}, &ClusterScanList{})
}
