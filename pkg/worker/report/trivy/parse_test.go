package trivy

import (
	"context"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		testfile string
		want     []v1alpha1.VulnerabilityReportSpec
		wantErr  bool
	}{
		{
			name:     "ok",
			testfile: "testdata/report.json",
			wantErr:  false,
			want: []v1alpha1.VulnerabilityReportSpec{
				{
					Image: "registry.k8s.io/kube-apiserver:v1.25.3",
					Resources: map[string][]string{
						"Pod": {"kube-system/kube-apiserver-kind-control-plane"},
					},
					Vulnerabilities: []v1alpha1.Vulnerability{
						{
							ID:          "CVE-2022-41723",
							Severity:    "HIGH",
							Description: "avoid quadratic complexity in HPACK decoding",
							Package:     "golang.org/x/net",
							Version:     "v0.0.0-20220722155237-a158d28d115b",
							FixVersion:  "0.7.0",
							URL:         "https://avd.aquasec.com/nvd/cve-2022-41723",
							Status:      "fixed",
						},
					},
				},
				{
					Image: "quay.io/kiwigrid/k8s-sidecar:1.22.0",
					Resources: map[string][]string{
						"Deployment": {
							"apps/app1",
							"apps/app2",
						},
					},
					Vulnerabilities: []v1alpha1.Vulnerability{
						{
							ID:          "CVE-2022-4450",
							Severity:    "HIGH",
							Description: "double free after calling PEM_read_bio_ex",
							Package:     "libssl1.1",
							Version:     "1.1.1s-r0",
							FixVersion:  "1.1.1t-r0",
							URL:         "https://avd.aquasec.com/nvd/cve-2022-4450",
							Status:      "fixed",
						},
						{
							ID:          "CVE-2022-4450",
							Severity:    "HIGH",
							Description: "double free after calling PEM_read_bio_ex",
							Package:     "libcrypto1.1",
							Version:     "1.1.1s-r0",
							FixVersion:  "1.1.1t-r0",
							URL:         "https://avd.aquasec.com/nvd/cve-2022-4450",
							Status:      "fixed",
						},
						{
							ID:          "CVE-2023-37920",
							Severity:    "CRITICAL",
							Description: "Removal of e-Tugra root certificate",
							Package:     "certifi",
							Version:     "2022.12.7",
							FixVersion:  "2023.7.22",
							URL:         "https://avd.aquasec.com/nvd/cve-2023-37920",
							Status:      "fixed",
						},
					},
				},
				{
					Image: "docker.io/istio/examples-bookinfo-ratings-v1:1.17.0",
					Resources: map[string][]string{
						"Deployment": {"apps/app1"},
					},
					Vulnerabilities: []v1alpha1.Vulnerability{
						{
							ID:          "DLA-3051-1",
							Severity:    "UNKNOWN",
							Description: "tzdata - new timezone database",
							Package:     "tzdata",
							Version:     "2019c-0+deb9u1",
							FixVersion:  "2021a-0+deb9u4",
							URL:         "",
							Status:      "fixed",
						},
						{
							ID:          "CVE-2016-2779",
							Severity:    "HIGH",
							Description: "util-linux: runuser tty hijack via TIOCSTI ioctl",
							Package:     "bsdutils",
							Version:     "1:2.29.2-1+deb9u1",
							FixVersion:  "",
							URL:         "https://avd.aquasec.com/nvd/cve-2016-2779",
							Status:      "affected",
						},
						{
							ID:          "GHSA-jmqm-f2gx-4fjv",
							Severity:    "MEDIUM",
							Description: "Sensitive information exposure through logs in npm-registry-fetch",
							Package:     "npm-registry-fetch",
							Version:     "4.0.4",
							FixVersion:  "8.1.1, 4.0.5",
							URL:         "https://github.com/advisories/GHSA-jmqm-f2gx-4fjv",
							Status:      "fixed",
						},
					},
				},
				{
					Image: "docker.io/istio/examples-bookinfo-details-v1:1.17.0",
					Resources: map[string][]string{
						"Deployment": {"apps/app2"},
					},
					Vulnerabilities: []v1alpha1.Vulnerability{
						{
							ID:          "CVE-2016-2781",
							Severity:    "LOW",
							Description: "coreutils: Non-privileged session can escape to the parent session in chroot",
							Package:     "coreutils",
							Version:     "8.30-3",
							FixVersion:  "",
							URL:         "https://avd.aquasec.com/nvd/cve-2016-2781",
							Status:      "will_not_fix",
						},
						{
							ID:          "CVE-2023-28755",
							Severity:    "HIGH",
							Description: "ReDoS vulnerability in URI",
							Package:     "uri",
							Version:     "0.10.0",
							FixVersion:  "~\u003e 0.10.0.1, ~\u003e 0.10.2, ~\u003e 0.11.1, \u003e= 0.12.1",
							URL:         "https://avd.aquasec.com/nvd/cve-2023-28755",
							Status:      "fixed",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.testfile)
			if err != nil {
				t.Errorf("Parse() setup error = %v", err)
			}
			got, err := Parse(context.TODO(), f)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sortVulns(got)
			sortVulns(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func sortVulns(specs []v1alpha1.VulnerabilityReportSpec) {
	sort.Slice(specs, func(i, j int) bool {
		return strings.Compare(specs[i].Image, specs[j].Image) == -1
	})
	for _, s := range specs {
		for _, v := range s.Resources {
			sort.Strings(v)
		}
		sort.Slice(s.Vulnerabilities, func(i, j int) bool {
			return strings.Compare(s.Vulnerabilities[i].ID, s.Vulnerabilities[j].ID) == -1
		})
	}
}
