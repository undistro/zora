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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// +k8s:deepcopy-gen=true
type Cluster struct {
	ApiVersion        string            `json:"apiVersion"`
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Environment       string            `json:"environment"`
	Provider          string            `json:"provider"`
	Region            string            `json:"region"`
	TotalNodes        *int              `json:"totalNodes"`
	Version           string            `json:"version"`
	Connection        *ConnectionStatus `json:"connection"`
	Resources         *Resources        `json:"resources"`
	CreationTimestamp metav1.Time       `json:"creationTimestamp"`
	// Deprecated
	TotalIssues  *int                     `json:"totalIssues"`
	PluginStatus map[string]*PluginStatus `json:"pluginStatus"`
}

// +k8s:deepcopy-gen=true
type PluginStatus struct {
	Scan                   *ScanStatus      `json:"scan"`
	IssueCount             *int             `json:"issueCount"`
	Issues                 []ResourcedIssue `json:"issues"`
	LastSuccessfulScanTime *metav1.Time     `json:"lastSuccessfulScanTime"`
	LastFinishedScanTime   *metav1.Time     `json:"lastFinishedScanTime"`
	NextScheduleScanTime   *metav1.Time     `json:"nextScheduleScanTime"`
}

// +k8s:deepcopy-gen=true
type NsName struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// +k8s:deepcopy-gen=true
type ResourcedIssue struct {
	Issue     `json:",inline"`
	Resources map[string][]NsName `json:"resources"`
}

// +k8s:deepcopy-gen=true
type Resources struct {
	Discovered bool      `json:"discovered"`
	Message    string    `json:"message"`
	Memory     *Resource `json:"memory"`
	CPU        *Resource `json:"cpu"`
}

// +k8s:deepcopy-gen=true
type Resource struct {
	Available       string `json:"available"`
	Usage           string `json:"usage"`
	UsagePercentage int32  `json:"usagePercentage"`
}

// +k8s:deepcopy-gen=true
type ScanStatus struct {
	Status  ScanStatusType `json:"status"`
	Message string         `json:"message"`
	Suspend bool           `json:"suspend"`
}

// +k8s:deepcopy-gen=true
type ConnectionStatus struct {
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
}

func NewCluster(cluster v1alpha1.Cluster, scans []v1alpha1.ClusterScan) Cluster {
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

	for _, cs := range scans {
		if cs.Status.TotalIssues != nil {
			if cl.TotalIssues == nil {
				cl.TotalIssues = new(int)
			}
			*cl.TotalIssues += *cs.Status.TotalIssues
		}
		for p, s := range cs.Status.Plugins {
			if cl.PluginStatus[p] == nil {
				if cl.PluginStatus == nil {
					cl.PluginStatus = map[string]*PluginStatus{}
				}
				cl.PluginStatus[p] = &PluginStatus{
					Scan: &ScanStatus{
						Status: Unknown,
					},
				}
			}
			cl.PluginStatus[p].Scan.Suspend = s.Suspend

			if s.IssueCount != nil {
				if cl.PluginStatus[p].IssueCount == nil {
					cl.PluginStatus[p].IssueCount = new(int)
				}
				*cl.PluginStatus[p].IssueCount += *s.IssueCount
			}

			switch s.LastFinishedStatus {
			case string(batchv1.JobComplete):
				cl.PluginStatus[p].Scan.Status = Scanned
			case string(batchv1.JobFailed):
				cl.PluginStatus[p].Scan.Status = Failed
				cl.PluginStatus[p].Scan.Message = s.LastErrorMsg
			case "":
				cl.PluginStatus[p].Scan.Message = "Scan not finished"
			}

			if cl.PluginStatus[p].LastSuccessfulScanTime == nil ||
				cl.PluginStatus[p].LastSuccessfulScanTime.Before(cs.Status.LastSuccessfulTime) {
				cl.PluginStatus[p].LastSuccessfulScanTime = cs.Status.LastSuccessfulTime
			}
			if cl.PluginStatus[p].LastFinishedScanTime == nil ||
				cl.PluginStatus[p].LastFinishedScanTime.Before(cs.Status.LastFinishedTime) {
				cl.PluginStatus[p].LastFinishedScanTime = cs.Status.LastFinishedTime
			}
			if cl.PluginStatus[p].NextScheduleScanTime == nil ||
				cs.Status.NextScheduleTime.Before(cl.PluginStatus[p].NextScheduleScanTime) {
				cl.PluginStatus[p].NextScheduleScanTime = cs.Status.NextScheduleTime
			}

		}
	}

	return cl
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

func NewClusterWithIssues(cluster v1alpha1.Cluster, scans []v1alpha1.ClusterScan, issues []v1alpha1.ClusterIssue) Cluster {
	c := NewCluster(cluster, scans)
	for _, i := range issues {
		c.PluginStatus[i.Labels[v1alpha1.LabelPlugin]].Issues = append(
			c.PluginStatus[i.Labels[v1alpha1.LabelPlugin]].Issues,
			NewResourcedIssue(i),
		)
	}
	return c
}

func NewClusterSlice(carr []v1alpha1.Cluster, csarr []v1alpha1.ClusterScan) []Cluster {
	scanm := map[string][]v1alpha1.ClusterScan{}
	clusters := []Cluster{}

	for _, cs := range csarr {
		nn := fmt.Sprintf("%s/%s", cs.Namespace, cs.Spec.ClusterRef.Name)
		scanm[nn] = append(scanm[nn], cs)
	}
	for _, c := range carr {
		clusters = append(clusters, NewCluster(
			c,
			scanm[fmt.Sprintf("%s/%s", c.Namespace, c.Name)],
		))
	}
	return clusters
}

func (r Cluster) Reader() (io.Reader, error) {
	jc, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(jc), nil
}
