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
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"github.com/getupio-undistro/zora/pkg/discovery"
	"github.com/getupio-undistro/zora/pkg/kubeconfig"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Config   *rest.Config
}

//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	cluster := &v1alpha1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("resourceVersion", cluster.ResourceVersion)
	ctx = ctrllog.IntoContext(ctx, log)

	defer func(t time.Time) {
		log.Info(fmt.Sprintf("Cluster has been reconciled in %v", time.Since(t)))
	}(time.Now())

	err := r.reconcile(ctx, cluster)
	if err := r.Status().Update(ctx, cluster); err != nil {
		log.Error(err, "failed to update Cluster status")
	}
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, err
}

func (r *ClusterReconciler) reconcile(ctx context.Context, cluster *v1alpha1.Cluster) error {
	log := ctrllog.FromContext(ctx)

	config := r.Config
	if cluster.Spec.KubeconfigRef != nil {
		key := cluster.KubeconfigRefKey()
		clusterConfig, err := kubeconfig.ConfigFromSecretName(ctx, r.Client, *key)
		if err != nil {
			log.Error(err, "failed to get config from kubeconfigRef")
			r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterReady, false, "KubeconfigError", err.Error())
			return err
		}
		config = clusterConfig
	}

	discoverer, err := discovery.NewForConfig(config)
	if err != nil {
		log.Error(err, "failed to get new discovery client from config")
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterReady, false, "ClusterNotConnected", err.Error())
		return err
	}

	version, err := discoverer.Version()
	if err != nil {
		log.Error(err, "failed to discover cluster version")
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterReady, false, "ClusterVersionError", err.Error())
		return err
	}
	cluster.Status.KubernetesVersion = version
	cluster.SetStatus(v1alpha1.ClusterReady, true, "ClusterConnected", fmt.Sprintf("cluster successfully connected, version %s", version))

	if info, err := discoverer.Info(ctx); err != nil {
		log.Error(err, "failed to discovery cluster info")
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterDiscovered, false, "ClusterInfoNotDiscovered", err.Error())
	} else {
		cluster.Status.ClusterInfo = *info
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterDiscovered, true, "ClusterInfoDiscovered", "cluster info successfully discovered")
	}

	if res, err := discoverer.Resources(ctx); err != nil {
		log.Error(err, "failed to discovery cluster resources")
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterResourcesDiscovered, false, "ClusterResourcesNotDiscovered", err.Error())
	} else {
		cluster.Status.SetResources(res)
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterResourcesDiscovered, true, "ClusterResourcesDiscovered", "cluster resources successfully discovered")
	}

	cluster.Status.LastReconciliationTime = metav1.Now()
	cluster.Status.ObservedGeneration = cluster.Generation
	return nil
}

// setStatusAndCreateEvent sets Cluster status and creates an event if the given status type is updated
func (r *ClusterReconciler) setStatusAndCreateEvent(cluster *v1alpha1.Cluster, statusType string, status bool, reason string, msg string) {
	before := cluster.Status.GetCondition(statusType).DeepCopy()
	cluster.SetStatus(statusType, status, reason, msg)
	if reflect.DeepEqual(before, cluster.Status.GetCondition(statusType)) {
		return
	}
	eventtype := corev1.EventTypeWarning
	if status {
		eventtype = corev1.EventTypeNormal
	}
	r.Recorder.Event(cluster, eventtype, reason, msg)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Cluster{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
