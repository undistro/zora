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
	"encoding/json"

	"github.com/go-logr/logr"
	marvin "github.com/undistro/marvin/pkg/types"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

var marvinToZoraSeverity = map[marvin.Severity]v1alpha1.ClusterIssueSeverity{
	marvin.SeverityLow:      v1alpha1.SeverityLow,
	marvin.SeverityMedium:   v1alpha1.SeverityMedium,
	marvin.SeverityHigh:     v1alpha1.SeverityHigh,
	marvin.SeverityCritical: v1alpha1.SeverityHigh,
}

func Parse(log logr.Logger, bs []byte) ([]*v1alpha1.ClusterIssueSpec, error) {
	report := &marvin.Report{}
	if err := json.Unmarshal(bs, report); err != nil {
		return nil, err
	}
	var css []*v1alpha1.ClusterIssueSpec
	for _, check := range report.Checks {
		if check.Status != marvin.StatusFailed {
			continue
		}
		if len(check.Errors) > 0 {
			log.Info("Marvin check with errors", "check", check.ID, "errors", check.Errors)
		}
		cs := clusterIssueSpec(report, check)
		css = append(css, cs)
	}
	return css, nil
}

func clusterIssueSpec(report *marvin.Report, check *marvin.CheckResult) *v1alpha1.ClusterIssueSpec {
	resources := map[string][]string{}
	for gvk, objs := range check.Failed {
		for _, obj := range objs {
			gvr := report.GVRs[gvk]
			resources[gvr] = append(resources[gvr], obj)
		}
	}
	custom := !check.Builtin
	category := "Security"
	if c, ok := check.Labels["category"]; ok && custom {
		category = c
	}
	url := urls[check.ID]
	if u, ok := check.Labels["url"]; ok && custom {
		url = u
	}
	return &v1alpha1.ClusterIssueSpec{
		ID:             check.ID,
		Message:        check.Message,
		Severity:       marvinToZoraSeverity[check.Severity],
		Category:       category,
		Resources:      resources,
		TotalResources: 0,
		Url:            url,
		Custom:         custom,
	}
}
