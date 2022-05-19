package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	inspectv1a1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/worker/report/popeye"
)

const (
	DefaultDoneDir        = "/tmp/undistro-inspect/results"
	DoneDirEnvVar         = "DONE_DIR"
	PluginEnvVar          = "PLUGIN_NAME"
	ClusterEnvVar         = "CLUSTER_NAME"
	ClusterIssuesNsEnvVar = "CLUSTER_ISSUES_NAMESPACE"
)

// PluginParsers correlates plugins with their respective parsing functions.
var PluginParsers = map[string]func([]byte) ([]*inspectv1a1.ClusterIssueSpec, error){
	"popeye": popeye.Parse,
}

// Config stores information used by the worker to create a list of
// <ClusterIssue> instances, and to specify the "done" file path.
type Config struct {
	DonePath        string `json:"donePath"`
	Plugin          string `json:"plugin"`
	Cluster         string `json:"cluster"`
	ClusterIssuesNs string `json:"listClusterIssueNs"`
}

// New instanciates a new <Config> struct, with the default path for the
// "done" file.
func New() *Config {
	return &Config{DonePath: fmt.Sprintf("%s/done", DefaultDoneDir)}
}

// Validate ensures a <Config> instance has all its fields populated, and the
// plugin specified is supported by the worker.
func (r *Config) Validate() bool {
	if len(r.DonePath) == 0 || len(r.Cluster) == 0 || len(r.ClusterIssuesNs) == 0 ||
		len(r.Plugin) == 0 {
		return false
	}
	if _, ok := PluginParsers[r.Plugin]; !ok {
		return false
	}
	return true
}

// HandleDonePath ensures the directory wherefrom the "done" file will be
// written exists.
func (r *Config) HandleDonePath() error {
	if len(r.DonePath) == 0 {
		return errors.New("Empty <DonePath>")

	}
	dir := path.Dir(r.DonePath)
	if _, err := os.Stat(dir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Unable to check existance of dir <%d>: %w", dir, err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("Unable to create results dir <%d>: %w", dir, err)
	}
	return nil
}
