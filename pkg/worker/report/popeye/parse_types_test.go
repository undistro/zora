// Copyright 2024 Undistro Authors
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

package popeye

import "testing"

func Test_getCategory(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{
			id:   "POP-100",
			want: "Container",
		},
		{
			id:   "POP-101",
			want: "Container",
		},
		{
			id:   "POP-199",
			want: "Container",
		},
		{
			id:   "POP-200",
			want: "Pod",
		},
		{
			id:   "POP-201",
			want: "Pod",
		},
		{
			id:   "POP-299",
			want: "Pod",
		},
		{
			id:   "POP-300",
			want: "Security",
		},
		{
			id:   "POP-301",
			want: "Security",
		},
		{
			id:   "POP-399",
			want: "Security",
		},
		{
			id:   "POP-400",
			want: "General",
		},
		{
			id:   "POP-401",
			want: "General",
		},
		{
			id:   "POP-499",
			want: "General",
		},
		{
			id:   "POP-500",
			want: "Workloads",
		},
		{
			id:   "POP-501",
			want: "Workloads",
		},
		{
			id:   "POP-599",
			want: "Workloads",
		},
		{
			id:   "POP-600",
			want: "HorizontalPodAutoscaler",
		},
		{
			id:   "POP-601",
			want: "HorizontalPodAutoscaler",
		},
		{
			id:   "POP-699",
			want: "HorizontalPodAutoscaler",
		},
		{
			id:   "POP-700",
			want: "Node",
		},
		{
			id:   "POP-701",
			want: "Node",
		},
		{
			id:   "POP-799",
			want: "Node",
		},
		{
			id:   "POP-800",
			want: "Namespace",
		},
		{
			id:   "POP-801",
			want: "Namespace",
		},
		{
			id:   "POP-899",
			want: "Namespace",
		},
		{
			id:   "POP-900",
			want: "PodDisruptionBudget",
		},
		{
			id:   "POP-901",
			want: "PodDisruptionBudget",
		},
		{
			id:   "POP-999",
			want: "PodDisruptionBudget",
		},
		{
			id:   "POP-1000",
			want: "Volumes",
		},
		{
			id:   "POP-1001",
			want: "Volumes",
		},
		{
			id:   "POP-1099",
			want: "Volumes",
		},
		{
			id:   "POP-1100",
			want: "Service",
		},
		{
			id:   "POP-1101",
			want: "Service",
		},
		{
			id:   "POP-1199",
			want: "Service",
		},
		{
			id:   "POP-1120",
			want: "ReplicaSet",
		},
		{
			id:   "POP-1200",
			want: "NetworkPolicies",
		},
		{
			id:   "POP-1201",
			want: "NetworkPolicies",
		},
		{
			id:   "POP-1299",
			want: "NetworkPolicies",
		},
		{
			id:   "POP-1300",
			want: "RBAC",
		},
		{
			id:   "POP-1301",
			want: "RBAC",
		},
		{
			id:   "POP-1399",
			want: "RBAC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := getCategory(tt.id); got != tt.want {
				t.Errorf("getCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}
