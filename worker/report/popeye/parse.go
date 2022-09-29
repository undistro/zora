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
	zorav1a1 "github.com/undistro/zora/apis/zora/v1alpha1"
)

var msgre = regexp.MustCompile(`^\[(POP-\d+)\]\s*(.*)$`)

// Extracts Popeye's issue code and description from the original issue
// message, ensuring the returned description doesn't contain specific data
// related to cluster resources.
func prepareIdAndMsg(msg string) (string, string, error) {
	s := msgre.FindStringSubmatch(msg)
	if len(s) != 3 {
		return "", "", errors.New("Unable to split Popeye error code from message.")
	}
	if msg, ok := IssueIDtoGenericMsg[s[1][strings.LastIndex(s[1], "-")+1:]]; ok {
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
			for _, iss := range issues {
				id, msg, err := prepareIdAndMsg(iss.Message)
				if err != nil {
					return nil, fmt.Errorf("Unable to parse Popeye issue on <%s>: %w", typ, err)
				}
				if ci, ok := issuesmap[id]; ok {
					ci.Resources[san.GVR] = append(ci.Resources[san.GVR], typ)
					ci.TotalResources++
				} else {
					issuesmap[id] = &zorav1a1.ClusterIssueSpec{
						ID:       id,
						Message:  msg,
						Severity: LevelToIssueSeverity[iss.Level],
						Category: san.Sanitizer,
						Resources: map[string][]string{
							san.GVR: {typ},
						},
						TotalResources: 1,
						Url:            IssueIDtoUrl[id[strings.LastIndex(id, "-")+1:]],
					}
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
