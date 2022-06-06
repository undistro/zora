package popeye

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
	Level   Level  `json:"level"`
	Message string `json:"message"`
}

// Sanitizer represents a Popeye sanitizer.
type Sanitizer struct {
	Sanitizer string              `json:"sanitizer"`
	GVR       string              `json:"gvr"`
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
