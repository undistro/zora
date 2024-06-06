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
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/undistro/zora/api/zora/v1alpha1"
	"github.com/undistro/zora/api/zora/v1alpha2"
)

var labels = map[string]string{
	v1alpha1.LabelScanID:     "50c8957e-c9e1-493a-9fa4-d0786deea017",
	v1alpha1.LabelCluster:    "cluster",
	v1alpha1.LabelPlugin:     "trivy",
	v1alpha1.LabelClusterUID: "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
}

var owners = []metav1.OwnerReference{
	{
		APIVersion: batchv1.SchemeGroupVersion.String(),
		Kind:       "Job",
		Name:       "cluster-trivy-28140229",
		UID:        types.UID("50c8957e-c9e1-493a-9fa4-d0786deea017"),
	},
}

func TestParseVulnResults(t *testing.T) {
	type args struct {
		cfg      *config
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    []v1alpha2.VulnerabilityReport
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
					ClusterUID:  "9a1d324c-9170-4aa7-9f64-76f01c9d7989",
					Namespace:   "ns",
					JobName:     "cluster-trivy-28140229",
					JobUID:      "50c8957e-c9e1-493a-9fa4-d0786deea017",
					PodName:     "cluster-trivy-28140229-h9kcn",
					suffix:      "h9kcn",
				},
				filename: "report/trivy/testdata/report.json",
			},
			want: []v1alpha2.VulnerabilityReport{
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:            "cluster-registryk8siokubeapiserverv1273-h9kcn",
						Namespace:       "ns",
						OwnerReferences: owners,
						Labels:          labels,
					},
					Spec: v1alpha2.VulnerabilityReportSpec{
						VulnerabilityReportCommon: v1alpha1.VulnerabilityReportCommon{
							Cluster:        "cluster",
							Image:          "registry.k8s.io/kube-apiserver:v1.27.3",
							Tags:           []string{"registry.k8s.io/kube-apiserver:v1.27.3"},
							Digest:         "registry.k8s.io/kube-apiserver@sha256:fd03335dd2e7163e5e36e933a0c735d7fec6f42b33ddafad0bc54f333e4a23c0",
							Architecture:   "amd64",
							OS:             "linux",
							Distro:         &v1alpha1.Distro{Name: "debian", Version: "11.7"},
							Resources:      map[string][]string{"Pod": {"kube-system/kube-apiserver-kind-control-plane"}},
							TotalResources: 1,
							Summary:        v1alpha1.VulnerabilitySummary{Total: 1, High: 1},
						},
						TotalPackages:       1,
						TotalUniquePackages: 1,
						Vulnerabilities: []v1alpha2.Vulnerability{
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "CVE-2022-41723",
									Severity:         "HIGH",
									Title:            "avoid quadratic complexity in HPACK decoding",
									Description:      "A maliciously crafted HTTP/2 stream could cause excessive CPU consumption in the HPACK decoder, sufficient to cause a denial of service from a small number of small requests.",
									URL:              "https://avd.aquasec.com/nvd/cve-2022-41723",
									Score:            "7.5",
									PublishedDate:    newTime("2023-02-28T18:15:00Z"),
									LastModifiedDate: newTime("2023-05-16T10:50:00Z"),
								},
								Packages: []v1alpha1.Package{{
									Package:    "golang.org/x/net",
									Version:    "v0.0.0-20220722155237-a158d28d115b",
									FixVersion: "0.7.0",
									Status:     "fixed",
									Type:       "gobinary",
								}},
							},
						},
					},
				},
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:            "cluster-quayiokiwigridk8ssidecar1220-h9kcn",
						Namespace:       "ns",
						OwnerReferences: owners,
						Labels:          labels,
					},
					Spec: v1alpha2.VulnerabilityReportSpec{
						VulnerabilityReportCommon: v1alpha1.VulnerabilityReportCommon{
							Cluster:        "cluster",
							Image:          "quay.io/kiwigrid/k8s-sidecar:1.22.0",
							Tags:           []string{"quay.io/kiwigrid/k8s-sidecar:1.22.0"},
							Digest:         "quay.io/kiwigrid/k8s-sidecar@sha256:eaa478cdd0b8e1be7a4813bc1b01948b838e2feaa6d999e60c997dc823013824",
							Architecture:   "amd64",
							OS:             "linux",
							Distro:         &v1alpha1.Distro{Name: "alpine", Version: "3.16.3"},
							Resources:      map[string][]string{"Deployment": {"apps/app1", "apps/app2"}},
							TotalResources: 2,
							Summary:        v1alpha1.VulnerabilitySummary{Total: 3, Critical: 1, High: 2},
						},
						TotalPackages:       4,
						TotalUniquePackages: 3,
						Vulnerabilities: []v1alpha2.Vulnerability{
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "CVE-2022-4450",
									Severity:         "HIGH",
									Title:            "double free after calling PEM_read_bio_ex",
									Description:      "The function PEM_read_bio_ex() reads a PEM file from a BIO and parses and decodes the \"name\" (e.g. \"CERTIFICATE\"), any header data and the payload data. If the function succeeds then the \"name_out\", \"header\" and \"data\" arguments are populated with pointers to buffers containing the relevant decoded data. The caller is responsible for freeing those buffers. It is possible to construct a PEM file that results in 0 bytes of payload data. In this case PEM_read_bio_ex() will return a failure code but will populate the header argument with a pointer to a buffer that has already been freed. If the caller also frees this buffer then a double free will occur. This will most likely lead to a crash. This could be exploited by an attacker who has the ability to supply malicious PEM files for parsing to achieve a denial of service attack. The functions PEM_read_bio() and PEM_read() are simple wrappers around PEM_read_bio_ex() and therefore these functions are also directly affected. These functions are also called indirectly by a number of other OpenSSL functions including PEM_X509_INFO_read_bio_ex() and SSL_CTX_use_serverinfo_file() which are also vulnerable. Some OpenSSL internal uses of these functions are not vulnerable because the caller does not free the header argument if PEM_read_bio_ex() returns a failure code. These locations include the PEM_read_bio_TYPE() functions as well as the decoders introduced in OpenSSL 3.0. The OpenSSL asn1parse command line application is also impacted by this issue.",
									URL:              "https://avd.aquasec.com/nvd/cve-2022-4450",
									Score:            "7.5",
									PublishedDate:    newTime("2023-02-08T20:15:00Z"),
									LastModifiedDate: newTime("2023-07-19T00:57:00Z"),
								},
								Packages: []v1alpha1.Package{
									{
										Package:    "libssl1.1",
										Version:    "1.1.1s-r0",
										FixVersion: "1.1.1t-r0",
										Status:     "fixed",
										Type:       "alpine",
									},
									{
										Package:    "libcrypto1.1",
										Version:    "1.1.1s-r0",
										FixVersion: "1.1.1t-r0",
										Status:     "fixed",
										Type:       "alpine",
									},
								},
							},
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "CVE-2023-37920",
									Severity:         "CRITICAL",
									Title:            "Removal of e-Tugra root certificate",
									Description:      "Certifi is a curated collection of Root Certificates for validating the trustworthiness of SSL certificates while verifying the identity of TLS hosts. Certifi prior to version 2023.07.22 recognizes \"e-Tugra\" root certificates. e-Tugra's root certificates were subject to an investigation prompted by reporting of security issues in their systems. Certifi 2023.07.22 removes root certificates from \"e-Tugra\" from the root store.",
									URL:              "https://avd.aquasec.com/nvd/cve-2023-37920",
									Score:            "9.8",
									PublishedDate:    newTime("2023-07-25T21:15:00Z"),
									LastModifiedDate: newTime("2023-08-12T06:16:00Z"),
								},
								Packages: []v1alpha1.Package{{
									Package:    "certifi",
									Version:    "2022.12.7",
									FixVersion: "2023.7.22",
									Status:     "fixed",
									Type:       "python-pkg",
								}},
							},
							{
								Packages: []v1alpha1.Package{
									{
										Package:    "libssl1.1",
										Version:    "1.1.1s-r0",
										FixVersion: "1.1.1t-r0",
										Status:     "fixed",
										Type:       "alpine",
									},
								},
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "CVE-2023-0286",
									Severity:         "HIGH",
									Title:            "openssl: X.400 address type confusion in X.509 GeneralName",
									Description:      "There is a type confusion vulnerability relating to X.400 address processing\ninside an X.509 GeneralName. X.400 addresses were parsed as an ASN1_STRING but\nthe public structure definition for GENERAL_NAME incorrectly specified the type\nof the x400Address field as ASN1_TYPE. This field is subsequently interpreted by\nthe OpenSSL function GENERAL_NAME_cmp as an ASN1_TYPE rather than an\nASN1_STRING.\n\nWhen CRL checking is enabled (i.e. the application sets the\nX509_V_FLAG_CRL_CHECK flag), this vulnerability may allow an attacker to pass\narbitrary pointers to a memcmp call, enabling them to read memory contents or\nenact a denial of service. In most cases, the attack requires the attacker to\nprovide both the certificate chain and CRL, neither of which need to have a\nvalid signature. If the attacker only controls one of these inputs, the other\ninput must already contain an X.400 address as a CRL distribution point, which\nis uncommon. As such, this vulnerability is most likely to only affect\napplications which have implemented their own functionality for retrieving CRLs\nover a network.\n\n",
									URL:              "https://avd.aquasec.com/nvd/cve-2023-0286",
									Score:            "7.4",
									PublishedDate:    newTime("2023-02-08T20:15:24.267Z"),
									LastModifiedDate: newTime("2024-02-04T09:15:09.113Z"),
								},
							},
						},
					},
				},
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:            "cluster-dockerioistioexamplesbookinforatingsv11170-h9kcn",
						Namespace:       "ns",
						OwnerReferences: owners,
						Labels:          labels,
					},
					Spec: v1alpha2.VulnerabilityReportSpec{
						VulnerabilityReportCommon: v1alpha1.VulnerabilityReportCommon{
							Cluster:        "cluster",
							Image:          "docker.io/istio/examples-bookinfo-ratings-v1:1.17.0",
							Tags:           []string{"istio/examples-bookinfo-ratings-v1:1.17.0"},
							Digest:         "istio/examples-bookinfo-ratings-v1@sha256:b6a6b88d35785c19f6dcb6acf055aa585511f2126bb0b5802f3107b7d37ead0b",
							Architecture:   "amd64",
							OS:             "linux",
							Distro:         &v1alpha1.Distro{Name: "debian", Version: "9.12"},
							Resources:      map[string][]string{"Deployment": {"apps/app1"}},
							TotalResources: 1,
							Summary:        v1alpha1.VulnerabilitySummary{Total: 3, High: 1, Medium: 1, Unknown: 1},
						},
						TotalPackages:       3,
						TotalUniquePackages: 3,
						Vulnerabilities: []v1alpha2.Vulnerability{
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "DLA-3051-1",
									Severity:         "UNKNOWN",
									Title:            "tzdata - new timezone database",
									Description:      "",
									URL:              "",
									PublishedDate:    nil,
									LastModifiedDate: nil,
								},
								Packages: []v1alpha1.Package{{
									Package:    "tzdata",
									Version:    "2019c-0+deb9u1",
									FixVersion: "2021a-0+deb9u4",
									Status:     "fixed",
									Type:       "debian",
								}},
							},
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "CVE-2016-2779",
									Severity:         "HIGH",
									Title:            "util-linux: runuser tty hijack via TIOCSTI ioctl",
									Description:      "runuser in util-linux allows local users to escape to the parent session via a crafted TIOCSTI ioctl call, which pushes characters to the terminal's input buffer.",
									URL:              "https://avd.aquasec.com/nvd/cve-2016-2779",
									Score:            "7.8",
									PublishedDate:    newTime("2017-02-07T15:59:00Z"),
									LastModifiedDate: newTime("2019-01-04T14:14:00Z"),
								},
								Packages: []v1alpha1.Package{{
									Package:    "bsdutils",
									Version:    "1:2.29.2-1+deb9u1",
									FixVersion: "",
									Status:     "affected",
									Type:       "debian",
								}},
							},
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "GHSA-jmqm-f2gx-4fjv",
									Severity:         "MEDIUM",
									Title:            "Sensitive information exposure through logs in npm-registry-fetch",
									Description:      "Affected versions of `npm-registry-fetch` are vulnerable to an information exposure vulnerability through log files. The cli supports URLs like `\u003cprotocol\u003e://[\u003cuser\u003e[:\u003cpassword\u003e]@]\u003chostname\u003e[:\u003cport\u003e][:][/]\u003cpath\u003e`. The password value is not redacted and is printed to stdout and also to any generated log files.",
									URL:              "https://github.com/advisories/GHSA-jmqm-f2gx-4fjv",
									Score:            "5.3",
									PublishedDate:    nil,
									LastModifiedDate: nil,
								},
								Packages: []v1alpha1.Package{{
									Package:    "npm-registry-fetch",
									Version:    "4.0.4",
									FixVersion: "8.1.1, 4.0.5",
									Status:     "fixed",
									Type:       "node-pkg",
								}},
							},
						},
					},
				},
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:            "cluster-dockerioistioexamplesbookinfodetailsv11170-h9kcn",
						Namespace:       "ns",
						OwnerReferences: owners,
						Labels:          labels,
					},
					Spec: v1alpha2.VulnerabilityReportSpec{
						VulnerabilityReportCommon: v1alpha1.VulnerabilityReportCommon{
							Cluster:        "cluster",
							Image:          "docker.io/istio/examples-bookinfo-details-v1:1.17.0",
							Tags:           []string{"istio/examples-bookinfo-details-v1:1.17.0"},
							Digest:         "istio/examples-bookinfo-details-v1@sha256:2b081e3c86dd8105040ea1f2adcc94cb473f41249dc9c91ebc1c2885ddd56c13",
							Architecture:   "amd64",
							OS:             "linux",
							Distro:         &v1alpha1.Distro{Name: "debian", Version: "10.5"},
							Resources:      map[string][]string{"Deployment": {"apps/app2"}},
							TotalResources: 1,
							Summary:        v1alpha1.VulnerabilitySummary{Total: 2, High: 1, Low: 1},
						},
						TotalPackages:       2,
						TotalUniquePackages: 2,
						Vulnerabilities: []v1alpha2.Vulnerability{
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "CVE-2016-2781",
									Severity:         "LOW",
									Title:            "coreutils: Non-privileged session can escape to the parent session in chroot",
									Description:      "chroot in GNU coreutils, when used with --userspec, allows local users to escape to the parent session via a crafted TIOCSTI ioctl call, which pushes characters to the terminal's input buffer.",
									URL:              "https://avd.aquasec.com/nvd/cve-2016-2781",
									Score:            "6.5",
									PublishedDate:    newTime("2017-02-07T15:59:00Z"),
									LastModifiedDate: newTime("2021-02-25T17:15:00Z"),
								},
								Packages: []v1alpha1.Package{{
									Package:    "coreutils",
									Version:    "8.30-3",
									FixVersion: "",
									Status:     "will_not_fix",
									Type:       "debian",
								}},
							},
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "CVE-2023-28755",
									Severity:         "HIGH",
									Title:            "ReDoS vulnerability in URI",
									Description:      "A ReDoS issue was discovered in the URI component through 0.12.0 in Ruby through 3.2.1. The URI parser mishandles invalid URLs that have specific characters. It causes an increase in execution time for parsing strings to URI objects. The fixed versions are 0.12.1, 0.11.1, 0.10.2 and 0.10.0.1.",
									URL:              "https://avd.aquasec.com/nvd/cve-2023-28755",
									Score:            "5.3",
									PublishedDate:    newTime("2023-03-31T04:15:00Z"),
									LastModifiedDate: newTime("2023-05-30T17:17:00Z"),
								},
								Packages: []v1alpha1.Package{{
									Package:    "uri",
									Version:    "0.10.0",
									FixVersion: "~\u003e 0.10.0.1, ~\u003e 0.10.2, ~\u003e 0.11.1, \u003e= 0.12.1",
									Status:     "fixed",
									Type:       "gemspec",
								}},
							},
						},
					},
				},
				{
					TypeMeta: vulnReportTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:            "cluster-nginxsha256af296b188c7b7df99ba960ca614439c99cb7cf252ed7bbc23e90cfda59092305-h9kcn",
						Namespace:       "ns",
						OwnerReferences: owners,
						Labels:          labels,
					},
					Spec: v1alpha2.VulnerabilityReportSpec{
						VulnerabilityReportCommon: v1alpha1.VulnerabilityReportCommon{
							Cluster:        "cluster",
							Image:          "nginx@sha256:af296b188c7b7df99ba960ca614439c99cb7cf252ed7bbc23e90cfda59092305",
							Tags:           []string{"nginx:1.25.0"},
							Digest:         "nginx@sha256:af296b188c7b7df99ba960ca614439c99cb7cf252ed7bbc23e90cfda59092305",
							Architecture:   "amd64",
							OS:             "linux",
							Distro:         &v1alpha1.Distro{Name: "debian", Version: "11.7"},
							TotalResources: 1,
							Resources:      map[string][]string{"Deployment": {"default/nginx"}},
							Summary:        v1alpha1.VulnerabilitySummary{Total: 1, Medium: 1},
						},
						TotalPackages:       1,
						TotalUniquePackages: 1,
						Vulnerabilities: []v1alpha2.Vulnerability{
							{
								VulnerabilityCommon: v1alpha1.VulnerabilityCommon{
									ID:               "CVE-2023-3446",
									Severity:         "MEDIUM",
									Title:            "Excessive time spent checking DH keys and parameters",
									Description:      "Issue summary: Checking excessively long DH keys or parameters may be very slow.\n\nImpact summary: Applications that use the functions DH_check(), DH_check_ex()\nor EVP_PKEY_param_check() to check a DH key or DH parameters may experience long\ndelays. Where the key or parameters that are being checked have been obtained\nfrom an untrusted source this may lead to a Denial of Service.\n\nThe function DH_check() performs various checks on DH parameters. One of those\nchecks confirms that the modulus ('p' parameter) is not too large. Trying to use\na very large modulus is slow and OpenSSL will not normally use a modulus which\nis over 10,000 bits in length.\n\nHowever the DH_check() function checks numerous aspects of the key or parameters\nthat have been supplied. Some of those checks use the supplied modulus value\neven if it has already been found to be too large.\n\nAn application that calls DH_check() and supplies a key or parameters obtained\nfrom an untrusted source could be vulernable to a Denial of Service attack.\n\nThe function DH_check() is itself called by a number of other OpenSSL functions.\nAn application calling any of those other functions may similarly be affected.\nThe other functions affected by this are DH_check_ex() and\nEVP_PKEY_param_check().\n\nAlso vulnerable are the OpenSSL dhparam and pkeyparam command line applications\nwhen using the '-check' option.\n\nThe OpenSSL SSL/TLS implementation is not affected by this issue.\nThe OpenSSL 3.0 and 3.1 FIPS providers are not affected by this issue.",
									URL:              "https://avd.aquasec.com/nvd/cve-2023-3446",
									Score:            "5.3",
									PublishedDate:    newTime("2023-07-19T12:15:00Z"),
									LastModifiedDate: newTime("2023-08-16T08:15:00Z"),
								},
								Packages: []v1alpha1.Package{{
									Package:    "openssl",
									Version:    "1.1.1n-0+deb11u4",
									FixVersion: "",
									Status:     "fix_deferred",
									Type:       "debian",
								}},
							},
						},
					},
				},
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

func sortVulns(vulns []v1alpha2.VulnerabilityReport) {
	sort.Slice(vulns, func(i, j int) bool {
		return strings.Compare(vulns[i].Spec.Image, vulns[j].Spec.Image) == -1
	})
	for _, v := range vulns {
		for _, r := range v.Spec.Resources {
			sort.Strings(r)
		}
		for _, vuln := range v.Spec.Vulnerabilities {
			sort.Slice(vuln.Packages, func(i, j int) bool {
				return strings.Compare(vuln.Packages[i].String(), vuln.Packages[j].String()) == -1
			})
		}
		sort.Slice(v.Spec.Vulnerabilities, func(i, j int) bool {
			return strings.Compare(v.Spec.Vulnerabilities[i].ID, v.Spec.Vulnerabilities[j].ID) == -1
		})
	}
}

func newTime(s string) *metav1.Time {
	p, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &metav1.Time{Time: p}
}

func Test_cleanString(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "ok",
			arg:  `asdfghjklç"!@#$%¨&*()_+'1234567890-=¬¹²³£¢¬{[]}\§[]{}ªº´~^,.;/<>:?°\|àáãâ`,
			want: "asdfghjkl1234567890",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanString(tt.arg); got != tt.want {
				t.Errorf("cleanString() = %v, want %v", got, tt.want)
			}
		})
	}
}
