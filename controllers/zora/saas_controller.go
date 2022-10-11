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

package zora

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/undistro/zora/apis/zora/v1alpha1"
	payloads "github.com/undistro/zora/pkg/payloads/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// SaasReconciler reconciles a Saas object
type SaasReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	HttpCli    *http.Client
	ID         string
	ServerAddr string
}

//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterscans,verbs=list;watch
//+kubebuilder:rbac:groups=zora.undistro.io,resources=cluster,verbs=list
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterissues,verbs=list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SaasReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	if !r.validAddress() {
		log.Error(fmt.Errorf("Invalid url <%s>", r.ServerAddr), "No valid server address provided, nothing to do")
		return ctrl.Result{}, nil
	}

	clist := &v1alpha1.ClusterList{}
	if err := r.List(ctx, clist); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	cslist := &v1alpha1.ClusterScanList{}
	if err := r.List(ctx, cslist); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	cilist := &v1alpha1.ClusterIssueList{}
	if err := r.List(ctx, cilist); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	defer func(t time.Time) {
		log.Info(fmt.Sprintf("SaasReconciler has executed in %v", time.Since(t)))
	}(time.Now())

	if err := r.reconcile(ctx, clist, cslist, cilist); err != nil {
		return ctrl.Result{RequeueAfter: 20 * time.Minute}, err
	}
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

func (r *SaasReconciler) validAddress() bool {
	if r.ServerAddr[:4] != "http" {
		r.ServerAddr = fmt.Sprintf("http://%s", r.ServerAddr)
	}
	if u, err := url.ParseRequestURI(r.ServerAddr); err != nil || len(u.Host) == 0 {
		return false
	}
	return true
}

func (r *SaasReconciler) reconcile(ctx context.Context, clist *v1alpha1.ClusterList, cslist *v1alpha1.ClusterScanList, cilist *v1alpha1.ClusterIssueList) error {
	log := ctrllog.FromContext(ctx)

	sendf := func(bod io.Reader) {
		if err := r.send(bod); err != nil {
			log.Error(err, "Unable to complete transfer")
		}
		time.Sleep(100 * time.Millisecond)
	}

	issues := payloads.NewIssues(cilist.Items)
	log.Info(fmt.Sprintf("Sending %d issues", len(issues)))
	for _, bod := range issues {
		sendf(bod)
	}

	clusters := payloads.NewClusterSlice(clist.Items, cslist.Items)
	log.Info(fmt.Sprintf("Sending %d clusters", len(clusters)))
	for _, bod := range clusters {
		sendf(bod)
	}

	return nil
}

func (r *SaasReconciler) send(bod io.Reader) error {
	req, err := http.NewRequest("POST", r.ServerAddr, bod)
	if err != nil {
		return err
	}

	req.Header.Set("user-agent", fmt.Sprintf("Zora/%s SaasReconciler", v1alpha1.GroupVersion.Version))
	req.Header.Set("content-type", "application/json")
	req.URL.Query().Add("id", r.ID)

	res, err := r.HttpCli.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("Request returned status <%d %s>", res.StatusCode, res.Status)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SaasReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterScan{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
