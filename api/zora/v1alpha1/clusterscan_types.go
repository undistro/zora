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
	"sort"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ClusterScanSpec defines the desired state of ClusterScan
type ClusterScanSpec struct {
	// ClusterRef is a reference to a Cluster in the same namespace
	ClusterRef corev1.LocalObjectReference `json:"clusterRef"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	Suspend *bool `json:"suspend,omitempty"`

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// The list of Plugin references that are used to scan the referenced Cluster.  Defaults to 'popeye'
	Plugins []PluginReference `json:"plugins,omitempty"`

	// SuccessfulScansHistoryLimit specifies the amount of successfully
	// completed scan Jobs to be kept in the cluster. This field is analogous
	// to <Cronjob.Spec.SuccessfulJobsHistoryLimit> from the <batch> package.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=3
	SuccessfulScansHistoryLimit *int32 `json:"successfulScansHistoryLimit,omitempty"`

	// FailedScansHistoryLimit specifies the amount of failed scan Jobs to be
	// kept in the cluster. This field is analogous to
	// <Cronjob.Spec.FailedScansHistoryLimit> from the <batch> package.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	FailedScansHistoryLimit *int32 `json:"failedScansHistoryLimit,omitempty"`
}

type PluginReference struct {
	// Name is unique within a namespace to reference a Plugin resource.
	Name string `json:"name"`

	// Namespace defines the space within which the Plugin name must be unique.
	Namespace string `json:"namespace,omitempty"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	Suspend *bool `json:"suspend,omitempty"`

	// The schedule in Cron format for this Plugin, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule,omitempty"`

	// List of environment variables to set in the Plugin container.
	Env []corev1.EnvVar `json:"env,omitempty"`
}

func (in *PluginReference) PluginKey(defaultNamespace string) types.NamespacedName {
	ns := in.Namespace
	if ns == "" {
		ns = defaultNamespace
	}
	return types.NamespacedName{Name: in.Name, Namespace: ns}
}

// ClusterScanStatus defines the observed state of ClusterScan
type ClusterScanStatus struct {
	Status `json:",inline"`

	// Information of the last scans of plugins
	Plugins map[string]*PluginScanStatus `json:"plugins,omitempty"`

	// Comma separated list of plugins
	PluginNames string `json:"pluginNames,omitempty"`

	// Suspend field value from ClusterScan spec
	Suspend bool `json:"suspend"`

	// Information when was the last time the job was scheduled.
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Information when was the last time the job was finished.
	LastFinishedTime *metav1.Time `json:"lastFinishedTime,omitempty"`

	// Status of the last finished scan. Complete or Failed
	LastFinishedStatus string `json:"lastFinishedStatus,omitempty"`

	// Status of the last scan. Active, Complete or Failed
	LastStatus string `json:"lastStatus,omitempty"`

	// Information when was the last time the job successfully completed.
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`

	// Time when the next job will schedule.
	NextScheduleTime *metav1.Time `json:"nextScheduleTime,omitempty"`

	// Total of ClusterIssues reported in the last successful scan
	TotalIssues *int `json:"totalIssues,omitempty"`
}

// GetPluginStatus returns a PluginScanStatus of a plugin
func (in *ClusterScanStatus) GetPluginStatus(name string) *PluginScanStatus {
	if in.Plugins == nil {
		in.Plugins = make(map[string]*PluginScanStatus)
	}
	if _, ok := in.Plugins[name]; !ok {
		in.Plugins[name] = &PluginScanStatus{}
	}
	return in.Plugins[name]
}

// SyncStatus updates ClusterScan status and time fields based on PluginStatus
func (in *ClusterScanStatus) SyncStatus() {
	var names, failed, active, complete []string
	var sechedule, finishedTime, successful, next *metav1.Time
	for name, plugin := range in.Plugins {
		names = append(names, name)
		if sechedule == nil || sechedule.Before(plugin.LastScheduleTime) {
			sechedule = plugin.LastScheduleTime
		}
		if finishedTime == nil || finishedTime.Before(plugin.LastFinishedTime) {
			finishedTime = plugin.LastFinishedTime
		}
		if successful == nil || successful.Before(plugin.LastSuccessfulTime) {
			successful = plugin.LastSuccessfulTime
		}
		if next == nil || plugin.NextScheduleTime.Before(next) {
			next = plugin.NextScheduleTime
		}
		if plugin.LastStatus == "Active" {
			active = append(active, name)
		}
		switch plugin.LastFinishedStatus {
		case string(batchv1.JobFailed):
			failed = append(failed, name)
		case string(batchv1.JobComplete):
			complete = append(complete, name)
		}
	}
	var finishedStatus, status string

	if len(failed) > 0 {
		finishedStatus = string(batchv1.JobFailed)
		status = string(batchv1.JobFailed)
	}
	if len(failed) == 0 && len(complete) > 0 {
		finishedStatus = string(batchv1.JobComplete)
		status = string(batchv1.JobComplete)
	}
	if len(active) > 0 {
		status = "Active"
	}

	in.LastScheduleTime = sechedule
	in.LastFinishedTime = finishedTime
	in.LastSuccessfulTime = successful
	in.NextScheduleTime = next
	in.LastFinishedStatus = finishedStatus
	in.LastStatus = status
	sort.Strings(names)
	in.PluginNames = strings.Join(names, ",")
}

// LastScanIDs returns a list of all the last scan IDs
func (in *ClusterScanStatus) LastScanIDs(successful bool) []string {
	lastScans := make([]string, 0, len(in.Plugins))
	for _, ps := range in.Plugins {
		sid := ps.LastScanID
		if successful {
			sid = ps.LastSuccessfulScanID
		}
		if sid != "" {
			lastScans = append(lastScans, sid)
		}
	}
	return lastScans
}

type PluginScanStatus struct {
	// Information when was the last time the job was scheduled.
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Information when was the last time the job was finished.
	LastFinishedTime *metav1.Time `json:"lastFinishedTime,omitempty"`

	// Information when was the last time the job successfully completed.
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`

	// Time when the next job will schedule.
	NextScheduleTime *metav1.Time `json:"nextScheduleTime,omitempty"`

	// ID of the last plugin scan
	LastScanID string `json:"lastScanID,omitempty"`

	// ID of the last successful plugin scan
	LastSuccessfulScanID string `json:"lastSuccessfulScanID,omitempty"`

	// Status of the last plugin scan. Active, Complete or Failed
	LastStatus string `json:"lastStatus,omitempty"`

	// Status of the last finished plugin scan. Complete or Failed
	LastFinishedStatus string `json:"lastFinishedStatus,omitempty"`

	// LastErrorMsg contains a plugin error message from the last failed scan.
	LastErrorMsg string `json:"lastErrorMsg,omitempty"`

	// IssueCount holds the sum of ClusterIssues found in the last successful scan.
	IssueCount *int `json:"issueCount,omitempty"`

	// Suspend field value from ClusterScan spec.plugins.*.suspend
	Suspend bool `json:"suspend,omitempty"`

	// The Cron schedule of this plugin
	Schedule string `json:"schedule,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName={scan,scans}
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterRef.name",priority=0
//+kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule",priority=0
//+kubebuilder:printcolumn:name="Suspend",type="boolean",JSONPath=".status.suspend",priority=0
//+kubebuilder:printcolumn:name="Plugins",type="string",JSONPath=".status.pluginNames",priority=0
//+kubebuilder:printcolumn:name="Last Status",type="string",JSONPath=".status.lastStatus",priority=0
//+kubebuilder:printcolumn:name="Last Schedule",type="date",JSONPath=".status.lastScheduleTime",priority=0
//+kubebuilder:printcolumn:name="Last Successful",type="date",JSONPath=".status.lastSuccessfulTime",priority=0
//+kubebuilder:printcolumn:name="Issues",type="integer",JSONPath=".status.totalIssues",priority=0
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",priority=0
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0
//+kubebuilder:printcolumn:name="SaaS",type="string",JSONPath=".status.conditions[?(@.type==\"SaaS\")].reason",priority=1
//+kubebuilder:printcolumn:name="Next Schedule",type="string",JSONPath=".status.nextScheduleTime",priority=1

// ClusterScan is the Schema for the clusterscans API
// +genclient
// +genclient:onlyVerbs=list,get
// +genclient:noStatus
type ClusterScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterScanSpec   `json:"spec,omitempty"`
	Status ClusterScanStatus `json:"status,omitempty"`
}

func (in *ClusterScan) SetReadyStatus(status bool, reason, msg string) {
	s := metav1.ConditionFalse
	if status {
		s = metav1.ConditionTrue
	}
	in.Status.SetCondition(metav1.Condition{
		Type:               "Ready",
		Status:             s,
		ObservedGeneration: in.Generation,
		Reason:             reason,
		Message:            msg,
	})
}

func (in *ClusterScan) SetSaaSStatus(status metav1.ConditionStatus, reason, msg string) {
	in.Status.SetCondition(metav1.Condition{
		Type:               "SaaS",
		Status:             status,
		ObservedGeneration: in.Generation,
		Reason:             reason,
		Message:            msg,
	})
}

func (in *ClusterScan) ClusterKey() types.NamespacedName {
	return types.NamespacedName{Name: in.Spec.ClusterRef.Name, Namespace: in.Namespace}
}

//+kubebuilder:object:root=true

// ClusterScanList contains a list of ClusterScan
type ClusterScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterScan{}, &ClusterScanList{})
}
