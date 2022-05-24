package report

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	inspectv1a1 "github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/worker/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Parse receives a reader pointing to a plugin's report file, transforming
// such report into an array of <ClusterIssue> pointers according to the
// cluster name and issues namespace specified on the <Config> struct. The
// parsing for each plugin is left to dedicated functions which are called
// according to the plugin type.
func Parse(r io.Reader, c *config.Config) ([]*inspectv1a1.ClusterIssue, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("Invalid configuration: %w", err)
	}
	repby, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Unable to read results of plugin <%s> from cluster <%s>: %w", c.Plugin, c.Cluster, err)
	}
	cispecs, err := config.PluginParsers[c.Plugin](repby)
	if err != nil {
		return nil, err
	}

	ciarr := make([]*inspectv1a1.ClusterIssue, len(cispecs))
	for i := 0; i < len(cispecs); i++ {
		cispecs[i].Cluster = c.Cluster
		jid := c.Job[strings.LastIndex(c.Job, "-")+1:]
		ciarr[i] = &inspectv1a1.ClusterIssue{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterIssue",
				APIVersion: inspectv1a1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: c.ClusterIssuesNs,
				Name:      fmt.Sprintf("%s-%s-%s", c.Cluster, strings.ToLower(cispecs[i].ID), jid),
				Labels: map[string]string{
					inspectv1a1.LabelExecutionID: jid,
					inspectv1a1.LabelCluster:     c.Cluster,
				},
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "batch/v1",
					Kind:       "Job",
					Name:       c.Job,
					UID:        types.UID(c.JobUid),
				}},
			},
			Spec: *cispecs[i],
		}
	}
	return ciarr, nil
}
