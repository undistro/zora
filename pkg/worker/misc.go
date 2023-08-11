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
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/undistro/zora/api/zora/v1alpha1"
	zora "github.com/undistro/zora/pkg/clientset/versioned"
	"github.com/undistro/zora/pkg/worker/report/marvin"
	"github.com/undistro/zora/pkg/worker/report/popeye"
)

// miscPlugins maps parse function by misc plugin name
var miscPlugins = map[string]func(ctx context.Context, reader io.Reader) ([]v1alpha1.ClusterIssueSpec, error){
	"popeye": popeye.Parse,
	"marvin": marvin.Parse,
}

var clusterIssueTypeMeta = metav1.TypeMeta{
	Kind:       "ClusterIssue",
	APIVersion: v1alpha1.SchemeGroupVersion.String(),
}

func handleMisconfiguration(ctx context.Context, cfg *config, results io.Reader, client *zora.Clientset) error {
	log := logr.FromContextOrDiscard(ctx)
	issues, err := parseMiscResults(ctx, cfg, results)
	if err != nil {
		return err
	}
	for _, issue := range issues {
		issue, err := client.ZoraV1alpha1().ClusterIssues(cfg.Namespace).Create(ctx, &issue, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create ClusterIssue %q: %v", issue.Name, err)
		}
		log.Info(fmt.Sprintf("cluster issue %q successfully created", issue.Name), "resource version", issue.ResourceVersion)
	}
	return nil
}

// parseMiscResults converts the given results into ClusterIssues
func parseMiscResults(ctx context.Context, cfg *config, results io.Reader) ([]v1alpha1.ClusterIssue, error) {
	parseFunc, ok := miscPlugins[cfg.PluginName]
	if !ok {
		return nil, errors.New(fmt.Sprintf("invalid plugin %q", cfg.PluginName))
	}
	specs, err := parseFunc(ctx, results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q results: %v", cfg.PluginName, err)
	}
	owner := metav1.OwnerReference{
		APIVersion: batchv1.SchemeGroupVersion.String(),
		Kind:       "Job",
		Name:       cfg.JobName,
		UID:        types.UID(cfg.JobUID),
	}
	issues := make([]v1alpha1.ClusterIssue, len(specs))
	for i := 0; i < len(specs); i++ {
		issues[i] = newClusterIssue(cfg, specs[i], owner)
	}
	return issues, nil
}

// newClusterIssue returns a new ClusterIssue from the given config, spec, and owner
func newClusterIssue(cfg *config, spec v1alpha1.ClusterIssueSpec, owner metav1.OwnerReference) v1alpha1.ClusterIssue {
	spec.Cluster = cfg.ClusterName
	return v1alpha1.ClusterIssue{
		TypeMeta: clusterIssueTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s-%s", cfg.ClusterName, strings.ToLower(spec.ID), cfg.suffix),
			Namespace:       cfg.Namespace,
			OwnerReferences: []metav1.OwnerReference{owner},
			Labels: map[string]string{
				v1alpha1.LabelScanID:   cfg.JobUID,
				v1alpha1.LabelCluster:  cfg.ClusterName,
				v1alpha1.LabelSeverity: string(spec.Severity),
				v1alpha1.LabelIssueID:  spec.ID,
				v1alpha1.LabelCategory: strings.ReplaceAll(spec.Category, " ", ""),
				v1alpha1.LabelPlugin:   cfg.PluginName,
				v1alpha1.LabelCustom:   strconv.FormatBool(spec.Custom),
			},
		},
		Spec: spec,
	}
}
