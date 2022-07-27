package payloads

import (
	"testing"

	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/google/go-cmp/cmp"
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
			t.Errorf("Mismatch between expected and obtained results: %s", cmp.Diff(c.cl, cl))
		}
	}
}
