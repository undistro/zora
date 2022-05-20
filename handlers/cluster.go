package handlers

import (
	"fmt"
	"net/http"

	"github.com/getupio-undistro/inspect/payloads"
	"github.com/getupio-undistro/inspect/pkg/clientset/versioned"
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

		cluster, err := client.InspectV1alpha1().Clusters(namespace).Get(r.Context(), clusterName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				RespondWithCode(w, http.StatusNotFound)
				return
			}
			log.Error(err, "failed to get Cluster")
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error getting Cluster", err.Error())
			return
		}

		issueList, err := client.InspectV1alpha1().ClusterIssues(namespace).List(r.Context(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("cluster=%s", clusterName),
		})
		if err != nil {
			log.Error(err, fmt.Sprintf("failed to list ClusterIssues from cluster %s", clusterName))
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error listing ClusterIssues", err.Error())
			return
		}

		log.Info(fmt.Sprintf("cluster %s returned with %d issues", clusterName, len(issueList.Items)))
		RespondWithJSON(w, http.StatusOK, payloads.NewClusterWithIssues(*cluster, issueList.Items))
	}
}
