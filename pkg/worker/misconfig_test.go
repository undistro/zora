// Copyright 2023 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worker

import (
	"context"
	"io"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

func TestParseMisconfigResults(t *testing.T) {
	type args struct {
		cfg      *config
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    []v1alpha1.ClusterIssue
		wantErr bool
	}{
		{
			name:    "invalid plugin",
			args:    args{cfg: &config{PluginName: "trivy"}}, // trivy is not a misconfiguration plugin
			want:    nil,
			wantErr: true,
		},
		{
			name: "directory reader",
			args: args{
				cfg:      &config{PluginName: "marvin"},
				filename: t.TempDir(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "marvin",
			args: args{
				cfg: &config{
					PluginName:  "marvin",
					ClusterName: "cluster",
					ClusterUID:  "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
					Namespace:   "ns",
					JobName:     "cluster-marvin-28140229",
					JobUID:      "50c8957e-c9e1-493a-9fa4-d0786deea017",
					PodName:     "cluster-marvin-28140229-h9kcn",
					suffix:      "h9kcn",
				},
				filename: "report/marvin/testdata/httpbin.json",
			},
			want: []v1alpha1.ClusterIssue{
				{
					TypeMeta: clusterIssueTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-m-400-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-marvin-28140229",
								UID:        "50c8957e-c9e1-493a-9fa4-d0786deea017",
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:     "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster:    "cluster",
							v1alpha1.LabelClusterUID: "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
							v1alpha1.LabelSeverity:   string(v1alpha1.SeverityMedium),
							v1alpha1.LabelIssueID:    "M-400",
							v1alpha1.LabelCategory:   "BestPractices",
							v1alpha1.LabelPlugin:     "marvin",
							v1alpha1.LabelCustom:     "false",
						},
					},
					Spec: v1alpha1.ClusterIssueSpec{
						Cluster:  "cluster",
						ID:       "M-400",
						Message:  "Image tagged latest",
						Severity: v1alpha1.SeverityMedium,
						Category: "Best Practices",
						Resources: map[string][]string{
							"apps/v1/deployments": {"httpbin/httpbin"},
							"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
						},
						Url: "https://kubernetes.io/docs/concepts/containers/images/#image-names",
					},
				},
				{
					TypeMeta: clusterIssueTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-m-407-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-marvin-28140229",
								UID:        "50c8957e-c9e1-493a-9fa4-d0786deea017",
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:     "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster:    "cluster",
							v1alpha1.LabelClusterUID: "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
							v1alpha1.LabelSeverity:   string(v1alpha1.SeverityMedium),
							v1alpha1.LabelIssueID:    "M-407",
							v1alpha1.LabelCategory:   "Reliability",
							v1alpha1.LabelPlugin:     "marvin",
							v1alpha1.LabelCustom:     "false",
						},
					},
					Spec: v1alpha1.ClusterIssueSpec{
						Cluster:  "cluster",
						ID:       "M-407",
						Message:  "CPU not limited",
						Severity: v1alpha1.SeverityMedium,
						Category: "Reliability",
						Resources: map[string][]string{
							"apps/v1/deployments": {"httpbin/httpbin"},
							"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
						},
						Url: "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
					},
				},
				{
					TypeMeta: clusterIssueTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-m-116-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-marvin-28140229",
								UID:        "50c8957e-c9e1-493a-9fa4-d0786deea017",
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:     "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster:    "cluster",
							v1alpha1.LabelClusterUID: "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
							v1alpha1.LabelSeverity:   string(v1alpha1.SeverityLow),
							v1alpha1.LabelIssueID:    "M-116",
							v1alpha1.LabelCategory:   "Security",
							v1alpha1.LabelPlugin:     "marvin",
							v1alpha1.LabelCustom:     "false",
						},
					},
					Spec: v1alpha1.ClusterIssueSpec{
						Cluster:  "cluster",
						ID:       "M-116",
						Message:  "Not allowed added/dropped capabilities",
						Severity: v1alpha1.SeverityLow,
						Category: "Security",
						Resources: map[string][]string{
							"apps/v1/deployments": {"httpbin/httpbin"},
							"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
						},
						Url: "https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted",
					},
				},
				{
					TypeMeta: clusterIssueTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-m-113-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-marvin-28140229",
								UID:        "50c8957e-c9e1-493a-9fa4-d0786deea017",
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:     "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster:    "cluster",
							v1alpha1.LabelClusterUID: "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
							v1alpha1.LabelSeverity:   string(v1alpha1.SeverityMedium),
							v1alpha1.LabelIssueID:    "M-113",
							v1alpha1.LabelCategory:   "Security",
							v1alpha1.LabelPlugin:     "marvin",
							v1alpha1.LabelCustom:     "false",
						},
					},
					Spec: v1alpha1.ClusterIssueSpec{
						Cluster:  "cluster",
						ID:       "M-113",
						Message:  "Container could be running as root user",
						Severity: v1alpha1.SeverityMedium,
						Category: "Security",
						Resources: map[string][]string{
							"apps/v1/deployments": {"httpbin/httpbin"},
							"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878", "httpbin/httpbin-6089d0e989"},
						},
						Url: "https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted",
					},
				},
				{
					TypeMeta: clusterIssueTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-m-115-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-marvin-28140229",
								UID:        "50c8957e-c9e1-493a-9fa4-d0786deea017",
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:     "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster:    "cluster",
							v1alpha1.LabelClusterUID: "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
							v1alpha1.LabelSeverity:   string(v1alpha1.SeverityLow),
							v1alpha1.LabelIssueID:    "M-115",
							v1alpha1.LabelCategory:   "Security",
							v1alpha1.LabelPlugin:     "marvin",
							v1alpha1.LabelCustom:     "false",
						},
					},
					Spec: v1alpha1.ClusterIssueSpec{
						Cluster:  "cluster",
						ID:       "M-115",
						Message:  "Not allowed seccomp profile",
						Severity: v1alpha1.SeverityLow,
						Category: "Security",
						Resources: map[string][]string{
							"apps/v1/deployments": {"httpbin/httpbin"},
							"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
						},
						Url: "https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted",
					},
				},
				{
					TypeMeta: clusterIssueTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-m-202-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-marvin-28140229",
								UID:        "50c8957e-c9e1-493a-9fa4-d0786deea017",
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:     "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster:    "cluster",
							v1alpha1.LabelClusterUID: "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
							v1alpha1.LabelSeverity:   string(v1alpha1.SeverityLow),
							v1alpha1.LabelIssueID:    "M-202",
							v1alpha1.LabelCategory:   "Security",
							v1alpha1.LabelPlugin:     "marvin",
							v1alpha1.LabelCustom:     "false",
						},
					},
					Spec: v1alpha1.ClusterIssueSpec{
						Cluster:  "cluster",
						ID:       "M-202",
						Message:  "Automounted service account token",
						Severity: v1alpha1.SeverityLow,
						Category: "Security",
						Resources: map[string][]string{
							"apps/v1/deployments": {"httpbin/httpbin"},
							"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
						},
						Url: "https://microsoft.github.io/Threat-Matrix-for-Kubernetes/mitigations/MS-M9025%20Disable%20Service%20Account%20Auto%20Mount/",
					},
				},
				{
					TypeMeta: clusterIssueTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-m-300-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-marvin-28140229",
								UID:        "50c8957e-c9e1-493a-9fa4-d0786deea017",
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:     "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster:    "cluster",
							v1alpha1.LabelClusterUID: "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
							v1alpha1.LabelSeverity:   string(v1alpha1.SeverityLow),
							v1alpha1.LabelIssueID:    "M-300",
							v1alpha1.LabelCategory:   "Security",
							v1alpha1.LabelPlugin:     "marvin",
							v1alpha1.LabelCustom:     "false",
						},
					},
					Spec: v1alpha1.ClusterIssueSpec{
						Cluster:  "cluster",
						ID:       "M-300",
						Message:  "Root filesystem write allowed",
						Severity: v1alpha1.SeverityLow,
						Category: "Security",
						Resources: map[string][]string{
							"apps/v1/deployments": {"httpbin/httpbin"},
							"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
						},
						Url: "https://media.defense.gov/2022/Aug/29/2003066362/-1/-1/0/CTR_KUBERNETES_HARDENING_GUIDANCE_1.2_20220829.PDF#page=50",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r io.Reader
			if tt.args.filename != "" {
				f, err := os.Open(tt.args.filename)
				if err != nil {
					t.Errorf("parseMisconfigResults() setup error = %v", err)
					return
				}
				r = f
			}
			got, err := parseMisconfigResults(context.TODO(), tt.args.cfg, r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMisconfigResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sortClusterIssues(got)
			sortClusterIssues(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMisconfigResults() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func sortClusterIssues(issues []v1alpha1.ClusterIssue) {
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Spec.ID > issues[j].Spec.ID
	})
	for i := 0; i < len(issues); i++ {
		for r := range issues[i].Spec.Resources {
			sort.Strings(issues[i].Spec.Resources[r])
		}
	}
}
