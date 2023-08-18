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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	zora "github.com/undistro/zora/pkg/clientset/versioned"
)

func Run(ctx context.Context) error {
	cfg, err := configFromEnv()
	if err != nil {
		return fmt.Errorf("failed to get config from env: %v", err)
	}
	client, err := getZoraClient()
	if err != nil {
		return err
	}
	results, err := gatherResults(ctx, cfg.WaitInterval, cfg.DoneFile, cfg.ErrorFile)
	if err != nil {
		return fmt.Errorf("failed to gather results: %v", err)
	}
	switch cfg.PluginType {
	case "misconfiguration":
		return handleMisconfiguration(ctx, cfg, results, client)
	case "vulnerability":
		return handleVulnerability(ctx, cfg, results, client)
	}
	return nil
}

// getZoraClient returns Zora clientset
func getZoraClient() (*zora.Clientset, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %v", err)
	}
	client, err := zora.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build zora client: %v", err)
	}
	return client, nil
}

// gatherResults waits for the "done file" to be present, then returns the indicated "results file" reader
func gatherResults(ctx context.Context, interval time.Duration, doneFile, errorFile string) (io.Reader, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("doneFile", doneFile, "errorFile", errorFile)
	for {
		if ok, err := fileExists(errorFile); ok || err != nil {
			if err == nil {
				err = errors.New("error file detected")
			}
			return nil, err
		}
		done, err := fileExists(doneFile)
		if err != nil {
			return nil, fmt.Errorf("failed to check if done file exists: %v", err)
		}
		if done {
			log.Info("done file detected")
			break
		}
		log.Info(fmt.Sprintf("done file not found, waiting %s...", interval.String()))
		time.Sleep(interval)
	}
	return readResultsFile(doneFile)
}

// readResultsFile returns the "results file" reader indicated into "done file"
func readResultsFile(doneFile string) (io.Reader, error) {
	b, err := os.ReadFile(doneFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read done file %q: %v", doneFile, err)
	}
	resultsFile := string(bytes.TrimSpace(b))
	file, err := os.Open(resultsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read results file %q: %v", resultsFile, err)
	}
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat results file %q: %v", resultsFile, err)
	}
	if info.IsDir() {
		return nil, errors.New("results file is a directory")
	}
	return file, nil
}

// fileExists returns true if the file at the given path exists and is not a directory
func fileExists(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return false, ignoreNotExist(err)
	}
	if fi.IsDir() {
		return false, errors.New("is directory")
	}
	return true, nil
}

func ignoreNotExist(err error) error {
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func ownerReference(cfg *config) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: batchv1.SchemeGroupVersion.String(),
		Kind:       "Job",
		Name:       cfg.JobName,
		UID:        types.UID(cfg.JobUID),
	}
}
