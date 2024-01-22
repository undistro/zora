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
	"os"
	"strconv"
	"strings"
	"time"

	trivyreport "github.com/aquasecurity/trivy/pkg/k8s/report"
	trivytypes "github.com/aquasecurity/trivy/pkg/types"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

func Parse(ctx context.Context, results io.Reader) ([]v1alpha1.VulnerabilityReportSpec, error) {
	log := logr.FromContextOrDiscard(ctx)
	report := &trivyreport.Report{}
	if err := json.NewDecoder(results).Decode(report); err != nil {
		return nil, err
	}
	ignoreDescriptions, _ := strconv.ParseBool(os.Getenv("TRIVY_IGNORE_VULN_DESCRIPTIONS"))
	vulnsByImage := make(map[string]*v1alpha1.VulnerabilityReportSpec)

	// map to control which image + class was parsed
	parsed := make(map[string]bool)

	for _, r := range report.Resources {
		if r.Kind == "" {
			continue
		}
		if len(r.Error) > 0 {
			log.Info(fmt.Sprintf("trivy error for %q \"%s/%s\": %s", r.Kind, r.Namespace, r.Name, r.Error))
			continue
		}
		img := getImage(r)
		if img == "" {
			log.Info(`skipping finding without "os-pkgs" result`)
			continue
		}
		for _, result := range r.Results {
			if len(result.Vulnerabilities) == 0 {
				log.Info("skipping result without vulnerabilities")
				continue
			}
			if _, ok := vulnsByImage[img]; !ok {
				vulnsByImage[img] = newSpec(img, r)
			}
			spec := vulnsByImage[img]
			addResource(spec, r.Kind, r.Namespace, r.Name)

			k := fmt.Sprintf("%s;%s", img, result.Class)
			if _, ok := parsed[k]; ok {
				continue
			}
			parsed[k] = true

			for _, vuln := range result.Vulnerabilities {
				spec.Vulnerabilities = append(spec.Vulnerabilities, newVulnerability(vuln, ignoreDescriptions, string(result.Type)))
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

func newSpec(img string, resource trivyreport.Resource) *v1alpha1.VulnerabilityReportSpec {
	meta := resource.Metadata
	s := &v1alpha1.VulnerabilityReportSpec{
		Image:        img,
		Tags:         meta.RepoTags,
		Architecture: meta.ImageConfig.Architecture,
		OS:           meta.ImageConfig.OS,
	}
	if len(meta.RepoDigests) > 0 {
		s.Digest = meta.RepoDigests[0]
	}
	if o := meta.OS; o != nil {
		s.Distro = &v1alpha1.Distro{
			Name:    string(o.Family),
			Version: o.Name,
		}
	}
	return s
}

func newVulnerability(vuln trivytypes.DetectedVulnerability, ignoreDescription bool, t string) v1alpha1.Vulnerability {
	description := ""
	if !ignoreDescription {
		description = vuln.Description
	}

	return v1alpha1.Vulnerability{
		ID:               vuln.VulnerabilityID,
		Severity:         vuln.Severity,
		Title:            vuln.Title,
		Description:      description,
		Package:          vuln.PkgName,
		Version:          vuln.InstalledVersion,
		FixVersion:       vuln.FixedVersion,
		URL:              vuln.PrimaryURL,
		Status:           vuln.Status.String(),
		Score:            getScore(vuln),
		Type:             t,
		PublishedDate:    parseTime(vuln.PublishedDate),
		LastModifiedDate: parseTime(vuln.LastModifiedDate),
	}
}

func parseTime(t *time.Time) *metav1.Time {
	if t == nil {
		return nil
	}
	return &metav1.Time{Time: *t}
}

func getScore(vuln trivytypes.DetectedVulnerability) string {
	var vendor *float64
	for id, cvss := range vuln.CVSS {
		cvss := cvss
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

func getImage(resource trivyreport.Resource) string {
	for _, r := range resource.Results {
		if r.Class == trivytypes.ClassOSPkg {
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
