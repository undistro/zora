package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/getupio-undistro/zora/payloads"
	"github.com/getupio-undistro/zora/pkg/clientset/versioned"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IssueListHandler(client versioned.Interface, logger logr.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.WithName("handlers.issues").WithValues("method", r.Method, "path", r.URL.Path)

		scanList, err := client.ZoraV1alpha1().ClusterScans("").List(r.Context(), metav1.ListOptions{})
		if err != nil {
			log.Error(err, "failed to list ClusterScans")
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error listing ClusterScans", err.Error())
			return
		}
		var lastScanIDs []string
		var ls string
		for _, cs := range scanList.Items {
			lastScanIDs = append(lastScanIDs, cs.Status.LastScanIDs(true)...)
		}
		if len(lastScanIDs) > 0 {
			ls = fmt.Sprintf("%s in (%s)", v1alpha1.LabelScanID, strings.Join(lastScanIDs, ","))
		}
		issueList, err := client.ZoraV1alpha1().ClusterIssues("").List(r.Context(), metav1.ListOptions{LabelSelector: ls})
		if err != nil {
			log.Error(err, "failed to list ClusterIssues")
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error listing ClusterIssues", err.Error())
			return
		}

		cslist, err := client.ZoraV1alpha1().ClusterScans("").List(r.Context(), metav1.ListOptions{})
		if err != nil {
			log.Error(err, "Failed to list ClusterScans")
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error listing ClusterScans", err.Error())
			return
		}

		failedscan := map[string]struct{}{}
		for _, cs := range cslist.Items {
			if cs.Status.LastFinishedStatus == string(batchv1.JobFailed) {
				failedscan[cs.Spec.ClusterRef.Name] = struct{}{}
			}
		}

		issues := payloads.NewIssues(issueList.Items, failedscan)
		log.Info(fmt.Sprintf("%d issue(s) returned", len(issues)))
		RespondWithJSON(w, http.StatusOK, issues)
	}
}
