// Copyright 2022 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	LabelScanID     = "scanID"
	LabelCluster    = "cluster"
	LabelClusterUID = "clusterUID"
	LabelSeverity   = "severity"
	LabelIssueID    = "id"
	LabelCategory   = "category"
	LabelPlugin     = "plugin"
	LabelCustom     = "custom"
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
	Custom         bool                 `json:"custom,omitempty"`
}

// AddResource appends the given resource to the Resources map, if it does not exist
func (r *ClusterIssueSpec) AddResource(gvr, resource string) {
	if res, ok := r.Resources[gvr]; ok {
		for _, re := range res {
			if re == resource {
				return
			}
		}
	}
	r.Resources[gvr] = append(r.Resources[gvr], resource)
	r.TotalResources++
}

// ClusterIssueStatus defines the observed state of ClusterIssue
type ClusterIssueStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName={issue,issues,misconfig,misconfigs,misconfigurations}
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.cluster",priority=0
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".spec.id",priority=0
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".spec.message",priority=0
//+kubebuilder:printcolumn:name="Severity",type="string",JSONPath=".spec.severity",priority=0
//+kubebuilder:printcolumn:name="Category",type="string",JSONPath=".spec.category",priority=0
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0
//+kubebuilder:printcolumn:name="Total",type="integer",JSONPath=".spec.totalResources",priority=1

// ClusterIssue is the Schema for the clusterissues API
// +genclient
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

func init() {
	SchemeBuilder.Register(&ClusterIssue{}, &ClusterIssueList{})
}
