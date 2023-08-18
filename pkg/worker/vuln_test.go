package worker

import (
	"context"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

func TestParseVulnResults(t *testing.T) {
	type args struct {
		cfg      *config
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    []v1alpha1.VulnerabilityReport
		wantErr bool
	}{
		{
			name:    "invalid plugin",
			args:    args{cfg: &config{PluginName: "marvin"}}, // marvin is not a vulnerability plugin
			want:    nil,
			wantErr: true,
		},
		{
			name: "directory reader",
			args: args{
				cfg:      &config{PluginName: "trivy"},
				filename: t.TempDir(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ok",
			args: args{
				cfg: &config{
					PluginName:  "trivy",
					ClusterName: "cluster",
					Namespace:   "ns",
					JobName:     "cluster-trivy-28140229",
					JobUID:      "50c8957e-c9e1-493a-9fa4-d0786deea017",
					PodName:     "cluster-trivy-28140229-h9kcn",
					suffix:      "h9kcn",
				},
				filename: "report/trivy/testdata/report.json",
			},
			want: []v1alpha1.VulnerabilityReport{
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-registry-k8s-io-kube-apiserver-v1-25-3-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-trivy-28140229",
								UID:        types.UID("50c8957e-c9e1-493a-9fa4-d0786deea017"),
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:  "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster: "cluster",
							v1alpha1.LabelPlugin:  "trivy",
						},
					},
					Spec: v1alpha1.VulnerabilityReportSpec{
						Cluster: "cluster",
						Image:   "registry.k8s.io/kube-apiserver:v1.25.3",
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
					}},
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-quay-io-kiwigrid-k8s-sidecar-1-22-0-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-trivy-28140229",
								UID:        types.UID("50c8957e-c9e1-493a-9fa4-d0786deea017"),
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:  "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster: "cluster",
							v1alpha1.LabelPlugin:  "trivy",
						},
					},
					Spec: v1alpha1.VulnerabilityReportSpec{
						Cluster: "cluster",
						Image:   "quay.io/kiwigrid/k8s-sidecar:1.22.0",
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
					}},
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-docker-io-istio-examples-bookinfo-ratings-v1-1-17-0-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-trivy-28140229",
								UID:        types.UID("50c8957e-c9e1-493a-9fa4-d0786deea017"),
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:  "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster: "cluster",
							v1alpha1.LabelPlugin:  "trivy",
						},
					},
					Spec: v1alpha1.VulnerabilityReportSpec{
						Cluster: "cluster",
						Image:   "docker.io/istio/examples-bookinfo-ratings-v1:1.17.0",
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
					}},
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cluster-docker-io-istio-examples-bookinfo-details-v1-1-17-0-h9kcn",
						Namespace: "ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: batchv1.SchemeGroupVersion.String(),
								Kind:       "Job",
								Name:       "cluster-trivy-28140229",
								UID:        types.UID("50c8957e-c9e1-493a-9fa4-d0786deea017"),
							},
						},
						Labels: map[string]string{
							v1alpha1.LabelScanID:  "50c8957e-c9e1-493a-9fa4-d0786deea017",
							v1alpha1.LabelCluster: "cluster",
							v1alpha1.LabelPlugin:  "trivy",
						},
					},
					Spec: v1alpha1.VulnerabilityReportSpec{
						Cluster: "cluster",
						Image:   "docker.io/istio/examples-bookinfo-details-v1:1.17.0",
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
					}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r io.Reader
			if tt.args.filename != "" {
				f, err := os.Open(tt.args.filename)
				if err != nil {
					t.Errorf("parseVulnResults() setup error = %v", err)
					return
				}
				r = f
			}
			got, err := parseVulnResults(context.TODO(), tt.args.cfg, r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVulnResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sortVulns(got)
			sortVulns(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseVulnResults() mismatch (-want +got):\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func sortVulns(vulns []v1alpha1.VulnerabilityReport) {
	sort.Slice(vulns, func(i, j int) bool {
		return strings.Compare(vulns[i].Spec.Image, vulns[j].Spec.Image) == -1
	})
	for _, v := range vulns {
		for _, r := range v.Spec.Resources {
			sort.Strings(r)
		}
		sort.Slice(v.Spec.Vulnerabilities, func(i, j int) bool {
			return strings.Compare(v.Spec.Vulnerabilities[i].ID, v.Spec.Vulnerabilities[j].ID) == -1
		})
	}
}
