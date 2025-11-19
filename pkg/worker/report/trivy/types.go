// Copyright 2025 Undistro Authors
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
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Report represents a Trivy v0.57.1 report format.
// Ref: https://github.com/aquasecurity/trivy/blob/v0.57.1/pkg/k8s/report/report.go#L40
type Report struct {
	ClusterName string
	Resources   []Resource `json:",omitempty"`
}

type Resource struct {
	Namespace string `json:",omitempty"`
	Kind      string
	Name      string
	Metadata  []Metadata `json:",omitempty"`
	Results   []Result   `json:",omitempty"`
	Error     string     `json:",omitempty"`
}

type Metadata struct {
	Size        int64         `json:",omitempty"`
	OS          *OS           `json:",omitempty"`
	ImageID     string        `json:",omitempty"`
	DiffIDs     []string      `json:",omitempty"`
	RepoTags    []string      `json:",omitempty"`
	RepoDigests []string      `json:",omitempty"`
	ImageConfig v1.ConfigFile `json:",omitempty"`
}

type OS struct {
	Family string
	Name   string
}

type Result struct {
	Target          string                  `json:"Target"`
	Class           string                  `json:"Class,omitempty"`
	Type            string                  `json:"Type,omitempty"`
	Vulnerabilities []DetectedVulnerability `json:"Vulnerabilities,omitempty"`
}

func (r *Result) IsOS() bool {
	return r.Class == "os-pkgs"
}

type DetectedVulnerability struct {
	VulnerabilityID  string   `json:",omitempty"`
	VendorIDs        []string `json:",omitempty"`
	PkgID            string   `json:",omitempty"` // It is used to construct dependency graph.
	PkgName          string   `json:",omitempty"`
	PkgPath          string   `json:",omitempty"` // This field is populated in the case of language-specific packages such as egg/wheel and gemspec
	InstalledVersion string   `json:",omitempty"`
	FixedVersion     string   `json:",omitempty"`
	Status           string   `json:",omitempty"`
	PrimaryURL       string   `json:",omitempty"`

	Title            string          `json:",omitempty"`
	Description      string          `json:",omitempty"`
	Severity         string          `json:",omitempty"` // Selected from VendorSeverity, depending on a scan target
	CweIDs           []string        `json:",omitempty"` // e.g. CWE-78, CWE-89
	CVSS             map[string]CVSS `json:",omitempty"`
	References       []string        `json:",omitempty"`
	PublishedDate    *time.Time      `json:",omitempty"` // Take from NVD
	LastModifiedDate *time.Time      `json:",omitempty"` // Take from NVD
}

type CVSS struct {
	V2Vector  string  `json:"V2Vector,omitempty"`
	V3Vector  string  `json:"V3Vector,omitempty"`
	V40Vector string  `json:"V40Vector,omitempty"`
	V2Score   float64 `json:"V2Score,omitempty"`
	V3Score   float64 `json:"V3Score,omitempty"`
	V40Score  float64 `json:"V40Score,omitempty"`
}
