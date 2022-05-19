package popeye

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	inspectv1a1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
)

// Extracts Popeye's error code and message from the original issue message.
func splitCodeAndMsg(msg string) (string, string, error) {
	msgre := regexp.MustCompile(`^\[(POP-\d+)\]\s*(.*)$`)
	s := msgre.FindStringSubmatch(msg)
	if len(s) != 3 {
		return "", "", errors.New("Unable to split Popeye error code from message.")
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
				id, msg, err := splitCodeAndMsg(iss.Message)
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
