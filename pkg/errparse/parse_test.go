package errparse

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		description string
		plugin      string
		testfile    string
		toerr       bool
		errmsg      string
	}{
		// Popeye
		{
			description: "Invalid authentication token",
			plugin:      "popeye",
			testfile:    "testdata/popeye_err_1.txt",
			errmsg:      "the server has asked for the client to provide credentials",
		},
		{
			description: "Invalid cluster server address",
			plugin:      "popeye",
			testfile:    "testdata/popeye_err_2.txt",
			errmsg:      `Get "http://localhost:8080/version?timeout=30s": dial tcp 127.0.0.1:8080: connect: connection refused`,
		},
		{
			description: "Invalid cluster context",
			plugin:      "popeye",
			testfile:    "testdata/popeye_err_3.txt",
			errmsg:      "invalid configuration: context was not found for specified context: gke_undistro-dev_us-east1-a_zored",
		},
		{
			description: "Incorrect flag",
			plugin:      "popeye",
			testfile:    "testdata/popeye_err_4.txt",
			errmsg:      "Exec failed unknown flag: --brokenflag",
		},
		{
			description: "Non existent error data source",
			plugin:      "popeye",
			toerr:       true,
		},
		{
			description: "Non existent error data",
			plugin:      "popeye",
			testfile:    "testdata/dummy_err_1.txt",
			toerr:       true,
		},

		// Kubescape
		{
			description: "Invalid cluster server address",
			plugin:      "kubescape",
			testfile:    "testdata/kubescape_err_1.txt",
			errmsg:      `failed to discover API server information. error: Get "https://35.236.51.220/version?timeout=32s": x509: certificate signed by unknown authority`,
		},
		{
			description: "Invalid cluster user",
			plugin:      "kubescape",
			testfile:    "testdata/kubescape_err_2.txt",
			errmsg:      kubescapeBigErrMsg,
		},
		{
			description: "Invalid token data",
			plugin:      "kubescape",
			testfile:    "testdata/kubescape_err_3.txt",
			errmsg:      "failed connecting to Kubernetes cluster",
		},
		{
			description: "Invalid token",
			plugin:      "kubescape",
			testfile:    "testdata/kubescape_err_4.txt",
			errmsg:      "failed to discover API server information. error: the server has asked for the client to provide credentials",
		},
		{
			description: "Non existent error data source",
			plugin:      "kubescape",
			toerr:       true,
		},
		{
			description: "Non existent error data",
			plugin:      "kubescape",
			testfile:    "testdata/dummy_err_1.txt",
			toerr:       true,
		},

		// Generic
		{
			description: "Invalid plugin",
			plugin:      "invalid_plug",
			toerr:       true,
		},
		{
			description: "No plugin informed",
			toerr:       true,
		},
	}

	for _, c := range cases {
		f, err := os.Open(c.testfile)
		if err != nil && !os.IsNotExist(err) && !os.IsPermission(err) {
			t.Errorf("Setup failed on case: %s\n", c.description)
			t.Fatal(err)
		}
		if errmsg, err := Parse(f, c.plugin); (err != nil) != c.toerr || c.errmsg != errmsg {
			if err != nil {
				t.Error(err)
			}
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Expected:\n\t<%s>\nBut got: \n\t<%s>", c.errmsg, errmsg)
		}
		f.Close()
	}
}
