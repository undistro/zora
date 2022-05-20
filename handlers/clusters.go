package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/getupio-undistro/inspect/payloads"
	"github.com/getupio-undistro/inspect/pkg/clientset/versioned"
)

func ClusterListHandler(client versioned.Interface, logger logr.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.WithName("handlers.clusters").WithValues("method", r.Method, "path", r.URL.Path)
		clusterList, err := client.InspectV1alpha1().Clusters("").List(r.Context(), metav1.ListOptions{})
		if err != nil {
			log.Error(err, "failed to list clusters")
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error listing Clusters", err.Error())
			return
		}
		clusters := make([]payloads.Cluster, 0, len(clusterList.Items))
		for _, c := range clusterList.Items {
			clusters = append(clusters, payloads.NewCluster(c))
		}
		log.Info(fmt.Sprintf("%d cluster(s) returned", len(clusters)))
		RespondWithJSON(w, http.StatusOK, clusters)
	}
}
