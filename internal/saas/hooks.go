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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/undistro/zora/api/zora/v1alpha1"
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
	clusterName := clusterScan.Spec.ClusterRef.Name
	var lastScanIDs []string
	for _, cs := range scanList.Items {
		if !cs.DeletionTimestamp.IsZero() {
			continue
		}
		lastScanIDs = append(lastScanIDs, cs.Status.LastScanIDs(true)...)
	}

	ls := &metav1.LabelSelector{
		MatchLabels: map[string]string{v1alpha1.LabelCluster: clusterName},
	}
	if len(lastScanIDs) > 0 {
		ls.MatchExpressions = []metav1.LabelSelectorRequirement{{
			Key:      v1alpha1.LabelScanID,
			Operator: metav1.LabelSelectorOpIn,
			Values:   lastScanIDs,
		}}
	}
	lss, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return err
	}
	issueList := &v1alpha1.ClusterIssueList{}
	if err := c.List(ctx, issueList, ctrlClient.MatchingLabelsSelector{Selector: lss}); err != nil {
		return err
	}

	status := NewScanStatusWithIssues(scanList.Items, issueList.Items)
	if status == nil {
		return nil
	}
	if err := saasClient.PutClusterScan(ctx, clusterScan.Namespace, clusterName, status); err != nil {
		if serr, ok := err.(*saasError); ok {
			clusterScan.SetSaaSStatus(metav1.ConditionFalse, serr.Err, serr.Detail)
			return nil
		}
		clusterScan.SetSaaSStatus(metav1.ConditionFalse, "Error", err.Error())
		return err
	}
	clusterScan.SetSaaSStatus(metav1.ConditionTrue, "OK", "cluster scan successfully synced with SaaS")
	return nil
}
