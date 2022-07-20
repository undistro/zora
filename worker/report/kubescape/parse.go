package kubescape

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	zorav1a1 "github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/go-logr/logr"
)

// ScoreFactorSeverity converts a Kubescape Control <ScoreFactor> to Zora's
// <ClusterIssueSeverity>.
func ScoreFactorSeverity(s float32) zorav1a1.ClusterIssueSeverity {
	switch {
	case s >= 7:
		return zorav1a1.SeverityHigh
	case s >= 4:
		return zorav1a1.SeverityMedium
	case s >= 1:
		return zorav1a1.SeverityLow
	default:
		return zorav1a1.SeverityUnknown
	}
}

// ExtractSeverity finds the Control record which contains a given <ControlID>
// then returns its <ScoreFactor> as Zora's <ClusterIssueSeverity>, together
// with the Control's scan status. If the Control is not found, the function
// returns unknown types for both types.
func ExtractSeverityAndState(cid string, r *PostureReport) (zorav1a1.ClusterIssueSeverity, ScanningStatus) {
	for k, c := range r.SummaryDetails.Controls {
		if k == cid {
			return ScoreFactorSeverity(c.ScoreFactor), c.Status
		}
	}
	return zorav1a1.SeverityUnknown, StatusUnknown
}

// ExtractGvrAndResourceName returns the GVR and the resource name from a
// Kubescape <object> record. The record may lack some of the GVR fields, in
// such a case, it'll return only the ones present.
//
// This function uses the lowercased instance kind as k8s resource, given that
// Kubescape's <object> record doesn't store the resource type of the scanned
// components.
func ExtractGvrAndResourceName(rid string, r *PostureReport) (string, string, error) {
	for _, res := range r.Resources {
		if res.ResourceID == rid {
			obj, ok := res.Object.(map[string]interface{})
			if !ok {
				return "", "", errors.New("Unknown type of Kubescape resource's <object>")
			}
			gvr := []string{}

			for _, f := range [...]string{"apiGroup", "apiVersion", "kind"} {
				if v, ok := obj[f]; ok {
					vstr, ok := v.(string)
					if !ok {
						return "", "", fmt.Errorf("Unknown type of <%s> from Kubescape resource's <object>", f)
					}
					gvr = append(gvr, strings.ToLower(vstr))
				}
			}
			if len(gvr) == 0 {
				return "", "", errors.New("No GVK information within Kubescape resource's <object>")
			}

			rname := ""
			if v, ok := obj["name"]; ok {
				vstr, ok := v.(string)
				if !ok {
					log.Error(errors.New("Unknown field type"), "Unknown type of <name> from Kubescape resource's <object>")
				}
				rname = vstr
			} else if m, ok := obj["metadata"]; ok {
				mmap, _ := m.(map[string]interface{})
				if n, ok := mmap["name"]; ok {
					nstr, ok := n.(string)
					if !ok {
						log.Error(errors.New("Unknown field type"), "Unknown type of <name> from Kubescape resource's <object.metadata>")
					}
					rname = nstr
				}
			}
			return strings.Join(gvr, "/"), rname, nil
		}
	}
	return "", "", errors.New("Unable to extract GVR")
}

// Parse transforms a Kubescape report into a slice of <ClusterIssueSpec>. This
// function is called by the <report> package when a Kubescape plugin is used.
func Parse(log logr.Logger, fcont []byte) ([]*zorav1a1.ClusterIssueSpec, error) {
	r := &PostureReport{}
	if err := json.Unmarshal(fcont, r); err != nil {
		return nil, err
	}
	issuesmap := map[string]*zorav1a1.ClusterIssueSpec{}
	for _, res := range r.Results {
		gvr, rname, err := ExtractGvrAndResourceName(res.ResourceID, r)
		if err != nil {
			return nil, fmt.Errorf("Failed to extract GVR: %w", err)
		}

		for _, c := range res.AssociatedControls {
			sev, st := ExtractSeverityAndState(c.ControlID, r)
			switch st {
			case StatusUnknown, StatusIrrelevant, StatusError:
				log.Info(fmt.Sprintf("Kubescape Control <%s> with status <%s> on instance <%s>", c.ControlID, st, rname))
				continue
			case StatusFailed:
				if ci, ok := issuesmap[c.ControlID]; ok {
					ci.Resources[gvr] = append(ci.Resources[gvr], rname)
					ci.TotalResources++
				} else {
					issuesmap[c.ControlID] = &zorav1a1.ClusterIssueSpec{
						ID:       c.ControlID,
						Message:  c.Name,
						Severity: sev,
						Category: gvr[strings.LastIndex(gvr, "/")+1:],
						Resources: map[string][]string{
							gvr: {rname},
						},
						TotalResources: 1,
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
