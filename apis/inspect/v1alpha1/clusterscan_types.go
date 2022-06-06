package v1alpha1

import (
	"strings"

	"github.com/getupio-undistro/inspect/pkg/apis"
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
	apis.Status `json:",inline"`

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
	TotalIssues int `json:"totalIssues"`
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

// SyncStatus fills PluginNames, NextScheduleTime, LastScheduleTime and LastSuccessfulTime fields based on PluginStatus
func (in *ClusterScanStatus) SyncStatus() {
	var names []string
	in.NextScheduleTime = nil
	for n, s := range in.Plugins {
		names = append(names, n)
		if in.LastScheduleTime == nil {
			in.LastScheduleTime = s.LastScheduleTime
			in.LastStatus = s.LastStatus
		}
		if in.LastFinishedTime == nil {
			in.LastFinishedTime = s.LastFinishedTime
			in.LastStatus = s.LastStatus
			in.LastFinishedStatus = s.LastStatus
		}
		if in.LastSuccessfulTime == nil {
			in.LastSuccessfulTime = s.LastSuccessfulTime
		}
		if in.NextScheduleTime == nil {
			in.NextScheduleTime = s.NextScheduleTime
		}
		if s.LastScheduleTime != nil && s.LastScheduleTime.After(in.LastScheduleTime.Time) {
			in.LastScheduleTime = s.LastScheduleTime
			in.LastStatus = s.LastStatus
		}
		if s.LastFinishedTime != nil && s.LastFinishedTime.After(in.LastScheduleTime.Time) {
			in.LastFinishedTime = s.LastFinishedTime
			in.LastStatus = s.LastStatus
			in.LastFinishedStatus = s.LastStatus
		}
		if s.LastSuccessfulTime != nil && s.LastSuccessfulTime.After(in.LastSuccessfulTime.Time) {
			in.LastSuccessfulTime = s.LastSuccessfulTime
		}
		if s.NextScheduleTime != nil && s.NextScheduleTime.Before(in.NextScheduleTime) {
			in.NextScheduleTime = s.NextScheduleTime
		}
	}
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

// +k8s:deepcopy-gen=true
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
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterRef.name",priority=0
//+kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule",priority=0
//+kubebuilder:printcolumn:name="Suspend",type="boolean",JSONPath=".status.suspend",priority=0
//+kubebuilder:printcolumn:name="Plugins",type="string",JSONPath=".status.pluginNames",priority=0
//+kubebuilder:printcolumn:name="Last Status",type="string",JSONPath=".status.lastScanStatus",priority=0
//+kubebuilder:printcolumn:name="Last Schedule",type="date",JSONPath=".status.lastScheduleTime",priority=0
//+kubebuilder:printcolumn:name="Last Successful",type="date",JSONPath=".status.lastSuccessfulTime",priority=0
//+kubebuilder:printcolumn:name="Issues",type="integer",JSONPath=".status.totalIssues",priority=0
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",priority=0
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0
//+kubebuilder:printcolumn:name="Next Schedule",type="string",JSONPath=".status.nextScheduleTime",priority=1

// ClusterScan is the Schema for the clusterscans API
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
