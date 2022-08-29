package zora

import (
	"context"
	"fmt"
	"time"

	"github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const Finalizer = "finalizer.undistro.io"

// ClusterIssueOverrideReconciler reconciles a ClusterIssueOverride object.
type ClusterIssueOverrideReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterissueoverrides,verbs=get;list;update;watch
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterissues,verbs=list;update
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterissues/status,verbs=update
//+kubebuilder:rbac:groups=batch,resources=clusterscans,verbs=get;watch

// Reconcile is part of the main Kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterIssueOverrideReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	ciolist := &v1alpha1.ClusterIssueOverrideList{}
	if err := r.List(ctx, ciolist); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	cilist := &v1alpha1.ClusterIssueList{}
	if err := r.List(ctx, cilist); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log = log.WithValues("resourceVersion", ciolist.ResourceVersion)
	ctx = ctrllog.IntoContext(ctx, log)

	defer func(t time.Time) {
		log.Info(fmt.Sprintf("ClusterIssueOverride has been reconciled in %v", time.Since(t)))
	}(time.Now())

	for _, cio := range ciolist.Items {
		if !cio.ObjectMeta.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(&cio, Finalizer) {
			log.Info("Instance under deletion")
			if err := r.reconcileDelete(ctx, &cio, cilist); err != nil {
				return ctrl.Result{}, err
			}
			log.Info(fmt.Sprintf("Removing Finalizer <%s>", Finalizer))
			controllerutil.RemoveFinalizer(&cio, Finalizer)
			if err := r.Update(ctx, &cio); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		if !controllerutil.ContainsFinalizer(&cio, Finalizer) {
			log.Info(fmt.Sprintf("Adding Finalizer <%s>", Finalizer))
			controllerutil.AddFinalizer(&cio, Finalizer)
			if err := r.Update(ctx, &cio); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if err := r.reconcile(ctx, ciolist, cilist); err != nil {
		return ctrl.Result{RequeueAfter: 20 * time.Minute}, err
	}
	return ctrl.Result{}, nil
}

// Mutate reflects overridden issue data on the issue itself, storing the
// original values on the issue's status.
func Mutate(ci *v1alpha1.ClusterIssue, cio *v1alpha1.ClusterIssueOverride) {
	ci.Status.Hidden = cio.Hidden()
	if ci.Status.Hidden {
		return
	}

	if ci.Status.OrigCategory == nil {
		ci.Status.OrigCategory = new(string)
		*ci.Status.OrigCategory = ci.Spec.Category
	}
	if ci.Status.OrigSeverity == nil {
		ci.Status.OrigSeverity = new(v1alpha1.ClusterIssueSeverity)
		*ci.Status.OrigSeverity = ci.Spec.Severity
	}
	if ci.Status.OrigMessage == nil {
		ci.Status.OrigMessage = new(string)
		*ci.Status.OrigMessage = ci.Spec.Message
	}

	if cio.Spec.Category != nil {
		ci.Spec.Category = *cio.Spec.Category
	} else if ci.Status.OrigCategory != nil {
		ci.Spec.Category = *ci.Status.OrigCategory
		ci.Status.OrigCategory = nil
	}
	if cio.Spec.Severity != nil {
		ci.Spec.Severity = *cio.Spec.Severity
	} else if ci.Status.OrigSeverity != nil {
		ci.Spec.Severity = *ci.Status.OrigSeverity
		ci.Status.OrigSeverity = nil
	}
	if cio.Spec.Message != nil {
		ci.Spec.Message = *cio.Spec.Message
	} else if ci.Status.OrigMessage != nil {
		ci.Spec.Message = *ci.Status.OrigMessage
		ci.Status.OrigMessage = nil
	}
}

// Outdated tells whether an issue differs from its override.
func Outdated(ci *v1alpha1.ClusterIssue, cio *v1alpha1.ClusterIssueOverride) bool {
	if cio.Hidden() != ci.Status.Hidden {
		return true
	}
	if (cio.Spec.Category != nil && ci.Spec.Category != *cio.Spec.Category) ||
		(cio.Spec.Category != ci.Status.OrigCategory) {
		return true
	}
	if (cio.Spec.Severity != nil && ci.Spec.Severity != *cio.Spec.Severity) ||
		(cio.Spec.Severity != ci.Status.OrigSeverity) {
		return true
	}
	if (cio.Spec.Message != nil && ci.Spec.Message != *cio.Spec.Message) ||
		(cio.Spec.Message != ci.Status.OrigMessage) {
		return true
	}

	return false
}

// This function holds implementation specific reconciliation flows.
func (r *ClusterIssueOverrideReconciler) reconcile(ctx context.Context, ciolist *v1alpha1.ClusterIssueOverrideList, cilist *v1alpha1.ClusterIssueList) error {
	ciom := map[string]v1alpha1.ClusterIssueOverride{}
	for _, cio := range ciolist.Items {
		ciom[cio.Name] = cio
	}

	for _, ci := range cilist.Items {
		if cio, overr := ciom[ci.Spec.ID]; overr && cio.InCluster(
			v1alpha1.NameNs{
				Name:      ci.Spec.Cluster,
				Namespace: ci.Namespace,
			}) && Outdated(&ci, &cio) {
			Mutate(&ci, &cio)
			st := ci.Status
			if err := r.Update(ctx, &ci); err != nil {
				return err
			}
			ci.Status = st
			if err := r.Status().Update(ctx, &ci); err != nil {
				return err
			}
		}
	}
	return nil
}

// Restore mutates an issue back to its original state.
func Restore(ci *v1alpha1.ClusterIssue) {
	if ci.Status.OrigCategory != nil {
		ci.Spec.Category = *ci.Status.OrigCategory
		ci.Status.OrigCategory = nil
	}
	if ci.Status.OrigSeverity != nil {
		ci.Spec.Severity = *ci.Status.OrigSeverity
		ci.Status.OrigSeverity = nil
	}
	if ci.Status.OrigMessage != nil {
		ci.Spec.Message = *ci.Status.OrigMessage
		ci.Status.OrigMessage = nil
	}
	ci.Status.Hidden = false
}

// This function is used to restore an issue to its original state whenever
// an override is deleted.
func (r *ClusterIssueOverrideReconciler) reconcileDelete(ctx context.Context,
	cio *v1alpha1.ClusterIssueOverride, cilist *v1alpha1.ClusterIssueList) error {
	for _, ci := range cilist.Items {
		if ci.Spec.ID == cio.Name && cio.InCluster(
			v1alpha1.NameNs{
				Name:      ci.Spec.Cluster,
				Namespace: ci.Namespace,
			}) {
			Restore(&ci)
			if err := r.Update(ctx, &ci); err != nil {
				return err
			}
		}
	}
	return nil
}

// This function is a <ClusterScan> trigger, whereby a request to this reconciler
// is created when the given Job is known to be from Zora's plugins.
func (r *ClusterIssueOverrideReconciler) completeScanTrigger(o client.Object) []ctrl.Request {
	ciolist := &v1alpha1.ClusterIssueOverrideList{}
	if err := r.List(context.Background(), ciolist, client.Limit(1)); err != nil || len(ciolist.Items) == 0 {
		return nil
	}

	return []ctrl.Request{{
		NamespacedName: client.ObjectKey{
			Name:      ciolist.Items[0].Name,
			Namespace: ciolist.Items[0].Namespace,
		},
	}}
}

// SetupWithManager sets up the controller with the Manager. This controller
// not only reconciles <ClusterIssueOverrides> but also <ClusterIssues> by
// watching Jobs from Zora's plugins.
func (r *ClusterIssueOverrideReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&v1alpha1.ClusterIssueOverride{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Watches(
			&source.Kind{Type: &v1alpha1.ClusterScan{}},
			handler.EnqueueRequestsFromMapFunc(r.completeScanTrigger),
		).Complete(r)
}
