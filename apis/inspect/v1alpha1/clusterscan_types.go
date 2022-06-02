package v1alpha1

import (
	"strings"
	"time"

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

	// Last scan ID, schedule and successful time of plugins
	PluginStatus map[string]*PluginCronJobStatus `json:"pluginStatus,omitempty"`

	// Comma separated list of plugins
	PluginNames string `json:"pluginNames,omitempty"`

	// Suspend field value from ClusterScan spec
	Suspend bool `json:"suspend"`

	// Information when was the last time the job was successfully scheduled.
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Information when was the last time the job successfully completed.
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`

	// Time when the next job will schedule.
	NextScheduleTime       *metav1.Time `json:"nextScheduleTime,omitempty"`
	NextScheduleTimeString string       `json:"nextScheduleTimeString,omitempty"`

	// Total of ClusterIssues reported by plugins
	TotalIssues int `json:"totalIssues"`
}

// SyncStatus fills PluginNames, NextScheduleTime, LastScheduleTime and LastSuccessfulTime fields based on Plugins status
func (in *ClusterScanStatus) SyncStatus() {
	var names []string
	in.NextScheduleTime = nil
	for n, s := range in.PluginStatus {
		names = append(names, n)
		if in.LastScheduleTime == nil {
			in.LastScheduleTime = s.LastScheduleTime
		}
		if in.LastSuccessfulTime == nil {
			in.LastSuccessfulTime = s.LastSuccessfulTime
		}
		if in.NextScheduleTime == nil {
			in.NextScheduleTime = s.NextScheduleTime
			in.NextScheduleTimeString = s.NextScheduleTime.Format(time.RFC3339)
		}
		if s.LastScheduleTime != nil && s.LastScheduleTime.After(in.LastScheduleTime.Time) {
			in.LastScheduleTime = s.LastScheduleTime
		}
		if s.LastSuccessfulTime != nil && s.LastSuccessfulTime.After(in.LastSuccessfulTime.Time) {
			in.LastSuccessfulTime = s.LastSuccessfulTime
		}
		if s.NextScheduleTime != nil && s.NextScheduleTime.Before(in.NextScheduleTime) {
			in.NextScheduleTime = s.NextScheduleTime
			in.NextScheduleTimeString = s.NextScheduleTime.Format(time.RFC3339)
		}
	}
	in.PluginNames = strings.Join(names, ",")
}

// LastScanIDs returns a list of all the last scan IDs
func (in *ClusterScanStatus) LastScanIDs() []string {
	lastScans := make([]string, 0, len(in.PluginStatus))
	for _, ps := range in.PluginStatus {
		if ps.LastScanID != "" {
			lastScans = append(lastScans, ps.LastScanID)
		}
	}
	return lastScans
}

// +k8s:deepcopy-gen=true
type PluginCronJobStatus struct {
	// Information when was the last time the job was successfully scheduled.
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Information when was the last time the job successfully completed.
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`

	// Time when the next job will schedule.
	NextScheduleTime *metav1.Time `json:"nextScheduleTime,omitempty"`

	// ID of the last plugin scan
	LastScanID string `json:"scanID,omitempty"`

	// Indicates whether this plugin is currently running
	Active *bool `json:"active,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterRef.name",priority=0
//+kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule",priority=0
//+kubebuilder:printcolumn:name="Suspend",type="boolean",JSONPath=".status.suspend",priority=0
//+kubebuilder:printcolumn:name="Plugins",type="string",JSONPath=".status.pluginNames",priority=0
//+kubebuilder:printcolumn:name="Last Schedule",type="date",JSONPath=".status.lastScheduleTime",priority=0
//+kubebuilder:printcolumn:name="Last Successful",type="date",JSONPath=".status.lastSuccessfulTime",priority=0
//+kubebuilder:printcolumn:name="Issues",type="integer",JSONPath=".status.totalIssues",priority=0
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",priority=0
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0
//+kubebuilder:printcolumn:name="Next Schedule",type="string",JSONPath=".status.nextScheduleTimeString",priority=1

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
