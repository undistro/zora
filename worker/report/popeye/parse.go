package popeye

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	inspectv1a1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
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
	if msg, ok := IssueIDtoGenericMsg[s[1]]; ok {
		return s[1], msg, nil
	}
	return s[1], s[2], nil
}

// Parse transforms a Popeye report into a slice of <ClusterIssueSpec>. This
// function is called by the <report> package when a Popeye plugin is used.
func Parse(popr []byte) ([]*inspectv1a1.ClusterIssueSpec, error) {
	r := &Report{}
	if err := json.Unmarshal(popr, r); err != nil {
		return nil, err
	}
	issuesmap := map[string]*inspectv1a1.ClusterIssueSpec{}
	for _, san := range r.Popeye.Sanitizers {
		for typ, issues := range san.Issues {
			for _, iss := range issues {
				id, msg, err := prepareIdAndMsg(iss.Message)
				if err != nil {
					return nil, fmt.Errorf("Unable to parse Popeye issue on <%s>: %w", typ, err)
				}
				if ci, ok := issuesmap[id]; ok {
					ci.Resources[iss.GVR] = append(ci.Resources[iss.GVR], typ)
					ci.TotalResources++
				} else {
					issuesmap[id] = &inspectv1a1.ClusterIssueSpec{
						ID:       id,
						Message:  msg,
						Severity: LevelToIssueSeverity[iss.Level],
						Category: san.Sanitizer,
						Resources: map[string][]string{
							iss.GVR: {typ},
						},
						TotalResources: 1,
					}
				}
			}
		}
	}

	res := []*inspectv1a1.ClusterIssueSpec{}
	for _, ci := range issuesmap {
		res = append(res, ci)
	}
	return res, nil
}
