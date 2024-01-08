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

	"github.com/google/go-cmp/cmp"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/undistro/zora/api/zora/v1alpha1"
)

var labels = map[string]string{
	v1alpha1.LabelScanID:  "50c8957e-c9e1-493a-9fa4-d0786deea017",
	v1alpha1.LabelCluster: "cluster",
	v1alpha1.LabelPlugin:  "trivy",
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
						Name:            "cluster-registryk8siokubeapiserverv1273-h9kcn",
						Namespace:       "ns",
						OwnerReferences: owners,
						Labels:          labels,
					},
					Spec: v1alpha1.VulnerabilityReportSpec{
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
						Vulnerabilities: []v1alpha1.Vulnerability{
							{
								ID:          "CVE-2022-41723",
								Severity:    "HIGH",
								Title:       "avoid quadratic complexity in HPACK decoding",
								Description: "A maliciously crafted HTTP/2 stream could cause excessive CPU consumption in the HPACK decoder, sufficient to cause a denial of service from a small number of small requests.",
								Package:     "golang.org/x/net",
								Version:     "v0.0.0-20220722155237-a158d28d115b",
								FixVersion:  "0.7.0",
								URL:         "https://avd.aquasec.com/nvd/cve-2022-41723",
								Status:      "fixed",
								Type:        "gobinary",
								Score:       "7.5",
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
					Spec: v1alpha1.VulnerabilityReportSpec{
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
						Vulnerabilities: []v1alpha1.Vulnerability{
							{
								ID:          "CVE-2022-4450",
								Severity:    "HIGH",
								Title:       "double free after calling PEM_read_bio_ex",
								Description: "The function PEM_read_bio_ex() reads a PEM file from a BIO and parses and decodes the \"name\" (e.g. \"CERTIFICATE\"), any header data and the payload data. If the function succeeds then the \"name_out\", \"header\" and \"data\" arguments are populated with pointers to buffers containing the relevant decoded data. The caller is responsible for freeing those buffers. It is possible to construct a PEM file that results in 0 bytes of payload data. In this case PEM_read_bio_ex() will return a failure code but will populate the header argument with a pointer to a buffer that has already been freed. If the caller also frees this buffer then a double free will occur. This will most likely lead to a crash. This could be exploited by an attacker who has the ability to supply malicious PEM files for parsing to achieve a denial of service attack. The functions PEM_read_bio() and PEM_read() are simple wrappers around PEM_read_bio_ex() and therefore these functions are also directly affected. These functions are also called indirectly by a number of other OpenSSL functions including PEM_X509_INFO_read_bio_ex() and SSL_CTX_use_serverinfo_file() which are also vulnerable. Some OpenSSL internal uses of these functions are not vulnerable because the caller does not free the header argument if PEM_read_bio_ex() returns a failure code. These locations include the PEM_read_bio_TYPE() functions as well as the decoders introduced in OpenSSL 3.0. The OpenSSL asn1parse command line application is also impacted by this issue.",
								Package:     "libssl1.1",
								Version:     "1.1.1s-r0",
								FixVersion:  "1.1.1t-r0",
								URL:         "https://avd.aquasec.com/nvd/cve-2022-4450",
								Status:      "fixed",
								Type:        "alpine",
								Score:       "7.5",
							},
							{
								ID:          "CVE-2022-4450",
								Severity:    "HIGH",
								Title:       "double free after calling PEM_read_bio_ex",
								Description: "The function PEM_read_bio_ex() reads a PEM file from a BIO and parses and decodes the \"name\" (e.g. \"CERTIFICATE\"), any header data and the payload data. If the function succeeds then the \"name_out\", \"header\" and \"data\" arguments are populated with pointers to buffers containing the relevant decoded data. The caller is responsible for freeing those buffers. It is possible to construct a PEM file that results in 0 bytes of payload data. In this case PEM_read_bio_ex() will return a failure code but will populate the header argument with a pointer to a buffer that has already been freed. If the caller also frees this buffer then a double free will occur. This will most likely lead to a crash. This could be exploited by an attacker who has the ability to supply malicious PEM files for parsing to achieve a denial of service attack. The functions PEM_read_bio() and PEM_read() are simple wrappers around PEM_read_bio_ex() and therefore these functions are also directly affected. These functions are also called indirectly by a number of other OpenSSL functions including PEM_X509_INFO_read_bio_ex() and SSL_CTX_use_serverinfo_file() which are also vulnerable. Some OpenSSL internal uses of these functions are not vulnerable because the caller does not free the header argument if PEM_read_bio_ex() returns a failure code. These locations include the PEM_read_bio_TYPE() functions as well as the decoders introduced in OpenSSL 3.0. The OpenSSL asn1parse command line application is also impacted by this issue.",
								Package:     "libcrypto1.1",
								Version:     "1.1.1s-r0",
								FixVersion:  "1.1.1t-r0",
								URL:         "https://avd.aquasec.com/nvd/cve-2022-4450",
								Status:      "fixed",
								Type:        "alpine",
								Score:       "7.5",
							},
							{
								ID:          "CVE-2023-37920",
								Severity:    "CRITICAL",
								Title:       "Removal of e-Tugra root certificate",
								Description: "Certifi is a curated collection of Root Certificates for validating the trustworthiness of SSL certificates while verifying the identity of TLS hosts. Certifi prior to version 2023.07.22 recognizes \"e-Tugra\" root certificates. e-Tugra's root certificates were subject to an investigation prompted by reporting of security issues in their systems. Certifi 2023.07.22 removes root certificates from \"e-Tugra\" from the root store.",
								Package:     "certifi",
								Version:     "2022.12.7",
								FixVersion:  "2023.7.22",
								URL:         "https://avd.aquasec.com/nvd/cve-2023-37920",
								Status:      "fixed",
								Type:        "python-pkg",
								Score:       "9.8",
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
					Spec: v1alpha1.VulnerabilityReportSpec{
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
						Vulnerabilities: []v1alpha1.Vulnerability{
							{
								ID:          "DLA-3051-1",
								Severity:    "UNKNOWN",
								Title:       "tzdata - new timezone database",
								Description: "",
								Package:     "tzdata",
								Version:     "2019c-0+deb9u1",
								FixVersion:  "2021a-0+deb9u4",
								URL:         "",
								Status:      "fixed",
								Type:        "debian",
							},
							{
								ID:          "CVE-2016-2779",
								Severity:    "HIGH",
								Title:       "util-linux: runuser tty hijack via TIOCSTI ioctl",
								Description: "runuser in util-linux allows local users to escape to the parent session via a crafted TIOCSTI ioctl call, which pushes characters to the terminal's input buffer.",
								Package:     "bsdutils",
								Version:     "1:2.29.2-1+deb9u1",
								FixVersion:  "",
								URL:         "https://avd.aquasec.com/nvd/cve-2016-2779",
								Status:      "affected",
								Type:        "debian",
								Score:       "7.8",
							},
							{
								ID:          "GHSA-jmqm-f2gx-4fjv",
								Severity:    "MEDIUM",
								Title:       "Sensitive information exposure through logs in npm-registry-fetch",
								Description: "Affected versions of `npm-registry-fetch` are vulnerable to an information exposure vulnerability through log files. The cli supports URLs like `\u003cprotocol\u003e://[\u003cuser\u003e[:\u003cpassword\u003e]@]\u003chostname\u003e[:\u003cport\u003e][:][/]\u003cpath\u003e`. The password value is not redacted and is printed to stdout and also to any generated log files.",
								Package:     "npm-registry-fetch",
								Version:     "4.0.4",
								FixVersion:  "8.1.1, 4.0.5",
								URL:         "https://github.com/advisories/GHSA-jmqm-f2gx-4fjv",
								Status:      "fixed",
								Type:        "node-pkg",
								Score:       "5.3",
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
					Spec: v1alpha1.VulnerabilityReportSpec{
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
						Vulnerabilities: []v1alpha1.Vulnerability{
							{
								ID:          "CVE-2016-2781",
								Severity:    "LOW",
								Title:       "coreutils: Non-privileged session can escape to the parent session in chroot",
								Description: "chroot in GNU coreutils, when used with --userspec, allows local users to escape to the parent session via a crafted TIOCSTI ioctl call, which pushes characters to the terminal's input buffer.",
								Package:     "coreutils",
								Version:     "8.30-3",
								FixVersion:  "",
								URL:         "https://avd.aquasec.com/nvd/cve-2016-2781",
								Status:      "will_not_fix",
								Type:        "debian",
								Score:       "6.5",
							},
							{
								ID:          "CVE-2023-28755",
								Severity:    "HIGH",
								Title:       "ReDoS vulnerability in URI",
								Description: "A ReDoS issue was discovered in the URI component through 0.12.0 in Ruby through 3.2.1. The URI parser mishandles invalid URLs that have specific characters. It causes an increase in execution time for parsing strings to URI objects. The fixed versions are 0.12.1, 0.11.1, 0.10.2 and 0.10.0.1.",
								Package:     "uri",
								Version:     "0.10.0",
								FixVersion:  "~\u003e 0.10.0.1, ~\u003e 0.10.2, ~\u003e 0.11.1, \u003e= 0.12.1",
								URL:         "https://avd.aquasec.com/nvd/cve-2023-28755",
								Status:      "fixed",
								Type:        "gemspec",
								Score:       "5.3",
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
					Spec: v1alpha1.VulnerabilityReportSpec{
						Cluster:        "cluster",
						Image:          "nginx@sha256:af296b188c7b7df99ba960ca614439c99cb7cf252ed7bbc23e90cfda59092305",
						Tags:           []string{"nginx:1.25.0"},
						Digest:         "nginx@sha256:af296b188c7b7df99ba960ca614439c99cb7cf252ed7bbc23e90cfda59092305",
						Architecture:   "amd64",
						OS:             "linux",
						Distro:         &v1alpha1.Distro{Name: "debian", Version: "11.7"},
						TotalResources: 1,
						Resources:      map[string][]string{"Deployment": {"default/nginx"}},
						Vulnerabilities: []v1alpha1.Vulnerability{
							{
								ID:          "CVE-2023-3446",
								Severity:    "MEDIUM",
								Title:       "Excessive time spent checking DH keys and parameters",
								Description: "Issue summary: Checking excessively long DH keys or parameters may be very slow.\n\nImpact summary: Applications that use the functions DH_check(), DH_check_ex()\nor EVP_PKEY_param_check() to check a DH key or DH parameters may experience long\ndelays. Where the key or parameters that are being checked have been obtained\nfrom an untrusted source this may lead to a Denial of Service.\n\nThe function DH_check() performs various checks on DH parameters. One of those\nchecks confirms that the modulus ('p' parameter) is not too large. Trying to use\na very large modulus is slow and OpenSSL will not normally use a modulus which\nis over 10,000 bits in length.\n\nHowever the DH_check() function checks numerous aspects of the key or parameters\nthat have been supplied. Some of those checks use the supplied modulus value\neven if it has already been found to be too large.\n\nAn application that calls DH_check() and supplies a key or parameters obtained\nfrom an untrusted source could be vulernable to a Denial of Service attack.\n\nThe function DH_check() is itself called by a number of other OpenSSL functions.\nAn application calling any of those other functions may similarly be affected.\nThe other functions affected by this are DH_check_ex() and\nEVP_PKEY_param_check().\n\nAlso vulnerable are the OpenSSL dhparam and pkeyparam command line applications\nwhen using the '-check' option.\n\nThe OpenSSL SSL/TLS implementation is not affected by this issue.\nThe OpenSSL 3.0 and 3.1 FIPS providers are not affected by this issue.",
								Package:     "openssl",
								Version:     "1.1.1n-0+deb11u4",
								FixVersion:  "",
								URL:         "https://avd.aquasec.com/nvd/cve-2023-3446",
								Status:      "fix_deferred",
								Type:        "debian",
								Score:       "5.3",
							},
						},
						Summary: v1alpha1.VulnerabilitySummary{Total: 1, Medium: 1},
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
