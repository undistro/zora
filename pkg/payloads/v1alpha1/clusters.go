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

package v1alpha1

import (
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/undistro/zora/apis/zora/v1alpha1"
	"github.com/undistro/zora/pkg/formats"
)

type ScanStatusType string

const (
	Failed  ScanStatusType = "failed"
	Unknown ScanStatusType = "unknown"
	Scanned ScanStatusType = "scanned"
)

type Cluster struct {
	ApiVersion        string                   `json:"apiVersion"`
	Name              string                   `json:"name"`
	Namespace         string                   `json:"namespace"`
	Environment       string                   `json:"environment"`
	Provider          string                   `json:"provider"`
	Region            string                   `json:"region"`
	TotalNodes        *int                     `json:"totalNodes"`
	Version           string                   `json:"version"`
	Connection        *ConnectionStatus        `json:"connection"`
	Resources         *Resources               `json:"resources"`
	CreationTimestamp metav1.Time              `json:"creationTimestamp"`
	TotalIssues       *int                     `json:"totalIssues"`
	PluginStatus      map[string]*PluginStatus `json:"pluginStatus"`
}

type PluginStatus struct {
	Scan                   *ScanStatus      `json:"scan"`
	IssueCount             *int             `json:"issueCount"`
	Issues                 []ResourcedIssue `json:"issues"`
	LastSuccessfulScanTime *metav1.Time     `json:"lastSuccessfulScanTime"`
	LastFinishedScanTime   *metav1.Time     `json:"lastFinishedScanTime"`
	NextScheduleScanTime   *metav1.Time     `json:"nextScheduleScanTime"`
	Schedule               string           `json:"schedule"`
}

type NsName struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type ResourcedIssue struct {
	Issue     `json:",inline"`
	Resources map[string][]NsName `json:"resources"`
}

type Resources struct {
	Discovered bool      `json:"discovered"`
	Message    string    `json:"message"`
	Memory     *Resource `json:"memory"`
	CPU        *Resource `json:"cpu"`
}

type Resource struct {
	Available       string `json:"available"`
	Usage           string `json:"usage"`
	UsagePercentage int32  `json:"usagePercentage"`
}

type ScanStatus struct {
	Status  ScanStatusType `json:"status"`
	Message string         `json:"message"`
	Suspend bool           `json:"suspend"`
}

type ConnectionStatus struct {
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
}

//NewCluster returns a Cluster without pluginStatus and issues
func NewCluster(cluster v1alpha1.Cluster) Cluster {
	cl := Cluster{
		ApiVersion:        "v1alpha1",
		Name:              cluster.Name,
		Namespace:         cluster.Namespace,
		Environment:       cluster.Labels[v1alpha1.LabelEnvironment],
		Provider:          cluster.Status.Provider,
		Region:            cluster.Status.Region,
		TotalNodes:        cluster.Status.TotalNodes,
		Version:           cluster.Status.KubernetesVersion,
		CreationTimestamp: cluster.Status.CreationTimestamp,
		Resources:         &Resources{},
		Connection:        &ConnectionStatus{},
	}

	for _, c := range cluster.Status.Conditions {
		switch c.Type {
		case v1alpha1.ClusterReady:
			cl.Connection.Connected = c.Status == metav1.ConditionTrue
			cl.Connection.Message = c.Message
		case v1alpha1.ClusterResourcesDiscovered:
			cl.Resources.Discovered = c.Status == metav1.ConditionTrue
			cl.Resources.Message = c.Message
		}
	}

	if cpu, ok := cluster.Status.Resources[corev1.ResourceCPU]; ok {
		cl.Resources.CPU = &Resource{
			Available:       formats.CPU(cpu.Available),
			Usage:           formats.CPU(cpu.Usage),
			UsagePercentage: cpu.UsagePercentage,
		}
	}
	if mem, ok := cluster.Status.Resources[corev1.ResourceMemory]; ok {
		cl.Resources.Memory = &Resource{
			Available:       formats.Memory(mem.Available),
			Usage:           formats.Memory(mem.Usage),
			UsagePercentage: mem.UsagePercentage,
		}
	}

	return cl
}

//NewClusterWithScans returns a Cluster with pluginStatus and without issues
func NewClusterWithScans(cluster v1alpha1.Cluster, scans []v1alpha1.ClusterScan) Cluster {
	cl := NewCluster(cluster)
	ps, totalIssues := NewScanStatus(scans)
	cl.PluginStatus = ps
	cl.TotalIssues = totalIssues
	return cl
}

//NewClusterWithIssues returns a Cluster with pluginStatus and their issues
func NewClusterWithIssues(cluster v1alpha1.Cluster, scans []v1alpha1.ClusterScan, issues []v1alpha1.ClusterIssue) Cluster {
	c := NewClusterWithScans(cluster, scans)
	for _, i := range issues {
		c.PluginStatus[i.Labels[v1alpha1.LabelPlugin]].Issues = append(
			c.PluginStatus[i.Labels[v1alpha1.LabelPlugin]].Issues,
			NewResourcedIssue(i),
		)
	}
	return c
}

func NewScanStatus(scans []v1alpha1.ClusterScan) (map[string]*PluginStatus, *int) {
	var pluginStatus map[string]*PluginStatus
	var totalIssues *int

	for _, cs := range scans {
		if cs.Status.TotalIssues != nil {
			if totalIssues == nil {
				totalIssues = new(int)
			}
			*totalIssues += *cs.Status.TotalIssues
		}
		for p, s := range cs.Status.Plugins {
			if pluginStatus[p] == nil {
				if pluginStatus == nil {
					pluginStatus = map[string]*PluginStatus{}
				}
				pluginStatus[p] = &PluginStatus{
					Scan: &ScanStatus{
						Status: Unknown,
					},
				}
			}
			pluginStatus[p].Scan.Suspend = s.Suspend
			pluginStatus[p].Schedule = s.Schedule

			if s.IssueCount != nil {
				if pluginStatus[p].IssueCount == nil {
					pluginStatus[p].IssueCount = new(int)
				}
				*pluginStatus[p].IssueCount += *s.IssueCount
			}

			switch s.LastFinishedStatus {
			case string(batchv1.JobComplete):
				pluginStatus[p].Scan.Status = Scanned
			case string(batchv1.JobFailed):
				pluginStatus[p].Scan.Status = Failed
				pluginStatus[p].Scan.Message = s.LastErrorMsg
			case "":
				pluginStatus[p].Scan.Message = "Scan not finished"
			}

			if pluginStatus[p].LastSuccessfulScanTime == nil ||
				pluginStatus[p].LastSuccessfulScanTime.Before(cs.Status.LastSuccessfulTime) {
				pluginStatus[p].LastSuccessfulScanTime = cs.Status.LastSuccessfulTime
			}
			if pluginStatus[p].LastFinishedScanTime == nil ||
				pluginStatus[p].LastFinishedScanTime.Before(cs.Status.LastFinishedTime) {
				pluginStatus[p].LastFinishedScanTime = cs.Status.LastFinishedTime
			}
			if pluginStatus[p].NextScheduleScanTime == nil ||
				cs.Status.NextScheduleTime.Before(pluginStatus[p].NextScheduleScanTime) {
				pluginStatus[p].NextScheduleScanTime = cs.Status.NextScheduleTime
			}
		}
	}

	return pluginStatus, totalIssues
}

func NewScanStatusWithIssues(scans []v1alpha1.ClusterScan, issues []v1alpha1.ClusterIssue) map[string]*PluginStatus {
	pluginStatus, _ := NewScanStatus(scans)
	for _, i := range issues {
		pluginStatus[i.Labels[v1alpha1.LabelPlugin]].Issues = append(
			pluginStatus[i.Labels[v1alpha1.LabelPlugin]].Issues,
			NewResourcedIssue(i),
		)
	}
	return pluginStatus
}

func NewResourcedIssue(i v1alpha1.ClusterIssue) ResourcedIssue {
	ri := ResourcedIssue{}
	ri.Issue = NewIssue(i)
	for r, narr := range i.Spec.Resources {
		for _, nspacedn := range narr {
			ns := strings.Split(nspacedn, "/")
			if len(ns) == 1 {
				ns = append([]string{""}, ns[0])
			}
			if ri.Resources == nil {
				ri.Resources = map[string][]NsName{
					r: {{
						Name:      ns[1],
						Namespace: ns[0],
					}},
				}
			} else {
				ri.Resources[r] = append(ri.Resources[r],
					NsName{
						Name:      ns[1],
						Namespace: ns[0],
					},
				)
			}
		}
	}
	return ri
}
