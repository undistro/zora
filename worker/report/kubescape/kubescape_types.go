package kubescape

// Adapted from Armosec's package
// <github.com/armosec/opa-utils/reporthandling/apis>

type ScanningStatus string

const (
	StatusExcluded   ScanningStatus = "excluded"
	StatusIgnored    ScanningStatus = "ignored"
	StatusPassed     ScanningStatus = "passed"
	StatusSkipped    ScanningStatus = "skipped"
	StatusFailed     ScanningStatus = "failed"
	StatusUnknown    ScanningStatus = ""
	StatusIrrelevant ScanningStatus = "irrelevant"
	StatusError      ScanningStatus = "error"
)

type StatusInfo struct {
	InnerStatus ScanningStatus `json:"status,omitempty"`
	InnerInfo   string         `json:"info,omitempty"`
}

// Adapted from Armosec's package
// <github.com/armosec/opa-utils/reporthandling/helpers/v1>

// ResourceCounters stores the total amount of Kubernetes resources which,
// passed, failed or where excluded from the scan.
type ResourceCounters struct {
	PassedResources   int `json:"passedResources"`
	FailedResources   int `json:"failedResources"`
	ExcludedResources int `json:"excludedResources"`
}

// Adapted from Armosec's package
// <github.com/armosec/opa-utils/reporthandling/results/v1/reportsummary>

// ControlSummary contains a summary of a scan Control with the score,
// status, counters, etc.
type ControlSummary struct {
	ControlID        string           `json:"controlID"`
	Name             string           `json:"name"`
	Status           ScanningStatus   `json:"status"`
	Score            float32          `json:"score"`
	ScoreFactor      float32          `json:"scoreFactor"`
	ResourceCounters ResourceCounters `json:",inline"`
}

// FrameworkSummary contains a summary of a scan Framework with the score,
// versions, counters, etc.
type FrameworkSummary struct {
	Name             string                    `json:"name"`
	Status           ScanningStatus            `json:"status"`
	Score            float32                   `json:"score"`
	Version          string                    `json:"version"`
	Controls         map[string]ControlSummary `json:"controls,omitempty"`
	ResourceCounters ResourceCounters          `json:",inline"`
}

// SummaryDetails contains a summary of the scan with the score, versions,
// counters, etc.
type SummaryDetails struct {
	Score            float32                   `json:"score"`
	Status           ScanningStatus            `json:"status"`
	Frameworks       []FrameworkSummary        `json:"frameworks"`
	Controls         map[string]ControlSummary `json:"controls,omitempty"`
	ResourceCounters ResourceCounters          `json:",inline"`
}

// Adapted from Armosec's package
// <github.com/armosec/opa-utils/reporthandling/results/v1/resourcesresults>

// ResourceAssociatedRule holds a failed rule associated to a Kubernetes
// resource.
type ResourceAssociatedRule struct {
	Name                  string              `json:"name"`
	Status                ScanningStatus      `json:"status"`
	ControlConfigurations map[string][]string `json:"controlConfigurations,omitempty"` // new
}

// ResourceAssociatedControl holds the Control that is associated to a
// Kubernetes resource.
type ResourceAssociatedControl struct {
	ControlID               string                   `json:"controlID"`
	Name                    string                   `json:"name"`
	ResourceAssociatedRules []ResourceAssociatedRule `json:"rules,omitempty"`
}

// Result holds a Kubernetes resource from scan results with the Controls that
// where tested against it. The resource is formatted as:
// 		<api_group_version>/<namespace>/<kind>/<name>
type Result struct {
	ResourceID         string                      `json:"resourceID"`
	AssociatedControls []ResourceAssociatedControl `json:"controls,omitempty"`
}

// Adapted from Armosec's package
// <github.com/armosec/opa-utils/reporthandling>

// Resource stores a Kubernetes resourcs and a full copy of it in Json. The
// resource is formatted as:
// 		<api_group_version>/<namespace>/<kind>/<name>
type Resource struct {
	ResourceID string      `json:"resourceID"`
	Object     interface{} `json:"object"`
}

// Adapted from Armosec's package
// <github.com/armosec/opa-utils/reporthandling/v2>

// PostureReport represents a Kubescape scan result.
type PostureReport struct {
	SummaryDetails SummaryDetails `json:"summaryDetails,omitempty"`
	Results        []Result       `json:"results,omitempty"`
	Resources      []Resource     `json:"resources,omitempty"`
}
