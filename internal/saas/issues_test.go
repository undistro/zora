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
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

func TestNewResourcedIssue(t *testing.T) {
	tests := []struct {
		name         string
		clusterIssue v1alpha1.ClusterIssue
		want         ResourcedIssue
	}{
		{
			name: "POP-402",
			clusterIssue: v1alpha1.ClusterIssue{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prd-pop-402-123",
					Namespace: "prd",
				},
				Spec: v1alpha1.ClusterIssueSpec{
					Cluster:        "prd",
					ID:             "POP-402",
					Message:        "No metrics-server detected",
					Severity:       "Low",
					Category:       "General",
					TotalResources: 0,
					Resources:      nil,
				},
			},
			want: ResourcedIssue{
				Issue: Issue{
					ApiVersion:    "v1alpha1",
					ID:            "POP-402",
					Message:       "No metrics-server detected",
					Severity:      "Low",
					Category:      "General",
					ClusterScoped: true,
				},
			},
		},
		{
			name: "POP-405",
			clusterIssue: v1alpha1.ClusterIssue{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prd-pop-405-123",
					Namespace: "prd",
				},
				Spec: v1alpha1.ClusterIssueSpec{
					Cluster:        "prd",
					ID:             "POP-405",
					Message:        "Is this a jurassic cluster? Might want to upgrade K8s a bit",
					Severity:       "Medium",
					Category:       "General",
					TotalResources: 0,
					Resources:      map[string][]string{},
				},
			},
			want: ResourcedIssue{
				Issue: Issue{
					ApiVersion:    "v1alpha1",
					ID:            "POP-405",
					Message:       "Is this a jurassic cluster? Might want to upgrade K8s a bit",
					Severity:      "Medium",
					Category:      "General",
					ClusterScoped: true,
				},
			},
		},
		{
			name: "POP-106",
			clusterIssue: v1alpha1.ClusterIssue{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prd-pop-106-123",
					Namespace: "prd",
					Labels:    map[string]string{v1alpha1.LabelScanID: "123"},
				},
				Spec: v1alpha1.ClusterIssueSpec{
					Cluster:        "prd",
					ID:             "POP-106",
					Message:        "No resources requests/limits defined",
					Severity:       "Medium",
					Category:       "Category",
					Url:            "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
					TotalResources: 1,
					Resources: map[string][]string{
						"apps/v1/deployments": {"ns/dep1"},
					},
				},
			},
			want: ResourcedIssue{
				Issue: Issue{
					ApiVersion: "v1alpha1",
					ID:         "POP-106",
					Message:    "No resources requests/limits defined",
					Severity:   "Medium",
					Category:   "Category",
					Url:        "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
				},
				Resources: map[string][]NamespacedName{
					"apps/v1/deployments": {NamespacedName{Name: "dep1", Namespace: "ns"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewResourcedIssue(tt.clusterIssue)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIssues() = %s", cmp.Diff(got, tt.want))
			}
		})
	}
}
