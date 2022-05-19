package inspect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/pkg/kubeconfig"
	"github.com/getupio-undistro/inspect/pkg/plugins/cronjobs"
)

// ClusterScanReconciler reconciles a ClusterScan object
type ClusterScanReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	DefaultPluginsNamespace string
	DefaultPluginsNames     []string
	WorkerImage             string
}

//+kubebuilder:rbac:groups=inspect.undistro.io,resources=clusterscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inspect.undistro.io,resources=clusterscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inspect.undistro.io,resources=clusterscans/finalizers,verbs=update
//+kubebuilder:rbac:groups=inspect.undistro.io,resources=plugins,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	clusterscan := &v1alpha1.ClusterScan{}
	if err := r.Get(ctx, req.NamespacedName, clusterscan); err != nil {
		log.Error(err, "failed to fetch ClusterIssue")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := r.reconcile(ctx, clusterscan)
	if err := r.Status().Update(ctx, clusterscan); err != nil {
		log.Error(err, "failed to update ClusterScan status")
	}

	return ctrl.Result{RequeueAfter: 5 * time.Minute}, err
}

func (r *ClusterScanReconciler) reconcile(ctx context.Context, clusterscan *v1alpha1.ClusterScan) error {
	log := ctrllog.FromContext(ctx)

	cluster := &v1alpha1.Cluster{}
	clusterKey := clusterscan.ClusterKey()
	if err := r.Get(ctx, clusterKey, cluster); err != nil {
		log.Error(err, "failed to fetch Cluster")
		clusterscan.SetReadyStatus(false, "ClusterFetchError", err.Error())
		return err
	}

	if !cluster.Status.ConditionIsTrue(v1alpha1.ClusterReady) {
		err := errors.New(fmt.Sprintf("the Cluster %s is not Ready", cluster.Name))
		log.Error(err, "cluster is not ready")
		clusterscan.SetReadyStatus(false, "ClusterNotReady", err.Error())
		return err
	}
	kubeconfigSecret, err := kubeconfig.SecretFromRef(ctx, r.Client, *cluster.KubeconfigRefKey())
	if err != nil {
		log.Error(err, "failed to get kubeconfig secret")
		clusterscan.SetReadyStatus(false, "ClusterKubeconfigError", err.Error())
		return err
	}

	pluginRefs := r.defaultPlugins()
	if len(clusterscan.Spec.Plugins) > 0 {
		log.Info("plugin references are provided in ClusterScan")
		pluginRefs = clusterscan.Spec.Plugins
	}

	pluginNames := make([]string, 0, len(pluginRefs))
	for _, ref := range pluginRefs {
		pluginKey := ref.PluginKey(r.DefaultPluginsNamespace)
		plugin := &v1alpha1.Plugin{}
		if err := r.Get(ctx, pluginKey, plugin); err != nil {
			log.Error(err, fmt.Sprintf("failed to fetch Plugin %v", pluginKey))
			clusterscan.SetReadyStatus(false, "PluginFetchError", err.Error())
			return err
		}
		cronJob := cronjobs.New(fmt.Sprintf("%s-%s", cluster.Name, plugin.Name), kubeconfigSecret.Namespace)

		cronJobMutator := cronjobs.Mutator{
			Scheme:           r.Scheme,
			WorkerImage:      r.WorkerImage,
			Existing:         cronJob,
			Plugin:           plugin,
			PluginRef:        ref,
			Clusterscan:      clusterscan,
			KubeconfigSecret: kubeconfigSecret,
		}

		result, err := ctrl.CreateOrUpdate(ctx, r.Client, cronJob, cronJobMutator.Mutate())
		if err != nil {
			log.Error(err, fmt.Sprintf("failed to apply CronJob %s", cronJob.Name))
			clusterscan.SetReadyStatus(false, "CronJobApplyError", err.Error())
			return err
		}
		msg := fmt.Sprintf("CronJob %s has been %v", cronJob.Name, result)
		log.Info(msg)
		if result != controllerutil.OperationResultNone {
			r.Recorder.Event(clusterscan, corev1.EventTypeNormal, "CronJobConfigured", msg)
		}
		pluginNames = append(pluginNames, plugin.Name)
	}

	// update ClusterScan status
	clusterscan.Status.Suspend = pointer.BoolDeref(clusterscan.Spec.Suspend, false)
	clusterscan.Status.ClusterNamespacedName = clusterKey.String()
	clusterscan.Status.Plugins = strings.Join(pluginNames, ",")
	clusterscan.Status.ObservedGeneration = clusterscan.Generation
	clusterscan.SetReadyStatus(true, "ClusterScanReconciled", fmt.Sprintf("cluster scan successfully configured for plugins %s", clusterscan.Status.Plugins))
	return nil
}

func (r *ClusterScanReconciler) defaultPlugins() []v1alpha1.PluginReference {
	defaultPlugins := make([]v1alpha1.PluginReference, 0, len(r.DefaultPluginsNames))
	for _, name := range r.DefaultPluginsNames {
		defaultPlugins = append(defaultPlugins, v1alpha1.PluginReference{
			Name:      name,
			Namespace: r.DefaultPluginsNamespace,
		})
	}
	return defaultPlugins
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterScan{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
