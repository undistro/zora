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

package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/undistro/zora/apis/zora/v1alpha1"
	"github.com/undistro/zora/payloads"
	"github.com/undistro/zora/pkg/clientset/versioned"
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

		issues := payloads.NewIssues(issueList.Items)
		log.Info(fmt.Sprintf("%d issue(s) returned", len(issues)))
		RespondWithJSON(w, http.StatusOK, issues)
	}
}
