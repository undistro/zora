// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	scheme "github.com/getupio-undistro/zora/pkg/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rest "k8s.io/client-go/rest"
)

// ClusterScansGetter has a method to return a ClusterScanInterface.
// A group's client should implement this interface.
type ClusterScansGetter interface {
	ClusterScans(namespace string) ClusterScanInterface
}

// ClusterScanInterface has methods to work with ClusterScan resources.
type ClusterScanInterface interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ClusterScan, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ClusterScanList, error)
	ClusterScanExpansion
}

// clusterScans implements ClusterScanInterface
type clusterScans struct {
	client rest.Interface
	ns     string
}

// newClusterScans returns a ClusterScans
func newClusterScans(c *ZoraV1alpha1Client, namespace string) *clusterScans {
	return &clusterScans{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the clusterScan, and returns the corresponding clusterScan object, and an error if there is any.
func (c *clusterScans) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ClusterScan, err error) {
	result = &v1alpha1.ClusterScan{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("clusterscans").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ClusterScans that match those selectors.
func (c *clusterScans) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ClusterScanList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ClusterScanList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("clusterscans").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}
