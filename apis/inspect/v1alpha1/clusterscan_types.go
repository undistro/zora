package v1alpha1

import (
	"github.com/getupio-undistro/inspect/pkg/apis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ClusterScanSpec defines the desired state of ClusterScan
type ClusterScanSpec struct {
	// ClusterRef is a reference to a Cluster
	ClusterRef ClusterReference `json:"clusterRef"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	Suspend *bool `json:"suspend,omitempty"`

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// The list of Plugin references that are used to scan the referenced Cluster.  Defaults to 'popeye'
	Plugins []PluginReference `json:"plugins,omitempty"`
}

// ClusterReference represents a Cluster Reference. It has enough information to retrieve cluster
// in any namespace
type ClusterReference struct {
	// Name is unique within a namespace to reference a Cluster resource.
	Name string `json:"name"`

	// Namespace defines the space within which the Cluster name must be unique.
	Namespace string `json:"namespace,omitempty"`
}

type PluginReference struct {
	// Name is unique within a namespace to reference a Plugin resource.
	Name string `json:"name"`

	// Namespace defines the space within which the Plugin name must be unique.
	Namespace string `json:"namespace,omitempty"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	Suspend *bool `json:"suspend,omitempty"`

	// The schedule in Cron format for this Plugin, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule,omitempty"`

	// List of environment variables to set in the Plugin container.
	Env []corev1.EnvVar `json:"env,omitempty"`
}

func (in *PluginReference) PluginKey(defaultNamespace string) types.NamespacedName {
	ns := in.Namespace
	if ns == "" {
		ns = defaultNamespace
	}
	return types.NamespacedName{Name: in.Name, Namespace: ns}
}

// ClusterScanStatus defines the observed state of ClusterScan
type ClusterScanStatus struct {
	apis.Status           `json:",inline"`
	Plugins               string `json:"plugins,omitempty"`
	ClusterNamespacedName string `json:"clusterName,omitempty"`
	Suspend               bool   `json:"suspend"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".status.clusterName"
//+kubebuilder:printcolumn:name="Suspend",type="boolean",JSONPath=".status.suspend"
//+kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule"
//+kubebuilder:printcolumn:name="Plugins",type="string",JSONPath=".status.plugins"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"

// ClusterScan is the Schema for the clusterscans API
type ClusterScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterScanSpec   `json:"spec,omitempty"`
	Status ClusterScanStatus `json:"status,omitempty"`
}

func (in *ClusterScan) SetReadyStatus(status bool, reason, msg string) {
	s := metav1.ConditionFalse
	if status {
		s = metav1.ConditionTrue
	}
	in.Status.SetCondition(metav1.Condition{
		Type:               "Ready",
		Status:             s,
		ObservedGeneration: in.Generation,
		Reason:             reason,
		Message:            msg,
	})
}

func (in *ClusterScan) ClusterKey() types.NamespacedName {
	ns := in.Spec.ClusterRef.Namespace
	if ns == "" {
		ns = in.Namespace
	}
	return types.NamespacedName{Name: in.Spec.ClusterRef.Name, Namespace: ns}
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
