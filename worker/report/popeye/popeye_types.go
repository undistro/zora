package popeye

import (
	inspectv1a1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
)

// LevelToIssueSeverity maps Popeye's <Level> type to Inspect's
// <ClusterIssueSeverity>.
var LevelToIssueSeverity = [4]inspectv1a1.ClusterIssueSeverity{
	inspectv1a1.SeverityNone,
	inspectv1a1.SeverityLow,
	inspectv1a1.SeverityMedium,
	inspectv1a1.SeverityHigh,
}

// Level tracks lint check level.
type Level int

// The <Level> type and enum were imported from Popeye's <config> package, whose
// import path is <github.com/derailed/popeye/pkg/config>.
const (
	// OkLevel denotes no linting issues.
	OkLevel Level = iota
	// InfoLevel denotes FIY linting issues.
	InfoLevel
	// WarnLevel denotes a warning issue.
	WarnLevel
	// ErrorLevel denotes a serious issue.
	ErrorLevel
)

// Issue represents a Popeye sanitizer issue.
type Issue struct {
	GVR     string `json:"gvr"`
	Level   Level  `json:"level"`
	Message string `json:"message"`
}

// Sanitizer represents a Popeye sanitizer.
type Sanitizer struct {
	Sanitizer string              `json:"sanitizer"`
	Issues    map[string][]*Issue `json:"issues"`
}

// Popeye represents a Popeye report.
type Popeye struct {
	Sanitizers []*Sanitizer `json:"sanitizers"`
}

// Report wraps a Popeye report.
type Report struct {
	Popeye *Popeye `json:"popeye"`
}
