package config

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFromEnv(t *testing.T) {
	cases := []struct {
		description string
		env         map[string]string
		config      *Config
		toerr       bool
	}{
		{
			description: "Full config with default 'done' path",
			env: map[string]string{
				PluginEnvVar:          "fake_plugin",
				ClusterEnvVar:         "fake_cluster",
				ClusterIssuesNsEnvVar: "fake_ns",
				JobEnvVar:             "fake_job",
				JobUIDEnvVar:          "fake_job_uid-666-666",
				PodEnvVar:             "fake_pod",
			},
			config: &Config{
				DonePath:        fmt.Sprintf("%s/done", DefaultDoneDir),
				ErrorPath:       fmt.Sprintf("%s/error", DefaultDoneDir),
				Plugin:          "fake_plugin",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job",
				JobUID:          "fake_job_uid-666-666",
				Pod:             "fake_pod",
			},
			toerr: false,
		},
		{
			description: "Config missing pod and job fields",
			env: map[string]string{
				PluginEnvVar:          "fake_plugin",
				ClusterEnvVar:         "fake_cluster",
				ClusterIssuesNsEnvVar: "fake_ns",
			},
			config: nil,
			toerr:  true,
		},
		{
			description: "Config missing cluster fields",
			env: map[string]string{
				DoneDirEnvVar: fmt.Sprintf("%s/done", DefaultDoneDir),
				PluginEnvVar:  "fake_plugin",
				JobEnvVar:     "fake_job",
				JobUIDEnvVar:  "fake_job_uid-666-666",
				PodEnvVar:     "fake_pod",
			},
			config: nil,
			toerr:  true,
		},
		{
			description: "Complete config with custom 'done' path",
			env: map[string]string{
				DoneDirEnvVar:         "/tmp/fake",
				PluginEnvVar:          "fake_plugin",
				ClusterEnvVar:         "fake_cluster",
				ClusterIssuesNsEnvVar: "fake_ns",
				JobEnvVar:             "fake_job",
				JobUIDEnvVar:          "fake_job_uid-666-666",
				PodEnvVar:             "fake_pod",
			},
			config: &Config{
				DonePath:        "/tmp/fake/done",
				ErrorPath:       "/tmp/fake/error",
				Plugin:          "fake_plugin",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job",
				JobUID:          "fake_job_uid-666-666",
				Pod:             "fake_pod",
			},
			toerr: false,
		},
	}

	for _, c := range cases {
		os.Clearenv()
		for e, v := range c.env {
			if err := os.Setenv(e, v); err != nil {
				t.Errorf("Setup failed on case: %s\n", c.description)
				t.Fatal(err)
			}
		}
		if config, err := FromEnv(); c.config != config && c.toerr {
			if err != nil {
				t.Error(err)
			}
			t.Errorf("Setup failed on case: %s\n", c.description)
			t.Errorf("Mismatch between expected and obtained values: \n%s\n", cmp.Diff(c.config, config))
		}
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		description string
		config      *Config
		toerr       bool
	}{
		{
			description: "Full config with default 'done' path",
			config: &Config{
				DonePath:        fmt.Sprintf("%s/done", DefaultDoneDir),
				ErrorPath:       fmt.Sprintf("%s/error", DefaultDoneDir),
				Plugin:          "popeye",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job",
				JobUID:          "fake_job_uid-666-666",
				Pod:             "fake_pod",
			},
			toerr: false,
		},
		{
			description: "Config missing pod field",
			config: &Config{
				DonePath:        fmt.Sprintf("%s/done", DefaultDoneDir),
				ErrorPath:       fmt.Sprintf("%s/error", DefaultDoneDir),
				Plugin:          "popeye",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job-666-666",
				JobUID:          "fake_job_uid",
			},
			toerr: true,
		},
		{
			description: "Config missing cluster fields",
			config: &Config{
				DonePath:  fmt.Sprintf("%s/done", DefaultDoneDir),
				ErrorPath: fmt.Sprintf("%s/error", DefaultDoneDir),
				Plugin:    "popeye",
				Job:       "fake_job-666-666",
				JobUID:    "fake_job_uid",
				Pod:       "fake_pod",
			},
			toerr: true,
		},
		{
			description: "Config missing <ErrorPath> field",
			config: &Config{
				DonePath:        "/tmp/fake/done",
				Plugin:          "unsupported_plugin",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job-666-666",
				JobUID:          "fake_job_uid",
				Pod:             "fake_pod",
			},
			toerr: true,
		},
		{
			description: "Config missing <DonePath> field",
			config: &Config{
				ErrorPath:       "/tmp/fake/error",
				Plugin:          "unsupported_plugin",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job-666-666",
				JobUID:          "fake_job_uid",
				Pod:             "fake_pod",
			},
			toerr: true,
		},
		{
			description: "Unsupported plugin in config",
			config: &Config{
				DonePath:        "/tmp/fake/done",
				ErrorPath:       "/tmp/fake/error",
				Plugin:          "unsupported_plugin",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "fake_job-666-666",
				JobUID:          "fake_job_uid",
				Pod:             "fake_pod",
			},
			toerr: true,
		},
		{
			description: "Invalid Job format",
			config: &Config{
				DonePath:        fmt.Sprintf("%s/done", DefaultDoneDir),
				ErrorPath:       fmt.Sprintf("%s/error", DefaultDoneDir),
				Plugin:          "popeye",
				Cluster:         "fake_cluster",
				ClusterIssuesNs: "fake_ns",
				Job:             "totally_fake_job",
				JobUID:          "fake_job_uid",
				Pod:             "fake_pod",
			},
			toerr: true,
		},
	}

	for _, c := range cases {
		if err := c.config.Validate(); err != nil && !c.toerr {
			t.Errorf("Setup failed on case: %s\n", c.description)
			t.Error(err)
		}
	}
}

func TestHandleDonePath(t *testing.T) {
	cases := []struct {
		description string
		config      *Config
		donedirmode os.FileMode
		toerr       bool
	}{
		{
			description: "Empty 'done' path",
			config:      &Config{},
			toerr:       true,
		},
		{
			description: "Path for 'done' file is created",
			config:      New(),
			toerr:       false,
		},
		{
			description: "Path for 'done' file already exists",
			config:      New(),
			donedirmode: os.FileMode(0755),
			toerr:       false,
		},
		{
			description: "Unable to check for existence of 'done' path",
			config:      New(),
			donedirmode: os.FileMode(0111),
			toerr:       true,
		},
	}

	for _, c := range cases {
		if c.donedirmode != 0 {
			if err := os.MkdirAll(path.Dir(c.config.DonePath), c.donedirmode); err != nil {
				t.Errorf("Setup failed on case: %s\n", c.description)
				t.Fatal(err)
			}
		}
		if err := c.config.HandleDonePath(); err != nil && !c.toerr {
			t.Errorf("Setup failed on case: %s\n", c.description)
			t.Error(err)
		}
		if c.donedirmode != 0 {
			if err := os.RemoveAll(path.Dir(c.config.DonePath)); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestHandleResultsPath(t *testing.T) {
	type fakeresults struct {
		path  string
		isdir bool
		mode  os.FileMode
	}
	cases := []struct {
		description string
		config      *Config
		results     *fakeresults
		toerr       bool
	}{
		{
			description: "Inexistent 'done' file",
			config: &Config{
				DonePath: "/no/way/this/path/exists/666/done",
			},
			toerr: true,
		},
		{
			description: "Empty 'done' file",
			config:      New(),
			results: &fakeresults{
				mode: os.FileMode(0644),
			},
			toerr: true,
		},
		{
			description: "White space in 'done' file",
			config:      New(),
			results: &fakeresults{
				path: "\t  /tmp/fake_results.txt \t   \t",
				mode: os.FileMode(0644),
			},
			toerr: false,
		},
		{
			description: "No read permission for results file",
			config:      New(),
			results: &fakeresults{
				path: "/tmp/fake_results.txt",
				mode: os.FileMode(0000),
			},
			toerr: true,
		},
		{
			description: "Results path is a directory",
			config:      New(),
			results: &fakeresults{
				path:  "/tmp/fake_results",
				mode:  os.FileMode(0550),
				isdir: true,
			},
			toerr: true,
		},
	}

	for _, c := range cases {
		if c.results != nil {
			if err := os.MkdirAll(path.Dir(c.config.DonePath), 0755); err != nil {
				t.Errorf("Setup failed on case: %s\n", c.description)
				t.Fatal(err)
			}
			if err := os.WriteFile(c.config.DonePath, []byte(c.results.path), 0644); err != nil {
				t.Errorf("Setup failed on case: %s\n", c.description)
				t.Fatal(err)
			}
			if len(c.results.path) != 0 {
				if c.results.isdir {
					if err := os.MkdirAll(c.results.path, 0755); err != nil {
						t.Errorf("Setup failed on case: %s\n", c.description)
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(strings.TrimSpace(c.results.path), []byte{}, c.results.mode); err != nil {
						t.Errorf("Setup failed on case: %s\n", c.description)
						t.Fatal(err)
					}
				}
			}
		}

		if _, err := c.config.HandleResultsPath(); (err != nil) != c.toerr {
			t.Errorf("Case: %s\n", c.description)
			t.Error(err)
		}

		if c.results != nil {
			if err := os.RemoveAll(path.Dir(c.config.DonePath)); err != nil {
				t.Errorf("Setup failed on case: %s\n", c.description)
				t.Fatal(err)
			}
			if err := os.RemoveAll(strings.TrimSpace(c.results.path)); err != nil {
				t.Errorf("Setup failed on case: %s\n", c.description)
				t.Fatal(err)
			}
		}
	}
}
