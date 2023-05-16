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

// +kubebuilder:object:generate=true
package discovery

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MeasuredResources = []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory}

// +kubebuilder:object:generate=false
type ClusterDiscoverer interface {
	Info(context.Context) (*ClusterInfo, error)
	Resources(ctx context.Context) (ClusterResources, error)
	Version() (string, error)
}

type ClusterInfo struct {
	// total of Nodes
	TotalNodes *int `json:"totalNodes,omitempty"`

	// CreationTimestamp is a timestamp representing the server time when the kube-system namespace was created.
	// It is represented in RFC3339 form and is in UTC.
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty"`

	// Provider stores the cluster's source.
	Provider string `json:"provider,omitempty"`

	// Region holds the geographic location with most nodes.
	Region string `json:"region,omitempty"`
}

type NodeInfo struct {
	// Node name
	Name string `json:"name,omitempty"`

	// Node labels
	Labels map[string]string `json:"labels,omitempty"`

	// Usage and available resources
	Resources map[corev1.ResourceName]Resources `json:"resources,omitempty"`

	// True if node is in ready condition
	Ready bool `json:"ready,omitempty"`

	// CreationTimestamp is a timestamp representing the server time when this object was created.
	// It is represented in RFC3339 form and is in UTC.
	CreationTimestamp metav1.Time `json:"-"`
}

type Resources struct {
	// Quantity of resources available for scheduling
	Available resource.Quantity `json:"available,omitempty"`

	// Quantity of resources in use
	Usage resource.Quantity `json:"usage,omitempty"`

	// Percentage of resources in use
	UsagePercentage int32 `json:"usagePercentage,omitempty"`
}

type ClusterResources map[corev1.ResourceName]Resources
