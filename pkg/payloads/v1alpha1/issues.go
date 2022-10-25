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
	"bytes"
	"encoding/json"
	"io"

	"github.com/undistro/zora/apis/zora/v1alpha1"
)

// +k8s:deepcopy-gen=true
type Issue struct {
	ApiVersion string             `json:"apiVersion"`
	ID         string             `json:"id"`
	Message    string             `json:"message"`
	Severity   string             `json:"severity"`
	Category   string             `json:"category"`
	Plugin     string             `json:"plugin"`
	Clusters   []ClusterReference `json:"clusters"`
	Url        string             `json:"url"`
}

// +k8s:deepcopy-gen=true
type ClusterReference struct {
	NsName
	TotalResources int `json:"totalResources"`
}

func NewIssue(clusterIssue v1alpha1.ClusterIssue) Issue {
	return Issue{
		ApiVersion: "v1alpha1",
		ID:         clusterIssue.Spec.ID,
		Message:    clusterIssue.Spec.Message,
		Severity:   string(clusterIssue.Spec.Severity),
		Category:   clusterIssue.Spec.Category,
		Plugin:     clusterIssue.Labels[v1alpha1.LabelPlugin],
		Url:        clusterIssue.Spec.Url,
	}
}

func NewIssues(clusterIssues []v1alpha1.ClusterIssue) []Issue {
	issuesByID := make(map[string]*Issue)
	clustersByIssue := make(map[string]map[string]*ClusterReference)
	for _, clusterIssue := range clusterIssues {
		clusterRef := &ClusterReference{
			NsName: NsName{
				Name:      clusterIssue.Spec.Cluster,
				Namespace: clusterIssue.Namespace,
			},
			TotalResources: clusterIssue.Spec.TotalResources,
		}
		newIssue := NewIssue(clusterIssue)
		issueID := clusterIssue.Spec.ID
		issuesByID[issueID] = &newIssue
		if _, ok := clustersByIssue[issueID]; !ok {
			clustersByIssue[issueID] = make(map[string]*ClusterReference)
		}
		clustersByIssue[issueID][clusterRef.Name] = clusterRef
	}
	res := make([]Issue, 0, len(issuesByID))
	for id, i := range issuesByID {
		for _, ref := range clustersByIssue[id] {
			i.Clusters = append(i.Clusters, *ref)
		}

		res = append(res, *i)
	}
	return res
}

func (r Issue) Reader() (io.Reader, error) {
	jc, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(jc), nil
}
