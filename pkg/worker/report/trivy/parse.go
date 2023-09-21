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

package trivy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	trivyreport "github.com/aquasecurity/trivy/pkg/k8s/report"
	trivytypes "github.com/aquasecurity/trivy/pkg/types"
	"github.com/go-logr/logr"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

func Parse(ctx context.Context, results io.Reader) ([]v1alpha1.VulnerabilityReportSpec, error) {
	log := logr.FromContextOrDiscard(ctx)
	report := &trivyreport.ConsolidatedReport{}
	if err := json.NewDecoder(results).Decode(report); err != nil {
		return nil, err
	}
	vulnsByImage := make(map[string]*v1alpha1.VulnerabilityReportSpec)

	// map to control which image + class was parsed
	parsed := make(map[string]bool)

	for _, f := range report.Findings {
		if f.Kind == "" {
			continue
		}
		if len(f.Error) > 0 {
			log.Info(fmt.Sprintf("trivy error for %q \"%s/%s\": %s", f.Kind, f.Namespace, f.Name, f.Error))
			continue
		}
		img := getImage(f)
		if img == "" {
			log.Info(`skipping finding without "os-pkgs" result`)
			continue
		}
		for _, result := range f.Results {
			if len(result.Vulnerabilities) == 0 {
				continue
			}
			if _, ok := vulnsByImage[img]; !ok {
				vulnsByImage[img] = &v1alpha1.VulnerabilityReportSpec{Image: img}
			}
			spec := vulnsByImage[img]
			addResource(spec, f.Kind, f.Namespace, f.Name)

			k := fmt.Sprintf("%s;%s", img, result.Class)
			if _, ok := parsed[k]; ok {
				continue
			}
			parsed[k] = true

			for _, vuln := range result.Vulnerabilities {
				spec.Vulnerabilities = append(spec.Vulnerabilities, newVulnerability(vuln, result.Type))
			}
		}
	}
	specs := make([]v1alpha1.VulnerabilityReportSpec, 0, len(vulnsByImage))
	for _, spec := range vulnsByImage {
		summarize(spec)
		specs = append(specs, *spec)
	}
	return specs, nil
}

func newVulnerability(vuln trivytypes.DetectedVulnerability, resultType string) v1alpha1.Vulnerability {
	return v1alpha1.Vulnerability{
		ID:          vuln.VulnerabilityID,
		Severity:    vuln.Severity,
		Title:       vuln.Title,
		Description: vuln.Description,
		Package:     vuln.PkgName,
		Version:     vuln.InstalledVersion,
		FixVersion:  vuln.FixedVersion,
		URL:         vuln.PrimaryURL,
		Status:      vuln.Status.String(),
		Score:       getScore(vuln),
		Type:        resultType,
	}
}

func getScore(vuln trivytypes.DetectedVulnerability) string {
	var vendor *float64
	for id, cvss := range vuln.CVSS {
		if cvss.V3Score == 0.0 {
			continue
		}
		if string(id) == "nvd" {
			return fmt.Sprintf("%v", cvss.V3Score)
		}
		vendor = &cvss.V3Score
	}
	if vendor == nil {
		return ""
	}
	return fmt.Sprintf("%v", *vendor)
}

func getImage(finding trivyreport.Resource) string {
	for _, r := range finding.Results {
		if r.Class == "os-pkgs" {
			return strings.SplitN(r.Target, " (", 2)[0]
		}
	}
	return ""
}

func addResource(spec *v1alpha1.VulnerabilityReportSpec, kind, namespace, name string) {
	if spec.Resources == nil {
		spec.Resources = map[string][]string{}
	}
	id := name
	if namespace != "" {
		id = fmt.Sprintf("%s/%s", namespace, name)
	}
	if res, ok := spec.Resources[kind]; ok {
		for _, re := range res {
			if re == id {
				return
			}
		}
	}
	spec.Resources[kind] = append(spec.Resources[kind], id)
	spec.TotalResources++
}

func summarize(spec *v1alpha1.VulnerabilityReportSpec) {
	s := &v1alpha1.VulnerabilitySummary{}
	for _, v := range spec.Vulnerabilities {
		s.Total++
		switch v.Severity {
		case "CRITICAL":
			s.Critical++
		case "HIGH":
			s.High++
		case "MEDIUM":
			s.Medium++
		case "LOW":
			s.Low++
		default:
			s.Unknown++
		}
	}
	spec.Summary = *s
}
