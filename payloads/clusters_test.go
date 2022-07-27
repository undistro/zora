package payloads

import (
	"testing"
	"time"

	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/getupio-undistro/zora/pkg/apis"
	"github.com/getupio-undistro/zora/pkg/discovery"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeriveStatus(t *testing.T) {
	cases := []struct {
		description string
		conds       []metav1.Condition
		cl          *Cluster
	}{
		{
			description: "Cluster connected, discovered and scanned",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected, version 1.19",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterInfoDiscovered,
					Message: "cluster info successfully discovered",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterScanned,
					Message: "cluster successfully scanned",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{Connected: true},
				Scan:       ScanStatus{Status: Scanned},
			},
		},
		{
			description: "Cluster connected, discovered and not scanned due to error",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterInfoDiscovered,
					Message: "cluster info successfully discovered",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterScanListError,
					Message: "error fetching <ClusterScan> instances",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{Connected: true},
				Scan: ScanStatus{
					Status:  Failed,
					Message: "error fetching <ClusterScan> instances",
				},
			},
		},
		{
			description: "Cluster connected and discovered. But errors when listing scans",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterInfoDiscovered,
					Message: "cluster info successfully discovered",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterScanListError,
					Message: "error fetching <ClusterScan> instances",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{Connected: true},
				Scan: ScanStatus{
					Status:  Failed,
					Message: "error fetching <ClusterScan> instances",
				},
			},
		},
		{
			description: "Cluster connected and discovered without configured scans",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterInfoDiscovered,
					Message: "cluster info successfully discovered",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterScanNotConfigured,
					Message: "no scan configured",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{Connected: true},
				Scan: ScanStatus{
					Status:  Unknown,
					Message: "no scan configured",
				},
			},
		},
		{
			description: "Cluster connected and discovered without finished scans",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterInfoDiscovered,
					Message: "cluster info successfully discovered",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterNotScanned,
					Message: "no finished scan yet",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{Connected: true},
				Scan: ScanStatus{
					Status:  Unknown,
					Message: "no finished scan yet",
				},
			},
		},
		{
			description: "Cluster connected and discovered with failed scans",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterInfoDiscovered,
					Message: "cluster info successfully discovered",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterScanFailed,
					Message: "last scan failed",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{Connected: true},
				Scan: ScanStatus{
					Status:  Failed,
					Message: "last scan failed",
				},
			},
		},
		{
			description: "Cluster connected and scanned but not discovered",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterInfoNotDiscovered,
					Message: "metrics server not reachable",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterScanned,
					Message: "cluster successfully scanned",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{
					Connected: true,
					Message:   "metrics server not reachable",
				},
				Scan: ScanStatus{Status: Scanned},
			},
		},
		{
			description: "Cluster connected but not discovered nor scanned due to error",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterInfoNotDiscovered,
					Message: "metrics server not reachable",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterScanFailed,
					Message: "last scan failed",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{
					Connected: true,
					Message:   "metrics server not reachable",
				},
				Scan: ScanStatus{
					Status:  Failed,
					Message: "last scan failed",
				},
			},
		},
		{
			description: "Cluster connected but not discovered without finished scans",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterInfoNotDiscovered,
					Message: "metrics server not reachable",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterNotScanned,
					Message: "no finished scan yet",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{
					Connected: true,
					Message:   "metrics server not reachable",
				},
				Scan: ScanStatus{
					Status:  Unknown,
					Message: "no finished scan yet",
				},
			},
		},
		{
			description: "Cluster connected but not discovered without configured scans",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterInfoNotDiscovered,
					Message: "metrics server not reachable",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterScanNotConfigured,
					Message: "no scan configured",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{
					Connected: true,
					Message:   "metrics server not reachable",
				},
				Scan: ScanStatus{
					Status:  Unknown,
					Message: "no scan configured",
				},
			},
		},
		{
			description: "Cluster disconnected without configured scans",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionTrue,
					Reason:  v1alpha1.ClusterConnected,
					Message: "cluster successfully connected",
				},
				{
					Type:    v1alpha1.ClusterDiscovered,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterInfoNotDiscovered,
					Message: "metrics server not reachable",
				},
				{
					Type:    v1alpha1.ClusterScanned,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterScanNotConfigured,
					Message: "no scan configured",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{
					Connected: true,
					Message:   "metrics server not reachable",
				},
				Scan: ScanStatus{
					Status:  Unknown,
					Message: "no scan configured",
				},
			},
		},
		{
			description: "Cluster disconnected with valid Kubeconfig",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.ClusterNotConnected,
					Message: "unauthorized access",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{Message: "unauthorized access"},
				Scan:       ScanStatus{Status: Unknown},
			},
		},
		{
			description: "Cluster disconnected with invalid Kubeconfig",
			conds: []metav1.Condition{
				{
					Type:    v1alpha1.ClusterReady,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.KubeconfigError,
					Message: "invalid configuration",
				},
			},
			cl: &Cluster{
				Connection: ConnectionStatus{Message: "invalid configuration"},
				Scan:       ScanStatus{Status: Unknown},
			},
		},
		{
			description: "Cluster controller error leads to empty conditions",
			conds:       []metav1.Condition{},
			cl:          &Cluster{Scan: ScanStatus{Status: Unknown}},
		},
	}

	for _, c := range cases {
		cl := &Cluster{}
		if deriveStatus(c.conds, cl); !cmp.Equal(cl, c.cl) {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Mismatch between expected and obtained results:\n%s", cmp.Diff(c.cl, cl))
		}
	}
}

func TestNewCluster(t *testing.T) {
	intpf := func(i int) *int { return &i }
	cases := []struct {
		description string
		v1a1cl      v1alpha1.Cluster
		cl          Cluster
	}{
		{
			description: "Cluster with discovered info and without scans",
			v1a1cl: v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test_ns",
				},
				Status: v1alpha1.ClusterStatus{
					ClusterInfo: discovery.ClusterInfo{
						Provider:   "test_provider",
						Region:     "test_region",
						TotalNodes: intpf(2),
						CreationTimestamp: metav1.NewTime(
							func(s string) time.Time {
								t, err := time.Parse(time.RFC3339, "2022-07-27T17:05:38Z")
								if err != nil {
									panic(err)
								}
								return t
							}("2022-07-27T17:05:38Z"),
						),
					},
					KubernetesVersion: "v1.19.2",
					Resources: discovery.ClusterResources{
						corev1.ResourceCPU: discovery.Resources{
							Available:       resource.MustParse("3860m"),
							Usage:           resource.MustParse("285605379n"),
							UsagePercentage: 7,
						},
						corev1.ResourceMemory: discovery.Resources{
							Available:       resource.MustParse("6843704Ki"),
							Usage:           resource.MustParse("2514116Ki"),
							UsagePercentage: 36,
						},
					},
				},
			},
			cl: Cluster{
				Name:       "test",
				Namespace:  "test_ns",
				Provider:   "test_provider",
				Region:     "test_region",
				TotalNodes: intpf(2),
				Version:    "v1.19.2",
				CreationTimestamp: metav1.Time{
					Time: func(s string) time.Time {
						t, err := time.Parse(time.RFC3339, "2022-07-27T17:05:38Z")
						if err != nil {
							panic(err)
						}
						return t
					}("2022-07-27T17:05:38Z"),
				},
				Resources: &Resources{
					CPU: &Resource{
						Available:       "3860m",
						Usage:           "286m",
						UsagePercentage: 7,
					},
					Memory: &Resource{
						Available:       "6683Mi",
						Usage:           "2455Mi",
						UsagePercentage: 36,
					},
				},
				Scan: ScanStatus{Status: Unknown},
			},
		},

		{
			description: "Cluster with discovered info and with scan",
			v1a1cl: v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test_ns",
				},
				Status: v1alpha1.ClusterStatus{
					Status: apis.Status{
						Conditions: []metav1.Condition{
							{
								Type:    v1alpha1.ClusterReady,
								Status:  metav1.ConditionTrue,
								Reason:  v1alpha1.ClusterConnected,
								Message: "cluster successfully connected, version 1.19",
							},
							{
								Type:    v1alpha1.ClusterDiscovered,
								Status:  metav1.ConditionTrue,
								Reason:  v1alpha1.ClusterInfoDiscovered,
								Message: "cluster info successfully discovered",
							},
							{
								Type:    v1alpha1.ClusterScanned,
								Status:  metav1.ConditionTrue,
								Reason:  v1alpha1.ClusterScanned,
								Message: "cluster successfully scanned",
							},
						},
					},
					ClusterInfo: discovery.ClusterInfo{
						Provider:   "test_provider",
						Region:     "test_region",
						TotalNodes: intpf(1),
						CreationTimestamp: metav1.Time{
							Time: func(s string) time.Time {
								t, _ := time.Parse(time.RFC3339, s)
								return t
							}("2022-07-27T18:21:01Z"),
						},
					},
					KubernetesVersion: "v1.23.1",
					Resources: discovery.ClusterResources{
						corev1.ResourceCPU: discovery.Resources{
							Available:       resource.MustParse("4096m"),
							Usage:           resource.MustParse("384958170n"),
							UsagePercentage: 9,
						},
						corev1.ResourceMemory: discovery.Resources{
							Available:       resource.MustParse("8024880Ki"),
							Usage:           resource.MustParse("7294501Ki"),
							UsagePercentage: 91,
						},
					},
					TotalIssues: intpf(27),
					LastSuccessfulScanTime: &metav1.Time{
						Time: func(s string) time.Time {
							t, err := time.Parse(time.RFC3339, s)
							if err != nil {
								panic(err)
							}
							return t
						}("2022-07-27T18:22:01Z"),
					},
					NextScheduleScanTime: &metav1.Time{
						Time: func(s string) time.Time {
							t, err := time.Parse(time.RFC3339, s)
							if err != nil {
								panic(err)
							}
							return t
						}("2022-07-27T18:24:40Z"),
					},
				},
			},
			cl: Cluster{
				Name:       "test",
				Namespace:  "test_ns",
				Provider:   "test_provider",
				Region:     "test_region",
				TotalNodes: intpf(1),
				// Environment:       ,
				Version: "v1.23.1",
				CreationTimestamp: metav1.NewTime(
					func(s string) time.Time {
						t, _ := time.Parse(time.RFC3339, s)
						return t
					}("2022-07-27T18:21:01Z"),
				),
				Resources: &Resources{
					CPU: &Resource{
						Available:       "4096m",
						Usage:           "385m",
						UsagePercentage: 9,
					},
					Memory: &Resource{
						Available:       "7836Mi",
						Usage:           "7123Mi",
						UsagePercentage: 91,
					},
				},
				Scan:        ScanStatus{Status: Scanned},
				Connection:  ConnectionStatus{Connected: true},
				TotalIssues: intpf(27),
				LastSuccessfulScanTime: metav1.Time{
					Time: func(s string) time.Time {
						t, err := time.Parse(time.RFC3339, s)
						if err != nil {
							panic(err)
						}
						return t
					}("2022-07-27T18:22:01Z"),
				},
				NextScheduleScanTime: metav1.Time{
					Time: func(s string) time.Time {
						t, err := time.Parse(time.RFC3339, s)
						if err != nil {
							panic(err)
						}
						return t
					}("2022-07-27T18:24:40Z"),
				},
			},
		},
	}

	for _, c := range cases {
		if cl := NewCluster(c.v1a1cl); !cmp.Equal(cl, c.cl) {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Mismatch between expected and obtained results:\n%s", cmp.Diff(c.cl, cl))
		}
	}
}
