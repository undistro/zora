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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"

	zorav1a1 "github.com/undistro/zora/api/zora/v1alpha1"
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
func Parse(log logr.Logger, popr []byte) ([]*zorav1a1.ClusterIssueSpec, error) {
	r := &Report{}
	if err := json.Unmarshal(popr, r); err != nil {
		return nil, err
	}
	issuesmap := map[string]*zorav1a1.ClusterIssueSpec{}
	for _, san := range r.Popeye.Sanitizers {
		for typ, issues := range san.Issues {
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
				if ci, ok := issuesmap[id]; ok {
					ci.AddResource(san.GVR, typ)
				} else {
					spec := &zorav1a1.ClusterIssueSpec{
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
						spec.Resources = map[string][]string{san.GVR: {typ}}
					}
					issuesmap[id] = spec
				}
			}
		}
	}

	res := []*zorav1a1.ClusterIssueSpec{}
	for _, ci := range issuesmap {
		res = append(res, ci)
	}
	return res, nil
}
