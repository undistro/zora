package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/getupio-undistro/zora/payloads"
	"github.com/getupio-undistro/zora/pkg/clientset/versioned"
	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ClusterHandler(client versioned.Interface, logger logr.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.WithName("handlers.cluster").WithValues("method", r.Method, "path", r.URL.Path)

		namespace := chi.URLParam(r, "namespace")
		clusterName := chi.URLParam(r, "clusterName")

		cluster, err := client.ZoraV1alpha1().Clusters(namespace).Get(r.Context(), clusterName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				RespondWithCode(w, http.StatusNotFound)
				return
			}
			log.Error(err, "failed to get Cluster")
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error getting Cluster", err.Error())
			return
		}

		ls := fmt.Sprintf("%s=%s", v1alpha1.LabelCluster, clusterName)
		if len(cluster.Status.LastScans) > 0 {
			ls = fmt.Sprintf("%s,%s in (%s)", ls, v1alpha1.LabelScanID, strings.Join(cluster.Status.LastScans, ","))
		}
		issueList, err := client.ZoraV1alpha1().ClusterIssues(namespace).List(r.Context(), metav1.ListOptions{LabelSelector: ls})
		if err != nil {
			log.Error(err, fmt.Sprintf("failed to list ClusterIssues by label selector %s", ls))
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error listing ClusterIssues", err.Error())
			return
		}

		log.Info(fmt.Sprintf("cluster %s returned with %d issues", clusterName, len(issueList.Items)))
		RespondWithJSON(w, http.StatusOK, payloads.NewClusterWithIssues(*cluster, issueList.Items))
	}
}
