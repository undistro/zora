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

package saas

import (
	"strings"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

type Issue struct {
	ApiVersion    string `json:"apiVersion"`
	ID            string `json:"id"`
	Message       string `json:"message"`
	Severity      string `json:"severity"`
	Category      string `json:"category"`
	Plugin        string `json:"plugin"`
	Url           string `json:"url"`
	ClusterScoped bool   `json:"clusterScoped"`
	Custom        bool   `json:"custom"`
}

type ResourcedIssue struct {
	Issue     `json:",inline"`
	Resources map[string][]NamespacedName `json:"resources"`
}

type ClusterReference struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	TotalResources int    `json:"totalResources"`
}

func NewIssue(clusterIssue v1alpha1.ClusterIssue) Issue {
	return Issue{
		ApiVersion:    "v1alpha1",
		ID:            clusterIssue.Spec.ID,
		Message:       clusterIssue.Spec.Message,
		Severity:      string(clusterIssue.Spec.Severity),
		Category:      clusterIssue.Spec.Category,
		Plugin:        clusterIssue.Labels[v1alpha1.LabelPlugin],
		Url:           clusterIssue.Spec.Url,
		ClusterScoped: len(clusterIssue.Spec.Resources) <= 0,
		Custom:        clusterIssue.Spec.Custom,
	}
}

func NewResourcedIssue(i v1alpha1.ClusterIssue) ResourcedIssue {
	ri := ResourcedIssue{}
	ri.Issue = NewIssue(i)
	for r, narr := range i.Spec.Resources {
		for _, nspacedn := range narr {
			ns := strings.Split(nspacedn, "/")
			if len(ns) == 1 {
				ns = append([]string{""}, ns[0])
			}
			if ri.Resources == nil {
				ri.Resources = map[string][]NamespacedName{
					r: {{
						Name:      ns[1],
						Namespace: ns[0],
					}},
				}
			} else {
				ri.Resources[r] = append(ri.Resources[r],
					NamespacedName{
						Name:      ns[1],
						Namespace: ns[0],
					},
				)
			}
		}
	}
	return ri
}
