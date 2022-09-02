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

	"github.com/getupio-undistro/zora/payloads"
	"github.com/getupio-undistro/zora/pkg/clientset/versioned"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ClusterListHandler(client versioned.Interface, logger logr.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.WithName("handlers.clusters").WithValues("method", r.Method, "path", r.URL.Path)
		clusterList, err := client.ZoraV1alpha1().Clusters("").List(r.Context(), metav1.ListOptions{})
		if err != nil {
			log.Error(err, "failed to list Clusters")
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error listing Clusters", err.Error())
			return
		}
		scanList, err := client.ZoraV1alpha1().ClusterScans("").List(r.Context(), metav1.ListOptions{})
		if err != nil {
			log.Error(err, "failed to list ClusterScans")
			RespondWithDetailedError(w, http.StatusInternalServerError, "Error listing ClusterScans", err.Error())
			return
		}

		clusters := payloads.NewClusterSlice(clusterList.Items, scanList.Items)
		log.Info(fmt.Sprintf("%d cluster(s) returned", len(clusters)))
		RespondWithJSON(w, http.StatusOK, clusters)
	}
}
