package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:validation:Enum=Unknown;Low;Medium;High

type ClusterIssueSeverity string

const (
	SeverityUnknown ClusterIssueSeverity = "Unknown"
	SeverityLow     ClusterIssueSeverity = "Low"
	SeverityMedium  ClusterIssueSeverity = "Medium"
	SeverityHigh    ClusterIssueSeverity = "High"

	LabelScanID   = "scanID"
	LabelCluster  = "cluster"
	LabelSeverity = "severity"
	LabelIssueID  = "id"
	LabelCategory = "category"
	LabelPlugin   = "plugin"
)

// ClusterIssueSpec defines the desired state of ClusterIssue
type ClusterIssueSpec struct {
	Cluster        string               `json:"cluster"`
	ID             string               `json:"id"`
	Message        string               `json:"message"`
	Severity       ClusterIssueSeverity `json:"severity"`
	Category       string               `json:"category,omitempty"`
	Resources      map[string][]string  `json:"resources,omitempty"`
	TotalResources int                  `json:"totalResources,omitempty"`
	Url            string               `json:"url,omitempty"`
}

// ClusterIssueStatus defines the observed state of ClusterIssue
type ClusterIssueStatus struct {
	Hidden       bool                  `json:"hidden,omitempty"`
	OrigMessage  *string               `json:"origMessage,omitempty"`
	OrigSeverity *ClusterIssueSeverity `json:"origSeverity,omitempty"`
	OrigCategory *string               `json:"origCategory,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName="ci"
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.cluster",priority=0
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".spec.id",priority=0
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".spec.message",priority=0
//+kubebuilder:printcolumn:name="Severity",type="string",JSONPath=".spec.severity",priority=0
//+kubebuilder:printcolumn:name="Category",type="string",JSONPath=".spec.category",priority=0
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0
//+kubebuilder:printcolumn:name="Total",type="integer",JSONPath=".spec.totalResources",priority=1

// ClusterIssue is the Schema for the clusterissues API
//+genclient
//+genclient:noStatus
type ClusterIssue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterIssueSpec   `json:"spec,omitempty"`
	Status ClusterIssueStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterIssueList contains a list of ClusterIssue
type ClusterIssueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterIssue `json:"items"`
}

// HasOverride tells whether an issue has an override.
func (r *ClusterIssue) HasOverride() bool {
	return r.Status.Hidden ||
		r.Status.OrigSeverity != nil || r.Status.OrigCategory != nil || r.Status.OrigMessage != nil
}

func init() {
	SchemeBuilder.Register(&ClusterIssue{}, &ClusterIssueList{})
}
