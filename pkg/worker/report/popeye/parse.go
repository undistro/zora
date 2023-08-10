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

package popeye

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/go-logr/logr"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

var (
	msgre              = regexp.MustCompile(`^\[(POP-\d+)\]\s*(.*)$`)
	clusterScopedTypes = map[string]bool{
		"Metrics": true,
		"Version": true,
	}
)

// Extracts Popeye's issue code and description from the original issue
// message, ensuring the returned description doesn't contain specific data
// related to cluster resources.
func prepareIdAndMsg(msg string) (string, string, error) {
	s := msgre.FindStringSubmatch(msg)
	if len(s) != 3 {
		return "", "", fmt.Errorf("unable to split Popeye error code from message %s", msg)
	}
	if msg, ok := IssueIDtoGenericMsg[s[1]]; ok {
		return s[1], msg, nil
	}
	return s[1], s[2], nil
}

// Parse transforms a Popeye report into a slice of <ClusterIssueSpec>. This
// function is called by the <report> package when a Popeye plugin is used.
func Parse(ctx context.Context, results io.Reader) ([]v1alpha1.ClusterIssueSpec, error) {
	log := logr.FromContextOrDiscard(ctx)
	report := &Report{}
	if err := json.NewDecoder(results).Decode(report); err != nil {
		return nil, err
	}
	issuesByID := map[string]*v1alpha1.ClusterIssueSpec{}
	for _, sanitizer := range report.Popeye.Sanitizers {
		for typ, issues := range sanitizer.Issues {
			if typ == "" {
				if len(issues) > 0 {
					if msg := issues[0].Message; strings.Contains(msg, "forbidden") {
						log.Info(msg)
						return nil, errors.New(msg)
					}
				}
				continue
			}
			clusterScoped := clusterScopedTypes[typ]
			for _, iss := range issues {
				id, msg, err := prepareIdAndMsg(iss.Message)
				if err != nil {
					return nil, fmt.Errorf("unable to parse Popeye issue on <%s>: %w", typ, err)
				}
				if iss.Level == OkLevel {
					log.Info("Skipping OK level issue", "id", id, "msg", msg)
					continue
				}
				if ci, ok := issuesByID[id]; ok {
					ci.AddResource(sanitizer.GVR, typ)
				} else {
					spec := &v1alpha1.ClusterIssueSpec{
						ID:             id,
						Message:        msg,
						Severity:       LevelToIssueSeverity[iss.Level],
						Category:       IssueIDtoCategory[id],
						Url:            IssueIDtoUrl[id],
						Resources:      map[string][]string{},
						TotalResources: 0,
						Custom:         false,
					}
					if !clusterScoped {
						spec.TotalResources = 1
						spec.Resources = map[string][]string{sanitizer.GVR: {typ}}
					}
					issuesByID[id] = spec
				}
			}
		}
	}

	res := make([]v1alpha1.ClusterIssueSpec, 0, len(issuesByID))
	for _, ci := range issuesByID {
		res = append(res, *ci)
	}
	return res, nil
}
