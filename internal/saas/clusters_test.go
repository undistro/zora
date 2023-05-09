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
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
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
