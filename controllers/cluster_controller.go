package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/getupio-undistro/snitch/api/v1alpha1"
	"github.com/getupio-undistro/snitch/pkg/discovery"
	"github.com/getupio-undistro/snitch/pkg/kubeconfig"
	"github.com/getupio-undistro/snitch/pkg/provider/cloud"
	"github.com/getupio-undistro/snitch/pkg/provider/cloud/eks"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Config   *rest.Config
}

//+kubebuilder:rbac:groups=snitch.undistro.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=snitch.undistro.io,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=snitch.undistro.io,resources=clusters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx, "name", req.Name, "namespace", req.Namespace)

	cluster := &v1alpha1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		log.Error(err, "failed to fetch Cluster")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	result, err := r.reconcile(ctx, cluster)
	if err != nil {
		r.Recorder.Event(cluster, corev1.EventTypeWarning, "ClusterReconcileFailed", err.Error())
	} else {
		r.Recorder.Event(cluster, corev1.EventTypeNormal, "ClusterReconciled", fmt.Sprintf("Cluster %v reconciled", req.NamespacedName))
	}
	return result, err
}

func (r *ClusterReconciler) reconcile(ctx context.Context, cluster *v1alpha1.Cluster) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx, "name", cluster.Name, "namespace", cluster.Namespace)

	// itself
	if cluster.Spec.KubeconfigRef == nil && cluster.Spec.Cloud == nil {
		if err := r.discoverAndUpdateStatus(ctx, cluster, r.Config); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
	}

	// raw kubeconfig
	if ref := cluster.Spec.KubeconfigRef; ref != nil {
		if ref.Namespace == "" {
			ref.Namespace = cluster.Namespace
		}
		key := client.ObjectKey{
			Namespace: ref.Namespace,
			Name:      ref.Name,
		}
		config, err := kubeconfig.ConfigFromSecretName(ctx, r.Client, key)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err := r.discoverAndUpdateStatus(ctx, cluster, config); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
	}

	// cloud
	cloudSpec := cluster.Spec.Cloud
	if cloudSpec == nil {
		return ctrl.Result{}, nil
	}
	secretKeySelector := cloud.DefaultSecretKeySelectorValueFunc(ctx, r.Client, cluster.Namespace)

	// EKS
	if cloudSpec.EKS != nil {
		ready, err := eks.ClusterIsReady(secretKeySelector, cloudSpec.EKS)
		if err != nil {
			log.Error(err, "failed to get EKS cluster status")
			return ctrl.Result{}, err
		}
		if !ready {
			log.Info("cluster is not ready, requeing after 1 minute")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
		}
		kcfg, err := eks.GetConfig(secretKeySelector, cloudSpec.EKS)
		if err != nil {
			log.Error(err, "failed to get config from EKS cluster")
			return ctrl.Result{}, err
		}
		secret, err := kubeconfig.ApplyKubeconfigSecret(ctx, r.Client, kcfg, cluster.GetKubeconfigSecretName(), cluster)
		if err != nil {
			log.Error(err, "failed to apply kubeconfig secret")
			return ctrl.Result{}, err
		}
		cfg, err := kubeconfig.ConfigFromSecret(secret)
		if err != nil {
			log.Error(err, "failed to get rest.Config from secret")
			return ctrl.Result{}, err
		}
		if err := r.discoverAndUpdateStatus(ctx, cluster, cfg); err != nil {
			return ctrl.Result{}, err
		}
		// the kubeconfig token is valid 14min, it will update token every 10min
		log.Info("reconciliation finished, next run in 10m")
		return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
	}

	// GKE
	if cloudSpec.GKE != nil {
		//TODO implements GKE (https://getupio.atlassian.net/browse/UD-32)
	}

	// AKS
	if cloudSpec.AKS != nil {
		//TODO implements AKS (https://getupio.atlassian.net/browse/UD-31)
	}

	return ctrl.Result{}, nil
}

// discoverAndUpdateStatus discover cluster info and update cluster status
func (r *ClusterReconciler) discoverAndUpdateStatus(ctx context.Context, cluster *v1alpha1.Cluster, config *rest.Config) error {
	log := ctrllog.FromContext(ctx, "name", cluster.Name, "namespace", cluster.Namespace)
	d, err := discovery.NewForConfig(config)
	if err != nil {
		log.Error(err, "failed to get new discovery client from config")
		return err
	}
	info, err := d.Discover(ctx)
	if err != nil {
		log.Error(err, "failed to discover cluster info")
		return err
	}
	cluster.Status.SetClusterInfo(*info)
	if err := r.Status().Update(ctx, cluster); err != nil {
		log.Error(err, "failed to update cluster status")
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Cluster{}).
		Complete(r)
}
