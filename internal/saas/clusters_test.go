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
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

func TestNewCluster(t *testing.T) {
	tests := []struct {
		name    string
		cluster string
		want    string
	}{
		{
			name:    "Cluster and ClusterScan OK",
			cluster: "ok.yml",
			want:    "1.json",
		},
		{
			name:    "Cluster disconnected since before and ClusterScan with all plugins failed",
			cluster: "always_disconnected.yml",
			want:    "2.json",
		},
		{
			name:    "Cluster without metrics and ClusterScan with all plugins active in the 1st scan",
			cluster: "always_without_metrics.yml",
			want:    "3.json",
		},
		{
			name:    "Cluster currently disconnected and ClusterScan with Active and Failed plugins",
			cluster: "disconnected.yml",
			want:    "4.json",
		},
		{
			name:    "Cluster currently without metrics and ClusterScan with Active and Complete plugins",
			cluster: "without_metrics.yml",
			want:    "5.json",
		},
		{
			name:    "Cluster without provider/region and ClusterScan OK",
			cluster: "without_provider.yml",
			want:    "6.json",
		},
		{
			name:    "Cluster currently without metrics and two ClusterScans",
			cluster: "without_metrics.yml",
			want:    "7.json",
		},
		{
			name:    "Cluster currently without metrics, provider, region and ClusterScan",
			cluster: "without_provider_and_metrics.yml",
			want:    "8.json",
		},
		{
			name:    "Cluster currently without metrics and two ClusterScans with issues",
			cluster: "without_metrics.yml",
			want:    "9.json",
		},
		{
			name:    "Cluster without provider/region and ClusterScan suspended",
			cluster: "without_provider.yml",
			want:    "10.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster, err := readCluster(tt.cluster)
			if err != nil {
				t.Errorf("failed to load Cluster testdata: %v", err)
			}
			payload, err := readClusterPayload(tt.want)
			if err != nil {
				t.Errorf("failed to load Cluster payload testdata: %v", err)
			}
			if got := NewCluster(cluster); !reflect.DeepEqual(got, payload) {
				t.Errorf("NewCluster() mismatch (-want +got):\n%s", cmp.Diff(payload, got))
			}
		})
	}
}

func readCluster(filename string) (v1alpha1.Cluster, error) {
	b, err := os.ReadFile("testdata/cluster/" + filename)
	var c v1alpha1.Cluster
	if err != nil {
		return c, err
	}
	if err := yaml.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, err
}

func readClusterPayload(filename string) (Cluster, error) {
	b, err := os.ReadFile("testdata/payload/" + filename)
	var cp Cluster
	if err != nil {
		return cp, err
	}
	if err := json.Unmarshal(b, &cp); err != nil {
		return cp, err
	}
	return cp, err
}

func TestNewScanStatus(t *testing.T) {
	type args struct {
		clusterScan *v1alpha1.ClusterScan
		scans       []v1alpha1.ClusterScan
	}
	tests := []struct {
		name            string
		args            args
		want            map[string]*PluginStatus
		wantTotalIssues *int
	}{
		{
			name: "current scan with 21 issues, previous with 19",
			args: args{
				clusterScan: &v1alpha1.ClusterScan{
					ObjectMeta: metav1.ObjectMeta{Name: "mycluster-misconfig"},
					Spec:       v1alpha1.ClusterScanSpec{Schedule: "17 * * * *"},
					Status: v1alpha1.ClusterScanStatus{
						TotalIssues:        pointer.Int(21),
						LastScheduleTime:   mustParseTime("2024-03-28T14:17:00Z"),
						LastFinishedTime:   mustParseTime("2024-03-28T14:17:27Z"),
						LastSuccessfulTime: mustParseTime("2024-03-28T14:17:27Z"),
						NextScheduleTime:   mustParseTime("2024-03-28T15:17:00Z"),
						LastStatus:         "Complete",
						LastFinishedStatus: "Complete",
						Plugins: map[string]*v1alpha1.PluginScanStatus{
							"marvin": {
								LastScheduleTime:     mustParseTime("2024-03-28T14:17:00Z"),
								LastFinishedTime:     mustParseTime("2024-03-28T14:17:27Z"),
								LastSuccessfulTime:   mustParseTime("2024-03-28T14:17:27Z"),
								NextScheduleTime:     mustParseTime("2024-03-28T15:17:00Z"),
								LastScanID:           "9c6706f7-4d0e-4c79-bf70-efe1f4adc722",
								LastSuccessfulScanID: "9c6706f7-4d0e-4c79-bf70-efe1f4adc722",
								LastStatus:           "Complete",
								LastFinishedStatus:   "Complete",
								TotalIssues:          pointer.Int(21),
							},
						},
					},
				},
				scans: []v1alpha1.ClusterScan{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "mycluster-misconfig"},
						Spec:       v1alpha1.ClusterScanSpec{Schedule: "17 * * * *"},
						Status: v1alpha1.ClusterScanStatus{
							TotalIssues:        pointer.Int(19),
							LastScheduleTime:   mustParseTime("2024-03-28T14:17:00Z"),
							LastFinishedTime:   nil,
							LastSuccessfulTime: mustParseTime("2024-03-28T13:17:27Z"), // referring to the previous scan (1h before)
							NextScheduleTime:   mustParseTime("2024-03-28T15:17:00Z"),
							LastStatus:         "Active",
							LastFinishedStatus: "Complete",
							Plugins: map[string]*v1alpha1.PluginScanStatus{
								"marvin": {
									LastScheduleTime:     mustParseTime("2024-03-28T14:17:00Z"),
									LastFinishedTime:     nil,
									LastSuccessfulTime:   mustParseTime("2024-03-28T13:17:27Z"), // referring to the previous scan (1h before)
									NextScheduleTime:     mustParseTime("2024-03-28T15:17:00Z"),
									LastScanID:           "9c6706f7-4d0e-4c79-bf70-efe1f4adc722",
									LastSuccessfulScanID: "3eadcc75-64c0-4498-a834-320c98a4da6c", // referring to the previous scan (1h before)
									LastStatus:           "Active",
									LastFinishedStatus:   "Complete",
									TotalIssues:          pointer.Int(19),
								},
							},
						},
					},
				},
			},
			want: map[string]*PluginStatus{"marvin": {
				Scan: &ScanStatus{
					Status:  Scanned,
					Suspend: false,
					ID:      "9c6706f7-4d0e-4c79-bf70-efe1f4adc722",
				},
				IssueCount:             pointer.Int(21),
				LastSuccessfulScanTime: mustParseTime("2024-03-28T14:17:27Z"),
				LastFinishedScanTime:   mustParseTime("2024-03-28T14:17:27Z"),
				NextScheduleScanTime:   mustParseTime("2024-03-28T15:17:00Z"),
				Schedule:               "17 * * * *",
				LastSuccessfulScanID:   "9c6706f7-4d0e-4c79-bf70-efe1f4adc722",
			}},
			wantTotalIssues: pointer.Int(21),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, totalIssues := NewScanStatus(tt.args.clusterScan, tt.args.scans)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScanStatus() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
			if pointer.IntDeref(totalIssues, 0) == pointer.IntDeref(tt.wantTotalIssues, 0) {
				t.Errorf("NewScanStatus() totalIssues = %v, want %v", totalIssues, tt.wantTotalIssues)
			}
		})
	}
}
func mustParseTime(v string) *metav1.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		panic(fmt.Sprintf("mustParseTime(%s): %s", v, err.Error()))
	}
	return &metav1.Time{Time: t}
}
