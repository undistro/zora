package trivy

import (
	"context"
	"encoding/json"
	"errors"
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
			log.Error(errors.New(f.Error), fmt.Sprintf("trivy error for %s %s/%s", f.Kind, f.Namespace, f.Name))
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
			addResource(vulnsByImage[img], f.Kind, f.Namespace, f.Name)

			k := fmt.Sprintf("%s;%s", img, result.Class)
			if _, ok := parsed[k]; ok {
				continue
			}
			parsed[k] = true

			for _, vuln := range result.Vulnerabilities {
				vulnsByImage[img].Vulnerabilities = append(vulnsByImage[img].Vulnerabilities, newVulnerability(vuln))
			}
		}
	}
	specs := make([]v1alpha1.VulnerabilityReportSpec, 0, len(vulnsByImage))
	for _, spec := range vulnsByImage {
		specs = append(specs, *spec)
	}
	return specs, nil
}

func newVulnerability(vuln trivytypes.DetectedVulnerability) v1alpha1.Vulnerability {
	return v1alpha1.Vulnerability{
		ID:          vuln.VulnerabilityID,
		Severity:    vuln.Severity,
		Description: vuln.Title,
		Package:     vuln.PkgName,
		Version:     vuln.InstalledVersion,
		FixVersion:  vuln.FixedVersion,
		URL:         vuln.PrimaryURL,
		Status:      vuln.Status.String(),
	}
}

func getImage(finding trivyreport.Resource) string {
	for _, r := range finding.Results {
		if r.Class == "os-pkgs" {
			return strings.SplitN(r.Target, " (", 2)[0]
		}
	}
	return ""
}

func addResource(in *v1alpha1.VulnerabilityReportSpec, kind, namespace, name string) {
	if in.Resources == nil {
		in.Resources = map[string][]string{}
	}
	id := name
	if namespace != "" {
		id = fmt.Sprintf("%s/%s", namespace, name)
	}
	if res, ok := in.Resources[kind]; ok {
		for _, re := range res {
			if re == id {
				return
			}
		}
	}
	in.Resources[kind] = append(in.Resources[kind], id)
}
