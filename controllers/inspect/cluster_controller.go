package inspect

import (
	"context"
	"fmt"
	"reflect"
	"time"

	batchv1 "k8s.io/api/batch/v1"
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

	"github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/pkg/discovery"
	"github.com/getupio-undistro/inspect/pkg/kubeconfig"
)

const clusterScanRefKey = ".metadata.cluster"

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Config   *rest.Config
}

//+kubebuilder:rbac:groups=inspect.undistro.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inspect.undistro.io,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inspect.undistro.io,resources=clusters/finalizers,verbs=update
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

	info, err := discoverer.Discover(ctx)
	if err != nil {
		log.Error(err, "failed to discovery cluster info")
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterDiscovered, false, "ClusterNotDiscovered", err.Error())
		return err
	}

	if err := r.updateScanStatus(ctx, cluster); err != nil {
		return err
	}

	cluster.Status.SetClusterInfo(*info)
	cluster.Status.LastReconciliationTime = metav1.NewTime(time.Now().UTC())
	cluster.Status.ObservedGeneration = cluster.Generation
	r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterDiscovered, true, "ClusterDiscovered", "cluster info successfully discovered")
	return nil
}

// updateScanStatus updates status of the given Cluster based on ClusterScans that reference it
func (r *ClusterReconciler) updateScanStatus(ctx context.Context, cluster *v1alpha1.Cluster) error {
	log := ctrllog.FromContext(ctx)

	clusterScanList := &v1alpha1.ClusterScanList{}
	if err := r.List(ctx, clusterScanList, client.MatchingFields{clusterScanRefKey: cluster.Name}); err != nil {
		log.Error(err, fmt.Sprintf("failed to list ClusterScan referenced by Cluster %s", cluster.Name))
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterScanned, false, "ClusterScanListError", err.Error())
		return err
	}

	var totalIssues int
	var lastScanIDs []string
	var failed []string
	var notFinished []string
	for _, cs := range clusterScanList.Items {
		totalIssues += cs.Status.TotalIssues
		lastScanIDs = append(lastScanIDs, cs.Status.LastScanIDs(true)...)
		if cs.Status.LastFinishedScanStatus == string(batchv1.JobFailed) {
			failed = append(failed, cs.Name)
		} else if cs.Status.LastFinishedTime == nil {
			notFinished = append(notFinished, cs.Name)
		}
	}

	if len(clusterScanList.Items) <= 0 {
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterScanned, false, "ClusterScanNotConfigured", "no scan configured")
	} else if len(failed) > 0 {
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterScanned, false, "ClusterScanFailed", "last scan failed")
	} else if len(notFinished) == len(clusterScanList.Items) {
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterScanned, false, "ClusterNotScanned", "no finished scan yet")
	} else {
		r.setStatusAndCreateEvent(cluster, v1alpha1.ClusterScanned, true, "ClusterScanned", "cluster successfully scanned")
	}

	cluster.Status.TotalIssues = totalIssues
	cluster.Status.LastScans = lastScanIDs

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
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1alpha1.ClusterScan{}, clusterScanRefKey, func(rawObj client.Object) []string {
		clusterScan := rawObj.(*v1alpha1.ClusterScan)
		if clusterName := clusterScan.Spec.ClusterRef.Name; clusterName == "" {
			return nil
		} else {
			return []string{clusterName}
		}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Cluster{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&v1alpha1.ClusterScan{}).
		Complete(r)
}
