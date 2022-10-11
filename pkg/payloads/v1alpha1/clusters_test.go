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
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/undistro/zora/apis/zora/v1alpha1"
	"sigs.k8s.io/yaml"
)

func TestNewCluster(t *testing.T) {
	type args struct {
		cluster string
		scans   []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Cluster and ClusterScan OK",
			args: args{
				cluster: "ok.yml",
				scans:   []string{"all_plugins_active.yml"},
			},
			want: "1.json",
		},
		{
			name: "Cluster disconnected since before and ClusterScan with all plugins failed",
			args: args{
				cluster: "always_disconnected.yml",
				scans:   []string{"all_plugins_failed.yml"},
			},
			want: "2.json",
		},
		{
			name: "Cluster without metrics and ClusterScan with all plugins active in the 1st scan",
			args: args{
				cluster: "always_without_metrics.yml",
				scans:   []string{"all_plugins_active_1st.yml"},
			},
			want: "3.json",
		},
		{
			name: "Cluster currently disconnected and ClusterScan with Active and Failed plugins",
			args: args{
				cluster: "disconnected.yml",
				scans:   []string{"plugins_active_and_failed.yml"},
			},
			want: "4.json",
		},
		{
			name: "Cluster currently without metrics and ClusterScan with Active and Complete plugins",
			args: args{
				cluster: "without_metrics.yml",
				scans:   []string{"plugins_active_and_complete.yml"},
			},
			want: "5.json",
		},
		{
			name: "Cluster without provider/region and ClusterScan OK",
			args: args{
				cluster: "without_provider.yml",
				scans:   []string{"ok.yml"},
			},
			want: "6.json",
		},
		{
			name: "Cluster currently without metrics and two ClusterScans",
			args: args{
				cluster: "without_metrics.yml",
				scans:   []string{"plugins_complete_and_failed.yml", "next.yml"},
			},
			want: "7.json",
		},
		{
			name: "Cluster currently without metrics, provider, region and ClusterScan",
			args: args{
				cluster: "without_provider_and_metrics.yml",
				scans:   []string{},
			},
			want: "8.json",
		},
		{
			name: "Cluster currently without metrics and two ClusterScans with issues",
			args: args{
				cluster: "without_metrics.yml",
				scans:   []string{"ok.yml", "plugins_active_and_complete.yml"},
			},
			want: "9.json",
		},
		{
			name: "Cluster without provider/region and ClusterScan suspended",
			args: args{
				cluster: "without_provider.yml",
				scans:   []string{"suspend.yml"},
			},
			want: "10.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster, err := readCluster(tt.args.cluster)
			if err != nil {
				t.Errorf("failed to load Cluster testdata: %v", err)
			}
			scans := make([]v1alpha1.ClusterScan, 0, len(tt.args.scans))
			for _, s := range tt.args.scans {
				if cs, err := readClusterScan(s); err != nil {
					t.Errorf("failed to load ClusterScan testdata: %v", err)
				} else {
					scans = append(scans, cs)
				}
			}
			payload, err := readClusterPayload(tt.want)
			if err != nil {
				t.Errorf("failed to load Cluster payload testdata: %v", err)
			}
			if got := NewCluster(cluster, scans); !reflect.DeepEqual(got, payload) {
				t.Errorf("NewCluster() = %s", cmp.Diff(got, payload))
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

func readClusterScan(filename string) (v1alpha1.ClusterScan, error) {
	b, err := os.ReadFile("testdata/clusterscan/" + filename)
	var cs v1alpha1.ClusterScan
	if err != nil {
		return cs, err
	}
	if err := yaml.Unmarshal(b, &cs); err != nil {
		return cs, err
	}
	return cs, err
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
