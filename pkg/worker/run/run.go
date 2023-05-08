// Copyright 2022 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	zorav1a1 "github.com/undistro/zora/api/zora/v1alpha1"
	"github.com/undistro/zora/pkg/clientset/versioned"
	"github.com/undistro/zora/pkg/worker/config"
	"github.com/undistro/zora/pkg/worker/report"
)

// CreateClusterIssues creates instances of <ClusterIssue> on the Kubernetes
// cluster which the client set points to.
func CreateClusterIssues(c *config.Config, ciarr []*zorav1a1.ClusterIssue) error {
	rconfig := ctrl.GetConfigOrDie()
	cset, err := versioned.NewForConfig(rconfig)
	if err != nil {
		return fmt.Errorf("Unable to instantiate REST config: %w", err)
	}
	for _, ci := range ciarr {
		_, err = cset.ZoraV1alpha1().ClusterIssues(c.ClusterIssuesNs).Create(context.Background(), ci, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to create <ClusterIssue> instance on cluster <%s>: %w", c.Cluster, err)
		}
	}
	return nil
}

// Done is used to check whether the "done" or "error" file have been created.
func Done(dpath string) bool {
	if finf, err := os.Stat(dpath); errors.Is(err, os.ErrNotExist) || finf.IsDir() {
		return false
	}
	return true
}

// Run performs a worker run, being the main point of entry for the component.
func Run(log logr.Logger) error {
	c, err := config.FromEnv()
	if err != nil {
		return fmt.Errorf("Unable to create config from environment: %w", err)
	}
	if err := c.HandleDonePath(); err != nil {
		return fmt.Errorf("Unable to ensure done path exists: %w", err)
	}

	for {
		if Done(c.ErrorPath) {
			return errors.New("Plugin crashed")
		}
		if Done(c.DonePath) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	fid, err := c.HandleResultsPath()
	if err != nil {
		return fmt.Errorf("Failed checking results path: %w", err)
	}

	ciarr, err := report.Parse(log, fid, c)
	if err != nil {
		return fmt.Errorf("Failed to parse results: %w", err)
	}
	if err = CreateClusterIssues(c, ciarr); err != nil {
		return fmt.Errorf("Failed to create issues: %w", err)
	}

	return nil
}
