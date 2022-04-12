package discovery

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNodeResources(t *testing.T) {
	type args struct {
		nodes       []corev1.Node
		nodeMetrics []v1beta1.NodeMetrics
	}
	tests := []struct {
		name string
		args args
		want []NodeInfo
	}{
		{
			name: "OK",
			args: args{
				nodes: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node1",
							Labels: map[string]string{"foo": "bar"},
						},
						Status: corev1.NodeStatus{
							Allocatable: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("2000m"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Conditions: []corev1.NodeCondition{
								{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node2",
							Labels: map[string]string{"foo": "foo"},
						},
						Status: corev1.NodeStatus{
							Allocatable: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceCPU:    resource.MustParse("2000m"),
								corev1.ResourceMemory: resource.MustParse("4Gi"),
							},
							Conditions: []corev1.NodeCondition{
								{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
							},
						},
					},
				},
				nodeMetrics: []v1beta1.NodeMetrics{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node1",
							Labels: map[string]string{"foo": "bar"},
						},
						Usage: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "node2",
							Labels: map[string]string{"foo": "foo"},
						},
						Usage: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
				},
			},
			want: []NodeInfo{
				{
					Name:   "node1",
					Labels: map[string]string{"foo": "bar"},
					Ready:  true,
					Resources: map[corev1.ResourceName]Resources{
						corev1.ResourceCPU: {
							Available:       resource.MustParse("2000m"),
							Usage:           resource.MustParse("1000m"),
							UsagePercentage: 50,
						},
						corev1.ResourceMemory: {
							Available:       resource.MustParse("4Gi"),
							Usage:           resource.MustParse("2Gi"),
							UsagePercentage: 50,
						},
					},
				},
				{
					Name:   "node2",
					Labels: map[string]string{"foo": "foo"},
					Ready:  true,
					Resources: map[corev1.ResourceName]Resources{
						corev1.ResourceCPU: {
							Available:       resource.MustParse("2000m"),
							Usage:           resource.MustParse("1000m"),
							UsagePercentage: 50,
						},
						corev1.ResourceMemory: {
							Available:       resource.MustParse("4Gi"),
							Usage:           resource.MustParse("2Gi"),
							UsagePercentage: 50,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeResources(tt.args.nodes, tt.args.nodeMetrics); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAvgNodeResources(t *testing.T) {
	type args struct {
		nodes []NodeInfo
	}
	tests := []struct {
		name string
		args args
		want map[corev1.ResourceName]Resources
	}{
		{
			name: "OK",
			args: args{
				nodes: []NodeInfo{
					{
						Name:   "node1",
						Labels: map[string]string{"foo": "bar"},
						Ready:  true,
						Resources: map[corev1.ResourceName]Resources{
							corev1.ResourceCPU: {
								Available:       resource.MustParse("2000m"),
								Usage:           resource.MustParse("1000m"),
								UsagePercentage: 50,
							},
							corev1.ResourceMemory: {
								Available:       resource.MustParse("4Gi"),
								Usage:           resource.MustParse("2Gi"),
								UsagePercentage: 50,
							},
						},
					},
					{
						Name:   "node2",
						Labels: map[string]string{"foo": "foo"},
						Ready:  true,
						Resources: map[corev1.ResourceName]Resources{
							corev1.ResourceCPU: {
								Available:       resource.MustParse("2000m"),
								Usage:           resource.MustParse("1000m"),
								UsagePercentage: 50,
							},
							corev1.ResourceMemory: {
								Available:       resource.MustParse("4Gi"),
								Usage:           resource.MustParse("2Gi"),
								UsagePercentage: 50,
							},
						},
					},
				},
			},
			want: map[corev1.ResourceName]Resources{
				corev1.ResourceCPU: {
					Available:       resource.MustParse("4000m"),
					Usage:           resource.MustParse("2000m"),
					UsagePercentage: 50,
				},
				corev1.ResourceMemory: {
					Available:       resource.MustParse("8Gi"),
					Usage:           resource.MustParse("4Gi"),
					UsagePercentage: 50,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := avgNodeResources(tt.args.nodes); !resourcesAreEqual(got, tt.want) {
				t.Errorf("avgNodeResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func resourcesAreEqual(x, y map[corev1.ResourceName]Resources) bool {
	if x == nil || y == nil {
		return x == nil && y == nil
	}
	for name, res := range x {
		if !y[name].Available.Equal(res.Available) {
			return false
		}
		if !y[name].Usage.Equal(res.Usage) {
			return false
		}
		if y[name].UsagePercentage != res.UsagePercentage {
			return false
		}
	}

	return true
}
