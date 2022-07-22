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
func ExtractSeverity(cid string, r *PostureReport) zorav1a1.ClusterIssueSeverity {
	for k, c := range r.SummaryDetails.Controls {
		if k == cid {
			return ScoreFactorSeverity(c.ScoreFactor)
		}
	}
	return zorav1a1.SeverityUnknown
}

// ExtractGvrAndInstanceName returns the GVR and the instance name from a
// Kubescape <object> record. The record may lack some of the GVR fields, in
// such a case, it'll return only the ones present. For <object> records which
// have the <relatedObjects> field populated, data from the first element of
// the later will be returned instead.
//
// This function uses the lowercased instance kind as k8s resource, given that
// Kubescape's <object> record doesn't store the resource type of the scanned
// components.
func ExtractGvrAndInstanceName(log logr.Logger, obj map[string]interface{}) (string, string, error) {
	if robj, ok := obj["relatedObjects"].([]interface{}); ok && len(robj) != 0 {
		if robj0, ok := robj[0].(map[string]interface{}); ok {
			obj = robj0
		} else {
			return "", "", errors.New("Unknown type of 1st Kubescape resource's <object.relatedObject>")
		}
	}

	gvr := []string{}
	for _, f := range [...]string{"apiGroup", "apiVersion", "kind"} {
		if v, ok := obj[f]; ok {
			vstr, ok := v.(string)
			if !ok {
				return "", "", fmt.Errorf("Unknown type of <%s> from Kubescape resource's <object>", f)
			}
			if f == "kind" {
				vstr = strings.ToLower(vstr)
			}
			gvr = append(gvr, vstr)
		}
	}
	if len(gvr) == 0 {
		return "", "", errors.New("No GVK information within Kubescape resource's <object>")
	}

	name := ""
	if v, ok := obj["name"]; ok {
		vstr, ok := v.(string)
		if !ok {
			log.Error(errors.New("Unknown field type"), "Unknown type of <name> from Kubescape resource's <object>")
		}
		name = vstr
	} else if m, ok := obj["metadata"]; ok {
		mmap, ok := m.(map[string]interface{})
		if !ok {
			log.Error(errors.New("Unknown field type"), "Unknown type of <metadata> from Kubescape resource's <object>")
		}
		if n, ok := mmap["name"]; ok {
			nstr, ok := n.(string)
			if !ok {
				log.Error(errors.New("Unknown field type"), "Unknown type of <name> from Kubescape resource's <object.metadata>")
			}
			name = nstr
		}
	}
	return strings.Join(gvr, "/"), name, nil
}

// ExtractStatus derives the scan status of a given Kubescape Control. The
// status Error, Unknown, Irrelevant and Failed have a higher priority over the
// others, given that these signal some caveat in the scan. In case no higher
// priority status is present, the most frequent is returned.
//
// The high priority status follow the hierarchy:
// 		Failed > Error > Unknown > Irrelevant
func ExtractStatus(con *ResourceAssociatedControl) ScanningStatus {
	stc := map[ScanningStatus]int{}
	for _, r := range con.ResourceAssociatedRules {
		stc[r.Status]++
	}

	for _, s := range [...]ScanningStatus{StatusFailed, StatusError, StatusUnknown, StatusIrrelevant} {
		if c, ok := stc[s]; ok && c > 0 {
			return s
		}
	}
	bigc := -1
	bigs := StatusUnknown
	for s, c := range stc {
		if c > bigc {
			bigc = c
			bigs = s
		}
	}
	return bigs
}

func PreprocessResources(r *PostureReport) (map[string]map[string]interface{}, error) {
	objmap := map[string]map[string]interface{}{}
	for _, res := range r.Resources {
		obj, ok := res.Object.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Unknown type of Kubescape resource's <object> with <resourceID>: <%s>", res.ResourceID)
		}
		objmap[res.ResourceID] = obj
	}
	return objmap, nil
}

// Parse transforms a Kubescape report into a slice of <ClusterIssueSpec>. This
// function is called by the <report> package when a Kubescape plugin is used.
func Parse(log logr.Logger, fcont []byte) ([]*zorav1a1.ClusterIssueSpec, error) {
	r := &PostureReport{}
	if err := json.Unmarshal(fcont, r); err != nil {
		return nil, err
	}
	issuesmap := map[string]*zorav1a1.ClusterIssueSpec{}
	objmap, err := PreprocessResources(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to preprocess Kubescape resources: %w", err)
	}
	for _, res := range r.Results {
		gvr, rname, err := ExtractGvrAndInstanceName(log, objmap[res.ResourceID])
		if err != nil {
			return nil, fmt.Errorf("Failed to extract GVR: %w", err)
		}

		for _, c := range res.AssociatedControls {
			st := ExtractStatus(&c)
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
						Severity: ExtractSeverity(c.ControlID, r),
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
