// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// FakeClusters implements ClusterInterface
type FakeClusters struct {
	Fake *FakeInspectV1alpha1
	ns   string
}

var clustersResource = schema.GroupVersionResource{Group: "inspect", Version: "v1alpha1", Resource: "clusters"}

var clustersKind = schema.GroupVersionKind{Group: "inspect", Version: "v1alpha1", Kind: "Cluster"}

// Get takes name of the cluster, and returns the corresponding cluster object, and an error if there is any.
func (c *FakeClusters) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Cluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(clustersResource, c.ns, name), &v1alpha1.Cluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cluster), err
}

// List takes label and field selectors, and returns the list of Clusters that match those selectors.
func (c *FakeClusters) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(clustersResource, clustersKind, c.ns, opts), &v1alpha1.ClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ClusterList{ListMeta: obj.(*v1alpha1.ClusterList).ListMeta}
	for _, item := range obj.(*v1alpha1.ClusterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}