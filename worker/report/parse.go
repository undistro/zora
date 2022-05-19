package report

import (
	"fmt"
	"io"

	inspectv1a1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/worker/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Parse receives a reader pointing to a plugin's report file, transforming such
// report into an instance of <ClusterIssueList> according to the cluster name
// and issues namespace specified on the <Config> struct. The parsing for each
// plugin is left to dedicated functions which are called according to the
// plugin type.
func Parse(r io.Reader, c *config.Config) (*inspectv1a1.ClusterIssueList, error) {
	repby := []byte{}
	if _, err := r.Read(repby); err != nil {
		return nil, fmt.Errorf("Unable to read results of plugin <%s> from cluster <%s>: %w", c.Plugin, c.Cluster, err)
	}
	cispecs, err := config.PluginParsers[c.Plugin](repby)
	if err != nil {
		return err
	}

	cilist := *inspectv1a1.ClusterIssueList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterIssueList",
			APIVersion: inspectv1a1.SchemeGroupVersion.String(),
		},
	}
	for _, cis := range cispecs {
		cis.Cluster = c.Cluster
		cilist.Items = append(cilist.Items, inspectv1a1.ClusterIssue{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterIssue",
				APIVersion: cilist.APIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.ClusterIssuesNs,
			},
			Spec: *cis,
		})
	}
	return cilist, nil
}
