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

package saas

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/undistro/zora/api/zora/v1alpha1"
	"github.com/undistro/zora/api/zora/v1alpha2"
)

type ClusterHook func(ctx context.Context, cluster *v1alpha1.Cluster) error

type ClusterScanHook func(ctx context.Context, clusterScan *v1alpha1.ClusterScan) error

func UpdateClusterHook(saasClient Client) ClusterHook {
	return func(ctx context.Context, cluster *v1alpha1.Cluster) error {
		cl := NewCluster(*cluster)
		return saasClient.PutCluster(ctx, cl)
	}
}

func DeleteClusterHook(saasClient Client) ClusterHook {
	return func(ctx context.Context, cluster *v1alpha1.Cluster) error {
		return saasClient.DeleteCluster(ctx, cluster.Namespace, cluster.Name)
	}
}

func UpdateClusterScanHook(saasClient Client, c ctrlClient.Client) ClusterScanHook {
	return func(ctx context.Context, clusterScan *v1alpha1.ClusterScan) error {
		scanList, err := getClusterScans(ctx, c, clusterScan.Namespace, clusterScan.Spec.ClusterRef.Name)
		if err != nil {
			return err
		}
		return updateClusterScan(saasClient, c, ctx, clusterScan, scanList)
	}
}

func DeleteClusterScanHook(saasClient Client, c ctrlClient.Client) ClusterScanHook {
	return func(ctx context.Context, clusterScan *v1alpha1.ClusterScan) error {
		clusterName := clusterScan.Spec.ClusterRef.Name
		scanList, err := getClusterScans(ctx, c, clusterScan.Namespace, clusterName)
		if err != nil {
			return err
		}
		if len(scanList.Items) <= 1 {
			return saasClient.DeleteClusterScan(ctx, clusterScan.Namespace, clusterName)
		}
		return updateClusterScan(saasClient, c, ctx, clusterScan, scanList)
	}
}

func getClusterScans(ctx context.Context, c ctrlClient.Client, namespace, clusterName string) (*v1alpha1.ClusterScanList, error) {
	scanList := &v1alpha1.ClusterScanList{}
	if err := c.List(ctx, scanList,
		ctrlClient.InNamespace(namespace),
		ctrlClient.MatchingLabels{v1alpha1.LabelCluster: clusterName},
	); err != nil {
		return nil, err
	}
	return scanList, nil
}

func updateClusterScan(saasClient Client, c ctrlClient.Client, ctx context.Context, clusterScan *v1alpha1.ClusterScan, scanList *v1alpha1.ClusterScanList) error {
	if err := pushVulns(saasClient, c, ctx, clusterScan); err != nil {
		return err
	}
	if updatedScan, err := pushMisconfigs(saasClient, c, ctx, clusterScan, scanList); err != nil {
		return err
	} else if !updatedScan {
		if err := pushStatusUpdate(saasClient, c, ctx, clusterScan, scanList); err != nil {
			return err
		}
	}
	return nil
}

func pushMisconfigs(saasClient Client, c ctrlClient.Client, ctx context.Context, clusterScan *v1alpha1.ClusterScan, scanList *v1alpha1.ClusterScanList) (bool, error) {
	clusterName := clusterScan.Spec.ClusterRef.Name
	var lastScanIDs []string
	for _, cs := range scanList.Items {
		if !cs.DeletionTimestamp.IsZero() {
			continue
		}
		lastScanIDs = append(lastScanIDs, cs.Status.LastScanIDs(true)...)
	}

	ls, err := buildLabelSelector(clusterName, lastScanIDs)
	if err != nil {
		return false, err
	}

	issueList := &v1alpha1.ClusterIssueList{}
	if err := c.List(ctx, issueList, ctrlClient.MatchingLabelsSelector{Selector: ls}); err != nil {
		return false, err
	}

	pluginProcessedResources := getMisconfigProcessedResources(clusterScan.Status.Plugins, issueList.Items)
	if reflect.DeepEqual(pluginProcessedResources, clusterScan.Status.ProcessedMisconfigurations) {
		log := log.FromContext(ctx)
		log.Info("Skipping misconfigurations, no changes from processed misconfigurations")
		return false, nil
	}

	status := NewScanStatusWithIssues(clusterScan, scanList.Items, issueList.Items)
	if status == nil {
		return true, nil
	}
	if err := saasClient.PutClusterScan(ctx, clusterScan.Namespace, clusterScan.Spec.ClusterRef.Name, status); err != nil {
		var serr *saasError
		if errors.As(err, &serr) {
			clusterScan.SetSaaSStatus(metav1.ConditionFalse, serr.Err, serr.Detail)
			return true, nil
		}
		clusterScan.SetSaaSStatus(metav1.ConditionFalse, "Error", err.Error())
		return false, err
	}
	clusterScan.Status.ProcessedMisconfigurations = pluginProcessedResources
	clusterScan.SetSaaSStatus(metav1.ConditionTrue, "OK", "cluster scan successfully synced with SaaS")
	return true, nil
}

func pushVulns(scl Client, cl ctrlClient.Client, ctx context.Context, cs *v1alpha1.ClusterScan) error {
	successfulScanIDs := cs.Status.LastScanIDs(true)
	if len(successfulScanIDs) == 0 {
		return nil
	}
	ls, err := buildLabelSelector(cs.Spec.ClusterRef.Name, successfulScanIDs)
	if err != nil {
		return err
	}

	metaList := &metav1.PartialObjectMetadataList{}
	metaList.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("VulnerabilityReportList"))
	if err := cl.List(ctx, metaList, ls); err != nil {
		return err
	}
	if len(metaList.Items) == 0 {
		return nil
	}

	pluginProcessedResources := getVulnerabilityProcessedResources(metaList.Items)
	if reflect.DeepEqual(pluginProcessedResources, cs.Status.ProcessedVulnerabilities) {
		log.FromContext(ctx).Info("Skipping vulnerabilities, no changes from processed vulnerabilities")
		return nil
	}
	split := map[string][]metav1.PartialObjectMetadata{}
	for _, i := range metaList.Items {
		if i.Labels != nil {
			if name, ok := i.Labels[v1alpha1.LabelName]; ok {
				split[name] = append(split[name], i)
				continue
			}
		}
		vulnReport := &v1alpha2.VulnerabilityReport{}
		if err := cl.Get(ctx, types.NamespacedName{Namespace: i.Namespace, Name: i.Name}, vulnReport); err != nil {
			return err
		}
		if vulnReport.SaaSStatusIsTrue() {
			continue
		}
		if err := scl.PutVulnerabilityReport(ctx, cs.Namespace, cs.Spec.ClusterRef.Name, vulnReport); err != nil {
			cs.SetSaaSStatus(metav1.ConditionFalse, "Error", err.Error())
			vulnReport.SetSaaSStatus(metav1.ConditionTrue, "Error", err.Error())
			_ = cl.Status().Update(ctx, vulnReport)
			return err
		}
		vulnReport.SetSaaSStatus(metav1.ConditionTrue, "OK", "VulnerabilityReport successfully pushed to SaaS")
		if err := cl.Status().Update(ctx, vulnReport); err != nil {
			return err
		}
	}
	for name, objs := range split {
		var merged *v1alpha2.VulnerabilityReport
		var reports []*v1alpha2.VulnerabilityReport
		for _, i := range objs {
			vr := &v1alpha2.VulnerabilityReport{}
			if err := cl.Get(ctx, types.NamespacedName{Namespace: i.Namespace, Name: i.Name}, vr); err != nil {
				return err
			}
			if vr.SaaSStatusIsTrue() {
				continue
			}
			reports = append(reports, vr)
			if merged == nil {
				merged = vr.DeepCopy()
				merged.Name = name
			} else {
				merged.Spec.Vulnerabilities = append(merged.Spec.Vulnerabilities, vr.Spec.Vulnerabilities...)
			}
		}
		if merged != nil {
			merged.Spec.Summarize()
		}
		if err := scl.PutVulnerabilityReport(ctx, cs.Namespace, cs.Spec.ClusterRef.Name, merged); err != nil {
			cs.SetSaaSStatus(metav1.ConditionFalse, "Error", err.Error())
			// update status of all reports
			for _, vr := range reports {
				vr.SetSaaSStatus(metav1.ConditionTrue, "Error", err.Error())
				_ = cl.Status().Update(ctx, vr)
			}
			return err
		}
		// update status of all reports
		for _, vr := range reports {
			vr.SetSaaSStatus(metav1.ConditionTrue, "OK", "VulnerabilityReport successfully pushed to SaaS")
			if err := cl.Status().Update(ctx, vr); err != nil {
				return err
			}
		}
	}
	cs.Status.ProcessedVulnerabilities = pluginProcessedResources
	cs.SetSaaSStatus(metav1.ConditionTrue, "OK", "cluster scan successfully synced with SaaS")
	return nil
}

func pushStatusUpdate(saasClient Client, c ctrlClient.Client, ctx context.Context, clusterScan *v1alpha1.ClusterScan, scanList *v1alpha1.ClusterScanList) error {
	status, _ := NewScanStatus(clusterScan, scanList.Items)
	processedPluginStatus := getPluginProcessedStatus(status)

	if reflect.DeepEqual(processedPluginStatus, clusterScan.Status.ProcessedPluginStatus) {
		log := log.FromContext(ctx)
		log.Info("Skipping status update, no changes from processed statuses")
		return nil
	}

	if status == nil {
		return nil
	}
	if err := saasClient.PutClusterStatus(ctx, clusterScan.Namespace, clusterScan.Spec.ClusterRef.Name, status); err != nil {
		var serr *saasError
		if errors.As(err, &serr) {
			clusterScan.SetSaaSStatus(metav1.ConditionFalse, serr.Err, serr.Detail)
			return nil
		}
		clusterScan.SetSaaSStatus(metav1.ConditionFalse, "Error", err.Error())
		return err
	}
	clusterScan.Status.ProcessedPluginStatus = processedPluginStatus
	clusterScan.SetSaaSStatus(metav1.ConditionTrue, "OK", "cluster scan successfully synced with SaaS")
	return nil
}

func getPluginProcessedStatus(statuses map[string]*PluginStatus) map[string]*v1alpha1.PluginScanProcessedStatus {
	result := map[string]*v1alpha1.PluginScanProcessedStatus{}
	for plugin, status := range statuses {
		processedStatus := &v1alpha1.PluginScanProcessedStatus{
			IssueCount:             status.IssueCount,
			LastSuccessfulScanTime: status.LastSuccessfulScanTime,
			LastFinishedScanTime:   status.LastFinishedScanTime,
			NextScheduleScanTime:   status.NextScheduleScanTime,
			Schedule:               status.Schedule,
			LastSuccessfulScanID:   status.LastSuccessfulScanID,
		}
		if status.Scan != nil {
			processedStatus.Scan = &v1alpha1.ProcessedScanStatus{
				Status:  string(status.Scan.Status),
				Message: status.Scan.Message,
				Suspend: status.Scan.Suspend,
				ID:      status.Scan.ID,
			}
		}
		result[plugin] = processedStatus
	}
	return result
}

func buildLabelSelector(clusterName string, scanIDs []string) (*ctrlClient.MatchingLabelsSelector, error) {
	sel := &metav1.LabelSelector{MatchLabels: map[string]string{v1alpha1.LabelCluster: clusterName}}
	if len(scanIDs) > 0 {
		sel.MatchExpressions = []metav1.LabelSelectorRequirement{{
			Key:      v1alpha1.LabelScanID,
			Operator: metav1.LabelSelectorOpIn,
			Values:   scanIDs,
		}}
	}
	ls, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		return nil, err
	}
	return &ctrlClient.MatchingLabelsSelector{Selector: ls}, nil
}

func getMisconfigProcessedResources(pluginStatus map[string]*v1alpha1.PluginScanStatus, clusterIssues []v1alpha1.ClusterIssue) map[string]v1alpha1.PluginScanProcessedResources {
	var pluginProcessedResources map[string]v1alpha1.PluginScanProcessedResources
	issuePluginIncluded := false
	for _, issue := range clusterIssues {
		plugin := issue.Labels[v1alpha1.LabelPlugin]
		if !issuePluginIncluded {
			_, issuePluginIncluded = pluginStatus[plugin]
		}
		if pluginProcessedResources == nil {
			pluginProcessedResources = map[string]v1alpha1.PluginScanProcessedResources{}
		}
		processedResources, ok := pluginProcessedResources[plugin]
		if !ok {
			processedResources = v1alpha1.PluginScanProcessedResources{}
			pluginProcessedResources[plugin] = processedResources
		}
		processedResources[fmt.Sprintf("%s/%s", issue.Namespace, issue.Name)] = issue.ResourceVersion
	}
	if issuePluginIncluded {
		return pluginProcessedResources
	} else {
		return nil
	}
}

func getVulnerabilityProcessedResources(vulnerabiltiyIssues []metav1.PartialObjectMetadata) map[string]v1alpha1.PluginScanProcessedResources {
	var pluginProcessedResources map[string]v1alpha1.PluginScanProcessedResources
	for _, issue := range vulnerabiltiyIssues {
		plugin := issue.Labels[v1alpha1.LabelPlugin]
		if pluginProcessedResources == nil {
			pluginProcessedResources = map[string]v1alpha1.PluginScanProcessedResources{}
		}
		processedResources, ok := pluginProcessedResources[plugin]
		if !ok {
			processedResources = v1alpha1.PluginScanProcessedResources{}
			pluginProcessedResources[plugin] = processedResources
		}
		processedResources[fmt.Sprintf("%s/%s", issue.Namespace, issue.Name)] = issue.ResourceVersion

	}
	return pluginProcessedResources
}
