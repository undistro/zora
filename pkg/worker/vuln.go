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

package worker

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/undistro/zora/api/zora/v1alpha1"
	zora "github.com/undistro/zora/pkg/clientset/versioned"
	"github.com/undistro/zora/pkg/worker/report/trivy"
)

var vulnPlugins = map[string]func(ctx context.Context, reader io.Reader) ([]v1alpha1.VulnerabilityReportSpec, error){
	"trivy": trivy.Parse,
}

var vulnReportTypeMeta = metav1.TypeMeta{
	Kind:       "VulnerabilityReport",
	APIVersion: v1alpha1.SchemeGroupVersion.String(),
}

var nonAlphanumericRegex = regexp.MustCompile(`\W+`)

func handleVulnerability(ctx context.Context, cfg *config, results io.Reader, client *zora.Clientset) error {
	log := logr.FromContextOrDiscard(ctx)
	vulns, err := parseVulnResults(ctx, cfg, results)
	if err != nil {
		return err
	}
	for _, vuln := range vulns {
		v, err := client.ZoraV1alpha1().VulnerabilityReports(cfg.Namespace).Create(ctx, &vuln, createOpts)
		if err != nil {
			return fmt.Errorf("failed to create VulnerabilityReport %q: %v", vuln.Name, err)
		}
		log.Info(fmt.Sprintf("VulnerabilityReport %q successfully created", v.Name), "resourceVersion", v.ResourceVersion)
	}
	return nil
}

func parseVulnResults(ctx context.Context, cfg *config, results io.Reader) ([]v1alpha1.VulnerabilityReport, error) {
	parseFunc, ok := vulnPlugins[cfg.PluginName]
	if !ok {
		return nil, fmt.Errorf("invalid plugin %q", cfg.PluginName)
	}
	specs, err := parseFunc(ctx, results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q results: %v", cfg.PluginName, err)
	}
	owner := ownerReference(cfg)
	vulns := make([]v1alpha1.VulnerabilityReport, 0, len(specs))
	for _, spec := range specs {
		vulns = append(vulns, newVulnReport(cfg, spec, owner))
	}
	return vulns, nil
}

func newVulnReport(cfg *config, spec v1alpha1.VulnerabilityReportSpec, owner metav1.OwnerReference) v1alpha1.VulnerabilityReport {
	spec.Cluster = cfg.ClusterName
	return v1alpha1.VulnerabilityReport{
		TypeMeta: vulnReportTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s-%s", cfg.ClusterName, strings.ToLower(cleanString(spec.Image)), cfg.suffix),
			Namespace:       cfg.Namespace,
			OwnerReferences: []metav1.OwnerReference{owner},
			Labels: map[string]string{
				v1alpha1.LabelScanID:  cfg.JobUID,
				v1alpha1.LabelCluster: cfg.ClusterName,
				v1alpha1.LabelPlugin:  cfg.PluginName,
			},
		},
		Spec: spec,
	}
}

func cleanString(s string) string {
	return nonAlphanumericRegex.ReplaceAllString(s, "")
}
