package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterIssueOverrideSpec stores user provided data to superseed the defaults
// from a given <ClusterIssue>. The last 3 fields carry the same type of
// information as its <ClusterIssue> equivalent, but the first is a list of
// clusters where this override is valid. An empty list emplies a global
// override.
type ClusterIssueOverrideSpec struct {
	Clusters []string              `json:"clusters,omitempty"`
	Message  *string               `json:"message,omitempty"`
	Severity *ClusterIssueSeverity `json:"severity,omitempty"`
	Category *string               `json:"category,omitempty"`
}

type ClusterIssueOverrideStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName="cio"
//+kubebuilder:printcolumn:name="Severity",type="string",JSONPath=".spec.severity",priority=0
//+kubebuilder:printcolumn:name="Category",type="string",JSONPath=".spec.category",priority=0
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".spec.message",priority=0
//+kubebuilder:printcolumn:name="Clusters",type="string",JSONPath=".spec.clusters",priority=1

// ClusterIssueOverride is the Schema for the clusterissueoverrides API. Its
// name is always the identifier of a cluster issue.
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterIssueOverride struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterIssueOverrideSpec   `json:"spec,omitempty"`
	Status ClusterIssueOverrideStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterIssueOverrideList contains a list of ClusterIssueOverride.
type ClusterIssueOverrideList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterIssueOverride `json:"items"`
}

func (r *ClusterIssueOverride) Hidden() bool {
	return r.Spec.Severity == nil && r.Spec.Category == nil && r.Spec.Message == nil
}

func (r *ClusterIssueOverride) InCluster(cl string) bool {
	if len(r.Spec.Clusters) == 0 {
		return true
	}
	for _, c := range r.Spec.Clusters {
		if c == cl {
			return true
		}
	}
	return false
}

func init() {
	SchemeBuilder.Register(&ClusterIssueOverride{}, &ClusterIssueOverrideList{})
}
