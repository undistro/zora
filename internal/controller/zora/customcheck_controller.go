// Copyright 2023 Undistro Authors
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
	"time"

	marvin "github.com/undistro/marvin/pkg/types"
	"github.com/undistro/marvin/pkg/validator"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"

	zorav1alpha1 "github.com/undistro/zora/api/zora/v1alpha1"
)

const (
	customChecksFinalizer = "customcheck.zora.undistro.io/finalizer"
)

// CustomCheckReconciler reconciles a CustomCheck object
type CustomCheckReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	ConfigMapName      string
	ConfigMapNamespace string
}

//+kubebuilder:rbac:groups=zora.undistro.io,resources=customchecks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=zora.undistro.io,resources=customchecks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=zora.undistro.io,resources=customchecks/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CustomCheckReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	check := &zorav1alpha1.CustomCheck{}
	if err := r.Get(ctx, req.NamespacedName, check); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("resourceVersion", check.ResourceVersion)
	defer func(t time.Time) {
		log.Info(fmt.Sprintf("CustomCheck has been reconciled in %v", time.Since(t)))
	}(time.Now())

	if check.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(check, customChecksFinalizer) {
			controllerutil.AddFinalizer(check, customChecksFinalizer)
			if err := r.Update(ctx, check); err != nil {
				log.Error(err, "failed to add finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		log.Info("the CustomCheck is being deleted")
		if controllerutil.ContainsFinalizer(check, customChecksFinalizer) {
			if err := r.onDelete(ctx, check); err != nil {
				log.Error(err, "error deleting CustomCheck")
			}
			controllerutil.RemoveFinalizer(check, customChecksFinalizer)
			if err := r.Update(ctx, check); err != nil {
				log.Error(err, "failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	err := r.apply(ctx, check)

	if err := r.Status().Update(ctx, check); err != nil {
		if apierrors.IsConflict(err) {
			log.Info("requeue after 1s due to conflict on update CustomCheck status", "error", err.Error())
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		log.Error(err, "failed to update CustomCheck status")
	}

	return ctrl.Result{}, err
}

func (r *CustomCheckReconciler) apply(ctx context.Context, check *zorav1alpha1.CustomCheck) error {
	log := ctrllog.FromContext(ctx)

	compiled := check.ToMarvin()
	if _, err := validator.Compile(*compiled, nil, nil); err != nil {
		log.Error(err, "failed to compile CustomCheck")
		check.SetReadyStatus(false, "CompileError", err.Error())
		return nil
	}

	cm := r.newConfigMap()
	result, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, r.mutateConfigMap(cm, check, compiled))
	if err != nil {
		log.Error(err, "failed to apply ConfigMap")
		check.SetReadyStatus(false, "ConfigMapApplyError", err.Error())
		return err
	}
	if result != controllerutil.OperationResultNone {
		log.Info(fmt.Sprintf("ConfigMap has been %s", result))
	}
	check.SetReadyStatus(true, "CustomCheckReconciled", "custom check successfully configured")
	return nil
}

func (r *CustomCheckReconciler) onDelete(ctx context.Context, check *zorav1alpha1.CustomCheck) error {
	cm := r.newConfigMap()
	key := types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name}
	if err := r.Get(ctx, key, cm); err != nil {
		return fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	name := check.FileName()

	if cm.Data == nil {
		return nil
	}
	if _, ok := cm.Data[name]; !ok {
		return nil
	}
	delete(cm.Data, name)
	if err := r.Update(ctx, cm); err != nil {
		return fmt.Errorf("failed to remove check from ConfigMap: %w", err)
	}
	return nil
}

func (r *CustomCheckReconciler) newConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: r.ConfigMapName, Namespace: r.ConfigMapNamespace}}
}

func (r *CustomCheckReconciler) mutateConfigMap(existing *corev1.ConfigMap, check *zorav1alpha1.CustomCheck, compiled *marvin.Check) controllerutil.MutateFn {
	return func() error {
		if existing.Labels == nil {
			existing.Labels = make(map[string]string)
		}
		existing.Labels["app.kubernetes.io/managed-by"] = "zora"
		if existing.Data == nil {
			existing.Data = make(map[string]string)
		}
		b, err := yaml.Marshal(compiled)
		if err != nil {
			return err
		}
		existing.Data[check.FileName()] = string(b)
		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomCheckReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&zorav1alpha1.CustomCheck{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
