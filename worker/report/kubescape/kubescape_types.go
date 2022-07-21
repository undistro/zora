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

// Adapted from Armosec's package
// <github.com/armosec/opa-utils/reporthandling/results/v1/reportsummary>

// ControlSummary contains the scan Control with the scorefactor.
type ControlSummary struct {
	ScoreFactor float32 `json:"scoreFactor"`
}

// SummaryDetails contains a summary of the scan with the status and per
// Control summary.
type SummaryDetails struct {
	Status   ScanningStatus            `json:"status"`
	Controls map[string]ControlSummary `json:"controls,omitempty"`
}

// Adapted from Armosec's package
// <github.com/armosec/opa-utils/reporthandling/results/v1/resourcesresults>

// ResourceAssociatedRule holds the REGO rule associated status.
type ResourceAssociatedRule struct {
	Status ScanningStatus `json:"status"`
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
