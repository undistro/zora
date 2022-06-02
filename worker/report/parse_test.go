package report

import (
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"

	inspectv1a1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/worker/config"
	"github.com/getupio-undistro/inspect/worker/report/popeye"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestParse(t *testing.T) {
	cases := []struct {
		description   string
		testrep       io.Reader
		config        *config.Config
		clusterissues []*inspectv1a1.ClusterIssue
		toerr         bool
	}{
		{
			description: "Single <ClusterIssue> instance with many resources",
			testrep:     strings.NewReader(popeye.TestReport1),
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "popeye",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job_id",
				JobUID:          "fake_job_uid-666-666",
				Pod:             "_",
			},
			clusterissues: []*inspectv1a1.ClusterIssue{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: inspectv1a1.SchemeGroupVersion.String(),
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
							inspectv1a1.LabelScanID:        "fake_job_uid-666-666",
							inspectv1a1.LabelCluster:       "fake_cluster",
							inspectv1a1.LabelSeverity:      "Low",
							inspectv1a1.LabelIssueID:       "POP-400",
							inspectv1a1.LabelIssueCategory: "clusterroles",
							inspectv1a1.LabelPlugin:        "popeye",
						},
					},
					Spec: inspectv1a1.ClusterIssueSpec{
						ID:       "POP-400",
						Message:  "Used? Unable to locate resource reference",
						Severity: inspectv1a1.ClusterIssueSeverity("Low"),
						Category: "clusterroles",
						Resources: map[string][]string{
							"rbac.authorization.k8s.io/v1/clusterroles": {
								"capi-kubeadm-control-plane-manager-role",
								"cert-manager-edit",
								"system:certificates.k8s.io:kube-apiserver-client-kubelet-approver",
								"system:persistent-volume-provisioner",
								"undistro-metrics-reader",
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
					},
				},
			},
			toerr: false,
		},

		{
			description: "Four <ClusterIssue> instances with many resources",
			testrep:     strings.NewReader(popeye.TestReport2),
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "popeye",
				Cluster:         "super_fake_cluster",
				ClusterIssuesNs: "super_fake_ns",
				Job:             "super_fake_job_id",
				JobUID:          "super_fake_job_uid-666-666",
				Pod:             "_",
			},
			clusterissues: []*inspectv1a1.ClusterIssue{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: inspectv1a1.SchemeGroupVersion.String(),
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
							inspectv1a1.LabelScanID:        "super_fake_job_uid-666-666",
							inspectv1a1.LabelCluster:       "super_fake_cluster",
							inspectv1a1.LabelSeverity:      "Low",
							inspectv1a1.LabelIssueID:       "POP-400",
							inspectv1a1.LabelIssueCategory: "clusterroles",
							inspectv1a1.LabelPlugin:        "popeye",
						},
					},
					Spec: inspectv1a1.ClusterIssueSpec{
						ID:       "POP-400",
						Message:  "Used? Unable to locate resource reference",
						Severity: inspectv1a1.ClusterIssueSeverity("Low"),
						Category: "clusterroles",
						Resources: map[string][]string{
							"rbac.authorization.k8s.io/v1/clusterroles": {"system:node-bootstrapper", "undistro-metrics-reader"},
						},
						TotalResources: 2,
						Cluster:        "super_fake_cluster",
					},
				},

				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: inspectv1a1.SchemeGroupVersion.String(),
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
							inspectv1a1.LabelScanID:        "super_fake_job_uid-666-666",
							inspectv1a1.LabelCluster:       "super_fake_cluster",
							inspectv1a1.LabelSeverity:      "Medium",
							inspectv1a1.LabelIssueID:       "POP-106",
							inspectv1a1.LabelIssueCategory: "daemonsets",
							inspectv1a1.LabelPlugin:        "popeye",
						},
					},
					Spec: inspectv1a1.ClusterIssueSpec{
						ID:       "POP-106",
						Message:  "No resources requests/limits defined",
						Severity: inspectv1a1.ClusterIssueSeverity("Medium"),
						Category: "daemonsets",
						Resources: map[string][]string{
							"containers": {"kube-system/aws-node", "cert-manager/cert-manager"},
						},
						TotalResources: 2,
						Cluster:        "super_fake_cluster",
					},
				},

				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: inspectv1a1.SchemeGroupVersion.String(),
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
							inspectv1a1.LabelScanID:        "super_fake_job_uid-666-666",
							inspectv1a1.LabelCluster:       "super_fake_cluster",
							inspectv1a1.LabelSeverity:      "Medium",
							inspectv1a1.LabelIssueID:       "POP-107",
							inspectv1a1.LabelIssueCategory: "daemonsets",
							inspectv1a1.LabelPlugin:        "popeye",
						},
					},
					Spec: inspectv1a1.ClusterIssueSpec{
						ID:       "POP-107",
						Message:  "No resource limits defined",
						Severity: inspectv1a1.ClusterIssueSeverity("Medium"),
						Category: "daemonsets",
						Resources: map[string][]string{
							"containers": {"kube-system/aws-node", "kube-system/kube-proxy"},
						},
						TotalResources: 2,
						Cluster:        "super_fake_cluster",
					},
				},

				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterIssue",
						APIVersion: inspectv1a1.SchemeGroupVersion.String(),
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
							inspectv1a1.LabelScanID:        "super_fake_job_uid-666-666",
							inspectv1a1.LabelCluster:       "super_fake_cluster",
							inspectv1a1.LabelSeverity:      "Low",
							inspectv1a1.LabelIssueID:       "POP-108",
							inspectv1a1.LabelIssueCategory: "deployments",
							inspectv1a1.LabelPlugin:        "popeye",
						},
					},
					Spec: inspectv1a1.ClusterIssueSpec{
						ID:       "POP-108",
						Message:  "Unnamed port",
						Severity: inspectv1a1.ClusterIssueSeverity("Low"),
						Category: "deployments",
						Resources: map[string][]string{
							"containers": {"cert-manager/cert-manager"},
						},
						TotalResources: 1,
						Cluster:        "super_fake_cluster",
					},
				},
			},
			toerr: false,
		},

		{
			description: "Invalid Popeye report",
			testrep:     strings.NewReader(popeye.TestReport3),
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "popeye",
				Cluster:         "_",
				ClusterIssuesNs: "_",
				Job:             "_",
				JobUID:          "fake_job_uid-666-666",
				Pod:             "_",
			},
			clusterissues: nil,
			toerr:         true,
		},
		{
			description: "Empty Popeye report",
			testrep:     strings.NewReader(popeye.TestReport4),
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "popeye",
				Cluster:         "_",
				ClusterIssuesNs: "_",
				Job:             "_",
				JobUID:          "fake_job_uid-666-666",
				Pod:             "_",
			},
			clusterissues: nil,
			toerr:         true,
		},
		{
			description: "Invalid plugin",
			testrep:     strings.NewReader(popeye.TestReport4),
			config: &config.Config{
				DonePath:        "_",
				ErrorPath:       "_",
				Plugin:          "fake_plugin",
				Cluster:         "_",
				ClusterIssuesNs: "_",
				Job:             "_",
				JobUID:          "fake_job_uid-666-666",
				Pod:             "_",
			},
			clusterissues: nil,
			toerr:         true,
		},
	}

	sfun := func(ciarr []*inspectv1a1.ClusterIssue) {
		sort.Slice(ciarr, func(i, j int) bool {
			return ciarr[i].Spec.ID > ciarr[j].Spec.ID
		})
		for c := 0; c < len(ciarr); c++ {
			for r, _ := range ciarr[c].Spec.Resources {
				sort.Strings(ciarr[c].Spec.Resources[r])
			}
		}
	}
	for _, c := range cases {
		ciarr, err := Parse(c.testrep, c.config)
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
