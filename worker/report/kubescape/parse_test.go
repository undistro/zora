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

package kubescape

import (
	"os"
	"sort"
	"testing"

	zorav1a1 "github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
)

func TestScoreFactorSeverity(t *testing.T) {
	cases := []struct {
		description string
		scorefactor float32
		severity    zorav1a1.ClusterIssueSeverity
	}{
		{
			description: "Negative score factor",
			scorefactor: -1.0,
			severity:    zorav1a1.SeverityUnknown,
		},
		{
			description: "Undefined score factor (uses default type value)",
			scorefactor: 0,
			severity:    zorav1a1.SeverityUnknown,
		},
		{
			description: "Tiny score factor",
			scorefactor: 0.59,
			severity:    zorav1a1.SeverityUnknown,
		},
		{
			description: "Exact match of score factor for Low severity",
			scorefactor: 1,
			severity:    zorav1a1.SeverityLow,
		},
		{
			description: "Approximate mid range score factor between Low and Medium severities",
			scorefactor: 2.305,
			severity:    zorav1a1.SeverityLow,
		},
		{
			description: "Score factor at border of Low to Medium severities",
			scorefactor: 3.999,
			severity:    zorav1a1.SeverityLow,
		},
		{
			description: "Exact match of score factor for Medium severity",
			scorefactor: 4,
			severity:    zorav1a1.SeverityMedium,
		},
		{
			description: "Approximate mid range score factor between Medium and High severities",
			scorefactor: 5.00139,
			severity:    zorav1a1.SeverityMedium,
		},
		{
			description: "Score factor at border of Medium to High severities",
			scorefactor: 6.999,
			severity:    zorav1a1.SeverityMedium,
		},
		{
			description: "Match of score factor for High severity",
			scorefactor: 7.0001,
			severity:    zorav1a1.SeverityHigh,
		},
		{
			description: "Critical Severity on Kubescape, but High on Zora",
			scorefactor: 19,
			severity:    zorav1a1.SeverityHigh,
		},
		{
			description: "Big score factor",
			scorefactor: 88.454601,
			severity:    zorav1a1.SeverityHigh,
		},
	}

	for _, c := range cases {
		if s := ScoreFactorSeverity(c.scorefactor); s != c.severity {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Expected <%s> but got <%s>\n", c.severity, s)
		}
	}

}

func TestExtractGvrAndInstanceName(t *testing.T) {
	cases := []struct {
		description string
		obj         map[string]interface{}
		gvr         string
		name        string
		toerr       bool
	}{
		{
			description: "Full GVR for Daemonset",
			obj: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "DaemonSet",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"deprecated.daemonset.template.generation": "1",
					},
					"creationTimestamp": "2022-07-12T14:03:21Z",
					"generation":        1,
					"labels": map[string]interface{}{
						"addonmanager.kubernetes.io/mode": "Reconcile",
						"component":                       "gke-metrics-agent",
						"k8s-app":                         "gke-metrics-agent",
					},
					"name":            "gke-metrics-agent",
					"namespace":       "kube-system",
					"resourceVersion": "821",
					"uid":             "8ddcc89a-aca0-4b25-b87f-ed8a4961c554",
				},
			},
			gvr:  "apps/v1/daemonset",
			name: "kube-system/gke-metrics-agent",
		},
		{
			description: "Only version and resource returned for GVR of Service Account",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"creationTimestamp": "2022-07-12T14:02:52Z",
					"name":              "resourcequota-controller",
					"namespace":         "kube-system",
					"resourceVersion":   "245",
					"uid":               "36191664-8277-4bc1-b71e-7c1f19ccb832",
				},
			},
			gvr:  "v1/serviceaccount",
			name: "kube-system/resourcequota-controller",
		},
		{
			description: "Only version and resource returned for GVR of Namespace",
			obj: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"creationTimestamp": "2022-07-12T14:02:46Z",
					"labels": map[string]interface{}{
						"kubernetes.io/metadata.name": "kube-system",
					},
					"name":            "kube-system",
					"resourceVersion": "15",
					"uid":             "a117aaf0-8fbf-4c95-a576-aa543e7c915a",
				},
			},
			gvr:  "v1/namespace",
			name: "kube-system",
		},
		{
			description: "Use related object's GVR and name for GKE's User kind",
			obj: map[string]interface{}{
				"apiGroup": "rbac.authorization.k8s.io",
				"kind":     "User",
				"name":     "system:kubestore-collector",
				"relatedObjects": []interface{}{
					map[string]interface{}{
						"apiVersion": "rbac.authorization.k8s.io/v1",
						"kind":       "ClusterRoleBinding",
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"components.gke.io/component-version": "kubestore-collector-rbac-1.0.0",
							},
							"creationTimestamp": "2022-07-12T14:03:21Z",
							"labels": map[string]interface{}{
								"addonmanager.kubernetes.io/mode": "Reconcile",
							},
							"name":            "system:kubestore-collector",
							"resourceVersion": "550",
							"uid":             "3ee4eae8-2889-48d2-a057-f1046c5770a5",
						},
						"roleRef": map[string]interface{}{
							"apiGroup": "rbac.authorization.k8s.io",
							"kind":     "ClusterRole",
							"name":     "system:kubestore-collector",
						},
						"subjects": []map[string]interface{}{
							{
								"apiGroup": "rbac.authorization.k8s.io",
								"kind":     "User",
								"name":     "system:kubestore-collector",
							},
						},
					},
					map[string]interface{}{
						"apiVersion": "rbac.authorization.k8s.io/v1",
						"kind":       "ClusterRole",
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"components.gke.io/component-version": "kubestore-collector-rbac-1.0.0",
							},
							"creationTimestamp": "2022-07-12T14:03:21Z",
							"labels": map[string]interface{}{
								"addonmanager.kubernetes.io/mode": "Reconcile",
							},
							"name":            "system:kubestore-collector",
							"resourceVersion": "548",
							"uid":             "712eb3e6-6e1a-427b-8cc4-d3afcb07e995",
						},
					},
				},
			},
			gvr:  "rbac.authorization.k8s.io/v1/clusterrolebinding",
			name: "system:kubestore-collector",
		},
		{
			description: "Record without GVK data",
			obj: map[string]interface{}{
				"metadata": map[string]interface{}{
					"creationTimestamp": "2022-07-12T14:03:12Z",
					"labels": map[string]interface{}{
						"addonmanager.kubernetes.io/mode": "EnsureExists",
					},
					"name":            "kube-dns",
					"namespace":       "kube-system",
					"resourceVersion": "357",
					"uid":             "4792cc93-1a27-4ebd-8c82-9c043ca7380f",
				},
			},
			toerr: true,
		},
		{
			description: "Record with invalid GVK data types",
			obj: map[string]interface{}{
				"apiGroup":   -1,
				"apiVersion": -1,
				"kind":       -1,
				"metadata": map[string]interface{}{
					"creationTimestamp": "2022-07-12T14:03:12Z",
					"labels": map[string]interface{}{
						"addonmanager.kubernetes.io/mode": "EnsureExists",
					},
					"name":            "kube-dns",
					"namespace":       "kube-system",
					"resourceVersion": "357",
					"uid":             "4792cc93-1a27-4ebd-8c82-9c043ca7380f",
				},
			},
			toerr: true,
		},
		{
			description: "Record with invalid data type of related object",
			obj: map[string]interface{}{
				"kind":           "ServiceAccount",
				"name":           "cloud-provider",
				"namespace":      "kube-system",
				"relatedObjects": []interface{}{1, 2, 3},
			},
			toerr: true,
		},
	}

	for _, c := range cases {
		if gvr, name, err := ExtractGvrAndInstanceName(logr.Discard(), c.obj); gvr != c.gvr || name != c.name || ((err != nil) != c.toerr) {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Expected:\n\t{gvr: <%s>, name: <%s>, toerr: <%t>}\nBut got:\n\t{gvr: <%s>, name: <%s>, err: <%v>}\n", c.gvr, c.name, c.toerr, gvr, name, err)
		}
	}
}

func TestExtractStatus(t *testing.T) {
	cases := []struct {
		description string
		control     *ResourceAssociatedControl
		status      ScanningStatus
	}{
		{
			description: "Prioritary status Error",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusError},
					{Status: StatusPassed},
					{Status: StatusPassed},
				},
			},
			status: StatusError,
		},
		{
			description: "Prioritary status Error over Skipped",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusError},
					{Status: StatusPassed},
					{Status: StatusSkipped},
				},
			},
			status: StatusError,
		},
		{
			description: "Prioritary status Failed",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusFailed},
					{Status: StatusPassed},
					{Status: StatusPassed},
				},
			},
			status: StatusFailed,
		},
		{
			description: "Prioritary status Failed over Skipped",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusFailed},
					{Status: StatusPassed},
					{Status: StatusSkipped},
				},
			},
			status: StatusFailed,
		},
		{
			description: "Prioritary status Unknown",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusUnknown},
					{Status: StatusPassed},
					{Status: StatusPassed},
				},
			},
			status: StatusUnknown,
		},
		{
			description: "Prioritary status Unknown over Skipped",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusUnknown},
					{Status: StatusPassed},
					{Status: StatusSkipped},
				},
			},
			status: StatusUnknown,
		},
		{
			description: "Prioritary status Irrelevant",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusIrrelevant},
					{Status: StatusPassed},
					{Status: StatusPassed},
				},
			},
			status: StatusIrrelevant,
		},
		{
			description: "Prioritary status Irrelevant over Skipped",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusIrrelevant},
					{Status: StatusPassed},
					{Status: StatusSkipped},
				},
			},
			status: StatusIrrelevant,
		},
		{
			description: "Prioritary status Error",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusError},
					{Status: StatusPassed},
					{Status: StatusPassed},
				},
			},
			status: StatusError,
		},
		{
			description: "Prioritary status Error over Skipped",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusError},
					{Status: StatusPassed},
					{Status: StatusSkipped},
				},
			},
			status: StatusError,
		},
		{
			description: "Prioritary status Failed over Error",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusError},
					{Status: StatusFailed},
					{Status: StatusPassed},
				},
			},
			status: StatusFailed,
		},
		{
			description: "Prioritary status Failed over Error and Unknown",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusError},
					{Status: StatusUnknown},
					{Status: StatusFailed},
				},
			},
			status: StatusFailed,
		},
		{
			description: "Majority Passed status",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusPassed},
					{Status: StatusPassed},
					{Status: StatusExcluded},
				},
			},
			status: StatusPassed,
		},
		{
			description: "Majority Excluded status",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusPassed},
					{Status: StatusExcluded},
					{Status: StatusExcluded},
				},
			},
			status: StatusExcluded,
		},
		{
			description: "Majority Skipped status",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{
					{Status: StatusExcluded},
					{Status: StatusSkipped},
					{Status: StatusSkipped},
				},
			},
			status: StatusSkipped,
		},
		{
			description: "Empty status on record leads to Unknown",
			control: &ResourceAssociatedControl{
				ResourceAssociatedRules: []ResourceAssociatedRule{},
			},
			status: StatusUnknown,
		},
	}

	for _, c := range cases {
		if s := ExtractStatus(c.control); s != c.status {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Expected <%s> but got <%s>\n", c.status, s)
		}
	}
}

func TestPreprocessResources(t *testing.T) {
	cases := []struct {
		description string
		report      *PostureReport
		rmap        map[string]map[string]interface{}
		toerr       bool
	}{
		{
			description: "Empty report",
			report:      &PostureReport{},
			rmap:        map[string]map[string]interface{}{},
		},
		{
			description: "Report with 1 record",
			report: &PostureReport{
				Resources: []Resource{
					{
						ResourceID: "/v1/kube-system/ServiceAccount/resourcequota-controller",
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ServiceAccount",
							"metadata": map[string]interface{}{
								"creationTimestamp": "2022-07-12T14:02:52Z",
								"name":              "resourcequota-controller",
								"namespace":         "kube-system",
								"resourceVersion":   "245",
								"uid":               "36191664-8277-4bc1-b71e-7c1f19ccb832",
							},
							"secrets": []map[string]interface{}{
								{
									"name": "resourcequota-controller-token-9fvh4",
								},
							},
						},
					},
				},
			},
			rmap: map[string]map[string]interface{}{
				"/v1/kube-system/ServiceAccount/resourcequota-controller": map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ServiceAccount",
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-07-12T14:02:52Z",
						"name":              "resourcequota-controller",
						"namespace":         "kube-system",
						"resourceVersion":   "245",
						"uid":               "36191664-8277-4bc1-b71e-7c1f19ccb832",
					},
					"secrets": []map[string]interface{}{
						{
							"name": "resourcequota-controller-token-9fvh4",
						},
					},
				},
			},
		},
		{
			description: "Report with 3 records",
			report: &PostureReport{
				Resources: []Resource{
					{
						ResourceID: "/v1/kube-system/ServiceAccount/resourcequota-controller",
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ServiceAccount",
							"metadata": map[string]interface{}{
								"creationTimestamp": "2022-07-12T14:02:52Z",
								"name":              "resourcequota-controller",
								"namespace":         "kube-system",
								"resourceVersion":   "245",
								"uid":               "36191664-8277-4bc1-b71e-7c1f19ccb832",
							},
							"secrets": []map[string]interface{}{
								{
									"name": "resourcequota-controller-token-9fvh4",
								},
							},
						},
					},
					{
						ResourceID: "/v1/kube-system/ServiceAccount/statefulset-controller",
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ServiceAccount",
							"metadata": map[string]interface{}{
								"creationTimestamp": "2022-07-12T14:02:52Z",
								"name":              "statefulset-controller",
								"namespace":         "kube-system",
								"resourceVersion":   "225",
								"uid":               "f1dbc7ab-4e87-4b96-b10f-805377997d08",
							},
							"secrets": []map[string]interface{}{
								{
									"name": "statefulset-controller-token-pxdsk",
								},
							},
						},
					},
					{
						ResourceID: "/v1/zora-system/ConfigMap/kube-root-ca.crt",
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"data": map[string]interface{}{
								"ca.crt": "XXXXXX",
							},
							"kind": "ConfigMap",
							"metadata": map[string]interface{}{
								"creationTimestamp": "2022-07-12T14:03:53Z",
								"name":              "kube-root-ca.crt",
								"namespace":         "zora-system",
								"resourceVersion":   "888",
								"uid":               "b9304d53-6c4e-4d84-9df6-445425e8e1cb",
							},
						},
					},
				},
			},
			rmap: map[string]map[string]interface{}{
				"/v1/kube-system/ServiceAccount/resourcequota-controller": map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ServiceAccount",
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-07-12T14:02:52Z",
						"name":              "resourcequota-controller",
						"namespace":         "kube-system",
						"resourceVersion":   "245",
						"uid":               "36191664-8277-4bc1-b71e-7c1f19ccb832",
					},
					"secrets": []map[string]interface{}{
						{
							"name": "resourcequota-controller-token-9fvh4",
						},
					},
				},
				"/v1/kube-system/ServiceAccount/statefulset-controller": map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ServiceAccount",
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-07-12T14:02:52Z",
						"name":              "statefulset-controller",
						"namespace":         "kube-system",
						"resourceVersion":   "225",
						"uid":               "f1dbc7ab-4e87-4b96-b10f-805377997d08",
					},
					"secrets": []map[string]interface{}{
						{
							"name": "statefulset-controller-token-pxdsk",
						},
					},
				},
				"/v1/zora-system/ConfigMap/kube-root-ca.crt": map[string]interface{}{
					"apiVersion": "v1",
					"data": map[string]interface{}{
						"ca.crt": "XXXXXX",
					},
					"kind": "ConfigMap",
					"metadata": map[string]interface{}{
						"creationTimestamp": "2022-07-12T14:03:53Z",
						"name":              "kube-root-ca.crt",
						"namespace":         "zora-system",
						"resourceVersion":   "888",
						"uid":               "b9304d53-6c4e-4d84-9df6-445425e8e1cb",
					},
				},
			},
		},
		{
			description: "Record with non expected object type",
			report: &PostureReport{
				Resources: []Resource{
					{
						Object: "Fail",
					},
				},
			},
			toerr: true,
		},
	}

	for _, c := range cases {
		if rmap, err := PreprocessResources(c.report); !cmp.Equal(rmap, c.rmap) || ((err != nil) != c.toerr) {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Mismatch between expected and obtained values: \n%s\n", cmp.Diff(c.rmap, rmap))
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		description string
		testrepname string
		cispecs     []*zorav1a1.ClusterIssueSpec
		toerr       bool
	}{
		{
			description: "Single <ClusterIssueSpec> instance with many resources",
			testrepname: "testdata/test_report_1.json",
			cispecs: []*zorav1a1.ClusterIssueSpec{
				{
					ID:       "C-0001",
					Message:  "Forbidden Container Registries",
					Severity: "Medium",
					Category: "deployment",
					Resources: map[string][]string{
						"apps/v1/daemonset": []string{
							"kube-system/gke-metrics-agent",
							"kube-system/gke-metrics-agent-scaling-20",
							"kube-system/fluentbit-gke",
							"kube-system/kube-proxy",
							"kube-system/metadata-proxy-v0.1",
							"kube-system/nvidia-gpu-device-plugin",
							"kube-system/pdcsi-node-windows",
							"kube-system/gke-metrics-agent-scaling-10",
							"kube-system/gke-metrics-agent-windows",
							"kube-system/pdcsi-node",
						},
						"apps/v1/deployment": []string{
							"kube-system/konnectivity-agent",
							"kube-system/metrics-server-v0.4.5",
							"kube-system/kube-dns",
							"kube-system/event-exporter-gke",
							"kube-system/kube-dns-autoscaler",
							"kube-system/konnectivity-agent-autoscaler",
						},
						"v1/pod": []string{
							"kube-system/kube-proxy-gke-zora-jzapzzpr-default-pool-b0f7ab4a-sg6t",
						},
					},
					TotalResources: 17,
				},
			},
			toerr: false,
		},

		{
			description: "Four <ClusterIssueSpec> instance with many resources",
			testrepname: "testdata/test_report_2.json",
			cispecs: []*zorav1a1.ClusterIssueSpec{
				{
					ID:       "C-0004",
					Message:  "Resources memory limit and request",
					Severity: "High",
					Category: "daemonset",
					Resources: map[string][]string{
						"apps/v1/daemonset": []string{
							"kube-system/kube-proxy",
						},
						"apps/v1/deployment": []string{
							"kube-system/kube-dns",
						},
					},
					TotalResources: 2,
				},
				{
					ID:       "C-0006",
					Message:  "Allowed hostPath",
					Severity: "Medium",
					Category: "daemonset",
					Resources: map[string][]string{
						"apps/v1/daemonset": []string{
							"kube-system/fluentbit-gke",
							"kube-system/kube-proxy",
						},
					},
					TotalResources: 2,
				},
				{
					ID:       "C-0013",
					Message:  "Non-root containers",
					Severity: "Medium",
					Category: "daemonset",
					Resources: map[string][]string{
						"apps/v1/daemonset": []string{
							"kube-system/fluentbit-gke",
							"kube-system/kube-proxy",
						},
					},
					TotalResources: 2,
				},
				{
					ID:       "C-0017",
					Message:  "Immutable container filesystem",
					Severity: "Low",
					Category: "deployment",
					Resources: map[string][]string{
						"apps/v1/daemonset": []string{
							"kube-system/gke-metrics-agent",
						},
						"apps/v1/deployment": []string{
							"kube-system/konnectivity-agent",
						},
					},
					TotalResources: 2,
				},
			},
			toerr: false,
		},

		{
			description: "Invalid Kubespace report",
			testrepname: "testdata/test_report_3.json",
			cispecs:     nil,
			toerr:       true,
		},
		{
			description: "Empty Kubescape report",
			testrepname: "testdata/test_report_4.json",
			cispecs:     nil,
			toerr:       true,
		},
	}

	sfun := func(cis []*zorav1a1.ClusterIssueSpec) {
		sort.Slice(cis, func(i, j int) bool {
			return cis[i].ID > cis[j].ID
		})
		for c := 0; c < len(cis); c++ {
			for r, _ := range cis[c].Resources {
				sort.Strings(cis[c].Resources[r])
			}
		}
	}
	for _, c := range cases {
		rep, err := os.ReadFile(c.testrepname)
		if err != nil {
			t.Errorf("Setup failed on case: %s\n", c.description)
			t.Fatal(err)
		}
		cispecs, err := Parse(logr.Discard(), rep)
		sfun(c.cispecs)
		sfun(cispecs)
		if (err != nil) != c.toerr || !cmp.Equal(c.cispecs, cispecs) {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Mismatch between expected and obtained values: \n%s\n", cmp.Diff(c.cispecs, cispecs))
		}
	}
}
