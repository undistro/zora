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
	"context"
	"encoding/json"
	"io"

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

func Parse(ctx context.Context, results io.Reader) ([]v1alpha1.ClusterIssueSpec, error) {
	log := logr.FromContextOrDiscard(ctx)
	report := &marvin.Report{}
	if err := json.NewDecoder(results).Decode(report); err != nil {
		return nil, err
	}
	css := make([]v1alpha1.ClusterIssueSpec, 0, len(report.Checks))
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

func clusterIssueSpec(report *marvin.Report, check *marvin.CheckResult) v1alpha1.ClusterIssueSpec {
	resources := map[string][]string{}
	for gvk, objs := range check.Failed {
		for _, obj := range objs {
			gvr := report.GVRs[gvk]
			resources[gvr] = append(resources[gvr], obj)
		}
	}
	return v1alpha1.ClusterIssueSpec{
		ID:             check.ID,
		Message:        check.Message,
		Severity:       marvinToZoraSeverity[check.Severity],
		Category:       getCategory(check),
		Resources:      resources,
		TotalResources: 0,
		Url:            getURL(check),
		Custom:         !check.Builtin,
	}
}

func getCategory(check *marvin.CheckResult) string {
	if c, ok := check.Labels["category"]; ok && !check.Builtin {
		return c
	}
	switch check.ID {
	case "M-400", "M-401":
		return "Best Practices"
	case "M-402", "M-403", "M-404", "M-405", "M-406", "M-407":
		return "Reliability"
	default:
		return "Security"
	}
}
