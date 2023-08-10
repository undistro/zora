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
	"reflect"
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    *config
		wantErr bool
	}{
		{
			name:    "empty",
			env:     nil,
			wantErr: true,
		},
		{
			name: "required only",
			env: map[string]string{
				"PLUGIN_NAME":  "plugin",
				"CLUSTER_NAME": "cluster",
				"NAMESPACE":    "ns",
				"JOB_NAME":     "cluster-plugin-28140229",
				"JOB_UID":      "50c8957e-c9e1-493a-9fa4-d0786deea017",
				"POD_NAME":     "cluster-plugin-28140229-h9kcn",
			},
			want: &config{
				DoneFile:     "/tmp/zora/results/done",
				ErrorFile:    "/tmp/zora/results/error",
				PluginName:   "plugin",
				ClusterName:  "cluster",
				Namespace:    "ns",
				JobName:      "cluster-plugin-28140229",
				JobUID:       "50c8957e-c9e1-493a-9fa4-d0786deea017",
				PodName:      "cluster-plugin-28140229-h9kcn",
				WaitInterval: time.Second,
				suffix:       "h9kcn",
			},
		},
		{
			name: "one required env missing",
			env: map[string]string{
				//"PLUGIN_NAME":  "plugin",
				"CLUSTER_NAME": "cluster",
				"NAMESPACE":    "ns",
				"JOB_NAME":     "cluster-plugin-28140229",
				"JOB_UID":      "50c8957e-c9e1-493a-9fa4-d0786deea017",
				"POD_NAME":     "cluster-plugin-28140229-h9kcn",
			},
			wantErr: true,
		},
		{
			name: "all",
			env: map[string]string{
				"PLUGIN_NAME":   "plugin",
				"CLUSTER_NAME":  "cluster",
				"NAMESPACE":     "ns",
				"JOB_NAME":      "cluster-plugin-28140229",
				"JOB_UID":       "50c8957e-c9e1-493a-9fa4-d0786deea017",
				"POD_NAME":      "cluster-plugin-28140229-h9kcn",
				"DONE_FILE":     "/done",
				"ERROR_FILE":    "/error",
				"WAIT_INTERVAL": "5s",
			},
			want: &config{
				DoneFile:     "/done",
				ErrorFile:    "/error",
				PluginName:   "plugin",
				ClusterName:  "cluster",
				Namespace:    "ns",
				JobName:      "cluster-plugin-28140229",
				JobUID:       "50c8957e-c9e1-493a-9fa4-d0786deea017",
				PodName:      "cluster-plugin-28140229-h9kcn",
				WaitInterval: 5 * time.Second,
				suffix:       "h9kcn",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			got, err := configFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("configFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("configFromEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}
