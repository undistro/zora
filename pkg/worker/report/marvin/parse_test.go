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

package marvin

import (
	"os"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"

	"github.com/undistro/zora/apis/zora/v1alpha1"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     []*v1alpha1.ClusterIssueSpec
		wantErr  bool
	}{
		{
			name:     "OK",
			filename: "httpbin.json",
			want: []*v1alpha1.ClusterIssueSpec{
				{
					ID:       "M-116",
					Message:  "Not allowed added/dropped capabilities",
					Severity: v1alpha1.SeverityLow,
					Category: "Security",
					Resources: map[string][]string{
						"apps/v1/deployments": {"httpbin/httpbin"},
						"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
					},
					Url: pssRestrictedURL,
				},
				{
					ID:       "M-113",
					Message:  "Container could be running as root user",
					Severity: v1alpha1.SeverityMedium,
					Category: "Security",
					Resources: map[string][]string{
						"apps/v1/deployments": {"httpbin/httpbin"},
						"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
					},
					Url: pssRestrictedURL,
				},
				{
					ID:       "M-115",
					Message:  "Not allowed seccomp profile",
					Severity: v1alpha1.SeverityLow,
					Category: "Security",
					Resources: map[string][]string{
						"apps/v1/deployments": {"httpbin/httpbin"},
						"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
					},
					Url: pssRestrictedURL,
				},
				{
					ID:       "M-202",
					Message:  "Automounted service account token",
					Severity: v1alpha1.SeverityLow,
					Category: "Security",
					Resources: map[string][]string{
						"apps/v1/deployments": {"httpbin/httpbin"},
						"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
					},
					Url: "https://microsoft.github.io/Threat-Matrix-for-Kubernetes/mitigations/MS-M9025%20Disable%20Service%20Account%20Auto%20Mount/",
				},
				{
					ID:       "M-300",
					Message:  "Root filesystem write allowed",
					Severity: v1alpha1.SeverityLow,
					Category: "Security",
					Resources: map[string][]string{
						"apps/v1/deployments": {"httpbin/httpbin"},
						"apps/v1/replicasets": {"httpbin/httpbin-5978c9d878"},
					},
					Url: "https://media.defense.gov/2022/Aug/29/2003066362/-1/-1/0/CTR_KUBERNETES_HARDENING_GUIDANCE_1.2_20220829.PDF#page=50",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := os.ReadFile("testdata/" + tt.filename)
			if err != nil {
				t.Errorf("Read testdata file error = %v", err)
			}
			got, err := Parse(logr.Discard(), bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %s", cmp.Diff(got, tt.want))
			}
		})
	}
}
