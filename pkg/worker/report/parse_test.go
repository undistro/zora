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

package report

import (
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	zorav1a1 "github.com/undistro/zora/apis/zora/v1alpha1"
	"github.com/undistro/zora/pkg/worker/config"
)

func TestParse(t *testing.T) {
	cases := []struct {
		description   string
		testrepname   string
		config        *config.Config
		clusterissues []*zorav1a1.ClusterIssue
		toerr         bool
	}{

		// Popeye specific.
		{
			description: "Single Popeye <ClusterIssue> instance with many resources",
			testrepname: "popeye/testdata/test_report_1.json",
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "popeye",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job_id",
				JobUID:          "fake_job_uid-666-666",
			},
			clusterissues: []*zorav1a1.ClusterIssue{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: zorav1a1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fake_cluster-pop-400-666",
						Namespace: "fake_ns",
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: "batch/v1",
							Kind:       "Job",
							Name:       "fake_job_id",
							UID:        types.UID("fake_job_uid-666-666"),
						}},
						Labels: map[string]string{
							zorav1a1.LabelScanID:   "fake_job_uid-666-666",
							zorav1a1.LabelCluster:  "fake_cluster",
							zorav1a1.LabelSeverity: "Low",
							zorav1a1.LabelIssueID:  "POP-400",
							zorav1a1.LabelCategory: "General",
							zorav1a1.LabelPlugin:   "popeye",
						},
					},
					Spec: zorav1a1.ClusterIssueSpec{
						ID:       "POP-400",
						Message:  "Used? Unable to locate resource reference",
						Severity: zorav1a1.ClusterIssueSeverity("Low"),
						Category: "General",
						Resources: map[string][]string{
							"rbac.authorization.k8s.io/v1/clusterroles": {
								"capi-kubeadm-control-plane-manager-role",
								"cert-manager-edit",
								"system:certificates.k8s.io:kube-apiserver-client-kubelet-approver",
								"system:persistent-volume-provisioner",
								"metrics-reader",
								"cert-manager-view",
								"system:heapster",
								"system:kube-aggregator",
								"admin",
								"system:metrics-server-aggregated-reader",
								"system:node-bootstrapper",
								"system:node-problem-detector",
								"view",
								"capi-manager-role",
								"system:certificates.k8s.io:kubelet-serving-approver",
								"system:certificates.k8s.io:legacy-unknown-approver",
							},
						},
						TotalResources: 16,
						Cluster:        "fake_cluster",
						Url:            "",
					},
				},
			},
			toerr: false,
		},

		{
			description: "Four Popeye <ClusterIssue> instances with many resources",
			testrepname: "popeye/testdata/test_report_2.json",
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "popeye",
				Cluster:         "super_fake_cluster",
				ClusterIssuesNs: "super_fake_ns",
				Job:             "super_fake_job_id",
				JobUID:          "super_fake_job_uid-666-666",
			},
			clusterissues: []*zorav1a1.ClusterIssue{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: zorav1a1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "super_fake_cluster-pop-400-666",
						Namespace: "super_fake_ns",
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: "batch/v1",
							Kind:       "Job",
							Name:       "super_fake_job_id",
							UID:        types.UID("super_fake_job_uid-666-666"),
						}},
						Labels: map[string]string{
							zorav1a1.LabelScanID:   "super_fake_job_uid-666-666",
							zorav1a1.LabelCluster:  "super_fake_cluster",
							zorav1a1.LabelSeverity: "Low",
							zorav1a1.LabelIssueID:  "POP-400",
							zorav1a1.LabelCategory: "General",
							zorav1a1.LabelPlugin:   "popeye",
						},
					},
					Spec: zorav1a1.ClusterIssueSpec{
						ID:       "POP-400",
						Message:  "Used? Unable to locate resource reference",
						Severity: zorav1a1.ClusterIssueSeverity("Low"),
						Category: "General",
						Resources: map[string][]string{
							"rbac.authorization.k8s.io/v1/clusterroles": {"system:node-bootstrapper", "metrics-reader"},
						},
						TotalResources: 2,
						Cluster:        "super_fake_cluster",
						Url:            "",
					},
				},

				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: zorav1a1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "super_fake_cluster-pop-106-666",
						Namespace: "super_fake_ns",
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: "batch/v1",
							Kind:       "Job",
							Name:       "super_fake_job_id",
							UID:        types.UID("super_fake_job_uid-666-666"),
						}},
						Labels: map[string]string{
							zorav1a1.LabelScanID:   "super_fake_job_uid-666-666",
							zorav1a1.LabelCluster:  "super_fake_cluster",
							zorav1a1.LabelSeverity: "Medium",
							zorav1a1.LabelIssueID:  "POP-106",
							zorav1a1.LabelCategory: "Container",
							zorav1a1.LabelPlugin:   "popeye",
						},
					},
					Spec: zorav1a1.ClusterIssueSpec{
						ID:       "POP-106",
						Message:  "No resources requests/limits defined",
						Severity: zorav1a1.ClusterIssueSeverity("Medium"),
						Category: "Container",
						Resources: map[string][]string{
							"apps/v1/daemonsets":  {"kube-system/aws-node"},
							"apps/v1/deployments": {"cert-manager/cert-manager"},
						},
						TotalResources: 2,
						Cluster:        "super_fake_cluster",
						Url:            "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
					},
				},

				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: zorav1a1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "super_fake_cluster-pop-107-666",
						Namespace: "super_fake_ns",
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: "batch/v1",
							Kind:       "Job",
							Name:       "super_fake_job_id",
							UID:        types.UID("super_fake_job_uid-666-666"),
						}},
						Labels: map[string]string{
							zorav1a1.LabelScanID:   "super_fake_job_uid-666-666",
							zorav1a1.LabelCluster:  "super_fake_cluster",
							zorav1a1.LabelSeverity: "Medium",
							zorav1a1.LabelIssueID:  "POP-107",
							zorav1a1.LabelCategory: "Container",
							zorav1a1.LabelPlugin:   "popeye",
						},
					},
					Spec: zorav1a1.ClusterIssueSpec{
						ID:       "POP-107",
						Message:  "No resource limits defined",
						Severity: zorav1a1.ClusterIssueSeverity("Medium"),
						Category: "Container",
						Resources: map[string][]string{
							"apps/v1/daemonsets": {"kube-system/aws-node", "kube-system/kube-proxy"},
						},
						TotalResources: 2,
						Cluster:        "super_fake_cluster",
						Url:            "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/",
					},
				},

				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: zorav1a1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "super_fake_cluster-pop-108-666",
						Namespace: "super_fake_ns",
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: "batch/v1",
							Kind:       "Job",
							Name:       "super_fake_job_id",
							UID:        types.UID("super_fake_job_uid-666-666"),
						}},
						Labels: map[string]string{
							zorav1a1.LabelScanID:   "super_fake_job_uid-666-666",
							zorav1a1.LabelCluster:  "super_fake_cluster",
							zorav1a1.LabelSeverity: "Low",
							zorav1a1.LabelIssueID:  "POP-108",
							zorav1a1.LabelCategory: "Container",
							zorav1a1.LabelPlugin:   "popeye",
						},
					},
					Spec: zorav1a1.ClusterIssueSpec{
						ID:       "POP-108",
						Message:  "Unnamed port",
						Severity: zorav1a1.ClusterIssueSeverity("Low"),
						Category: "Container",
						Resources: map[string][]string{
							"apps/v1/deployments": {"cert-manager/cert-manager"},
						},
						TotalResources: 1,
						Cluster:        "super_fake_cluster",
						Url:            "",
					},
				},
			},
			toerr: false,
		},

		{
			description: "Invalid Popeye report",
			testrepname: "popeye/testdata/test_report_3.json",
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "popeye",
				Cluster:         "_",
				ClusterIssuesNs: "_",
				Job:             "_",
				JobUID:          "fake_job_uid-666-666",
			},
			clusterissues: nil,
			toerr:         true,
		},
		{
			description: "Empty Popeye report",
			testrepname: "popeye/testdata/test_report_4.json",
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "popeye",
				Cluster:         "_",
				ClusterIssuesNs: "_",
				Job:             "_",
				JobUID:          "fake_job_uid-666-666",
			},
			clusterissues: nil,
			toerr:         true,
		},

		// Generic.
		{
			description: "Invalid plugin",
			testrepname: "popeye/testdata/test_report_4.json",
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "fake_plugin",
				Cluster:         "_",
				ClusterIssuesNs: "_",
				Job:             "_",
				JobUID:          "fake_job_uid-666-666",
			},
			clusterissues: nil,
			toerr:         true,
		},
	}

	sfun := func(ciarr []*zorav1a1.ClusterIssue) {
		sort.Slice(ciarr, func(i, j int) bool {
			return ciarr[i].Spec.ID > ciarr[j].Spec.ID
		})
		for c := 0; c < len(ciarr); c++ {
			for r := range ciarr[c].Spec.Resources {
				sort.Strings(ciarr[c].Spec.Resources[r])
			}
		}
	}
	for _, c := range cases {
		fid, err := os.Open(c.testrepname)
		if err != nil {
			t.Errorf("Setup failed on case: %s\n", c.description)
			t.Fatal(err)
		}
		ciarr, err := Parse(logr.Discard(), fid, c.config)
		sfun(c.clusterissues)
		sfun(ciarr)
		if (err != nil) != c.toerr || !reflect.DeepEqual(c.clusterissues, ciarr) {
			if err != nil {
				t.Error(err)
			}
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Mismatch between expected and obtained values: \n%s\n", cmp.Diff(c.clusterissues, ciarr))
		}
	}
}
