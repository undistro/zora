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
	"sort"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/undistro/zora/pkg/saas"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/undistro/zora/apis/zora/v1alpha1"
	"github.com/undistro/zora/pkg/kubeconfig"
	"github.com/undistro/zora/pkg/plugins/cronjobs"
	"github.com/undistro/zora/pkg/plugins/errparse"
)

const (
	jobOwnerKey          = ".metadata.controller"
	defaultReqTime       = 5
	clusterscanFinalizer = "clusterscan.zora.undistro.io/finalizer"
)

// ClusterScanReconciler reconciles a ClusterScan object
type ClusterScanReconciler struct {
	client.Client
	K8sClient               *kubernetes.Clientset
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	DefaultPluginsNamespace string
	DefaultPluginsNames     []string
	WorkerImage             string
	ClusterRoleBindingName  string
	ServiceAccountName      string
	OnUpdate                saas.ClusterScanHook
	OnDelete                saas.ClusterScanHook
}

//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterscans/finalizers,verbs=update
//+kubebuilder:rbac:groups=zora.undistro.io,resources=plugins,verbs=get;list;watch
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterissues,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=zora.undistro.io,resources=clusterissues/status,verbs=get
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs/status,verbs=get
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/status,verbs=get
//+kubebuilder:rbac:groups="",resources=pods;pods/log,verbs=get;list
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	clusterscan := &v1alpha1.ClusterScan{}
	if err := r.Get(ctx, req.NamespacedName, clusterscan); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("resourceVersion", clusterscan.ResourceVersion)
	ctx = ctrllog.IntoContext(ctx, log)

	defer func(t time.Time) {
		log.Info(fmt.Sprintf("ClusterScan has been reconciled in %v", time.Since(t)))
	}(time.Now())

	if clusterscan.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(clusterscan, clusterscanFinalizer) {
			controllerutil.AddFinalizer(clusterscan, clusterscanFinalizer)
			if err := r.Update(ctx, clusterscan); err != nil {
				log.Error(err, "failed to add finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		log.Info("the ClusterScan is being deleted")
		if controllerutil.ContainsFinalizer(clusterscan, clusterscanFinalizer) {
			if r.OnDelete != nil {
				if err := r.OnDelete(ctx, *clusterscan); err != nil {
					log.Error(err, "error in delete hook")
				}
			}
			controllerutil.RemoveFinalizer(clusterscan, clusterscanFinalizer)
			if err := r.Update(ctx, clusterscan); err != nil {
				log.Error(err, "failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	t, err := r.reconcile(ctx, clusterscan)
	if err := r.Status().Update(ctx, clusterscan); err != nil {
		log.Error(err, "failed to update ClusterScan status")
	}

	if r.OnUpdate != nil {
		if err := r.OnUpdate(ctx, *clusterscan); err != nil {
			log.Error(err, "error in update hook")
		}
	}

	return ctrl.Result{RequeueAfter: time.Duration(t) * time.Minute}, err
}

func (r *ClusterScanReconciler) reconcile(ctx context.Context, clusterscan *v1alpha1.ClusterScan) (int, error) {
	var notReadyErr error
	reqTime := defaultReqTime
	log := ctrllog.FromContext(ctx)

	cluster := &v1alpha1.Cluster{}
	if err := r.Get(ctx, clusterscan.ClusterKey(), cluster); err != nil {
		log.Error(err, fmt.Sprintf("failed to fetch Cluster %s", clusterscan.Spec.ClusterRef.Name))
		clusterscan.SetReadyStatus(false, "ClusterFetchError", err.Error())
		return reqTime, err
	}

	if !cluster.Status.ConditionIsTrue(v1alpha1.ClusterReady) {
		notReadyErr = fmt.Errorf("the Cluster %s is not Ready", cluster.Name)
		log.Error(notReadyErr, "Cluster is not ready")
		clusterscan.SetReadyStatus(false, "ClusterNotReady", notReadyErr.Error())
	}
	kubeconfigKey := cluster.KubeconfigRefKey()
	kubeconfigSecret, err := kubeconfig.SecretFromRef(ctx, r.Client, *kubeconfigKey)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to get kubeconfig secret %s", kubeconfigKey.String()))
		clusterscan.SetReadyStatus(false, "ClusterKubeconfigError", err.Error())
		return reqTime, err
	}

	if err := r.setControllerReference(ctx, clusterscan, cluster); err != nil {
		clusterscan.SetReadyStatus(false, "ClusterScanSetOwnerError", err.Error())
		return reqTime, err
	}

	if err := r.applyRBAC(ctx, clusterscan); err != nil {
		return reqTime, err
	}

	pluginRefs := r.defaultPlugins()
	if len(clusterscan.Spec.Plugins) > 0 {
		log.V(1).Info("plugin references are provided in ClusterScan")
		pluginRefs = clusterscan.Spec.Plugins
	}

	for _, ref := range pluginRefs {
		pluginKey := ref.PluginKey(r.DefaultPluginsNamespace)
		plugin := &v1alpha1.Plugin{}
		if err := r.Get(ctx, pluginKey, plugin); err != nil {
			log.Error(err, fmt.Sprintf("failed to fetch Plugin %s", pluginKey.String()))
			clusterscan.SetReadyStatus(false, "PluginFetchError", err.Error())
			return reqTime, err
		}
		cronJob := cronjobs.New(fmt.Sprintf("%s-%s", clusterscan.Name, plugin.Name), kubeconfigSecret.Namespace)

		cronJobMutator := &cronjobs.Mutator{
			Scheme:             r.Scheme,
			Existing:           cronJob,
			Plugin:             plugin,
			PluginRef:          ref,
			ClusterScan:        clusterscan,
			KubeconfigSecret:   kubeconfigSecret,
			WorkerImage:        r.WorkerImage,
			ServiceAccountName: r.ServiceAccountName,
			Suspend:            notReadyErr != nil,
		}

		f, rem := cronJobMutator.Mutate()
		if rem != 0 && (reqTime == defaultReqTime || reqTime >= rem) {
			reqTime = rem
			log.Info(fmt.Sprintf("Delaying Cronjob creation for plugin <%s> by %d minutes", plugin.Name, reqTime))
			continue
		}

		result, err := ctrl.CreateOrUpdate(ctx, r.Client, cronJob, f)
		if err != nil {
			log.Error(err, fmt.Sprintf("failed to apply CronJob %s", cronJob.Name))
			clusterscan.SetReadyStatus(false, "CronJobApplyError", err.Error())
			return reqTime, err
		}
		if result != controllerutil.OperationResultNone {
			msg := fmt.Sprintf("CronJob %s has been %s", cronJob.Name, result)
			log.Info(msg)
			r.Recorder.Event(clusterscan, corev1.EventTypeNormal, "CronJobConfigured", msg)
		}

		pluginStatus := clusterscan.Status.GetPluginStatus(plugin.Name)
		pluginStatus.Suspend = *cronJob.Spec.Suspend
		if sched, err := cron.ParseStandard(cronJob.Spec.Schedule); err != nil {
			log.Error(err, "failed to parse CronJob Schedule")
		} else {
			pluginStatus.NextScheduleTime = &metav1.Time{Time: sched.Next(time.Now().UTC())}
		}
		if cronJob.Status.LastScheduleTime != nil {
			log.V(1).Info(fmt.Sprintf("CronJob %s has scheduled jobs", cronJob.Name))
			if j, err := r.getLastJob(ctx, cronJob); err != nil {
				clusterscan.SetReadyStatus(false, "JobListError", err.Error())
				return reqTime, err
			} else if j != nil {
				isFinished, status, finTime := getFinishedStatus(j)
				if isFinished {
					pluginStatus.LastFinishedStatus = string(status)

					if status == batchv1.JobFailed {
						if err := r.pluginErrorMsg(ctx, pluginStatus, plugin, j); err != nil {
							return reqTime, err
						}
					}
				} else if len(cronJob.Status.Active) > 0 {
					status = "Active"
				}
				pluginStatus.LastStatus = string(status)
				pluginStatus.LastScanID = string(j.UID)
				pluginStatus.LastScheduleTime = cronJob.Status.LastScheduleTime
				pluginStatus.LastFinishedTime = finTime
				if status == batchv1.JobComplete {
					pluginStatus.LastSuccessfulScanID = string(j.UID)
					pluginStatus.LastSuccessfulTime = finTime
				}
			}
		}
	}

	if issues, err := r.getClusterIssues(ctx, clusterscan.Status.LastScanIDs(true)...); err != nil {
		clusterscan.SetReadyStatus(false, "ClusterIssueListError", err.Error())
		return reqTime, err
	} else if issues != nil {
		issc := map[string]int{}
		for _, i := range issues {
			issc[i.Labels[v1alpha1.LabelPlugin]]++
		}
		for p, c := range issc {
			if clusterscan.Status.Plugins[p].IssueCount == nil {
				clusterscan.Status.Plugins[p].IssueCount = new(int)
			}
			*clusterscan.Status.Plugins[p].IssueCount = c
		}
		clusterscan.Status.TotalIssues = pointer.Int(len(issues))
	}

	clusterscan.Status.SyncStatus()
	clusterscan.Status.Suspend = notReadyErr != nil
	if notReadyErr == nil {
		clusterscan.Status.Suspend = pointer.BoolDeref(clusterscan.Spec.Suspend, false)
	}
	clusterscan.Status.ObservedGeneration = clusterscan.Generation
	clusterscan.SetReadyStatus(true, "ClusterScanReconciled", fmt.Sprintf("cluster scan successfully configured for plugins: %s", clusterscan.Status.PluginNames))
	return reqTime, notReadyErr
}

func (r *ClusterScanReconciler) pluginErrorMsg(ctx context.Context, ps *v1alpha1.PluginScanStatus, p *v1alpha1.Plugin, j *batchv1.Job) error {
	log := ctrllog.FromContext(ctx, v1alpha1.LabelPlugin, p.Name)

	plist, err := r.K8sClient.CoreV1().Pods(j.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", j.Name),
	})
	if err != nil {
		return fmt.Errorf("Unable to list pods: %w", err)
	} else if len(plist.Items) == 0 {
		return fmt.Errorf("The job <%s/%s> has no pods", j.Namespace, j.Name)
	}

	lr, err := r.K8sClient.CoreV1().Pods(plist.Items[0].Namespace).GetLogs(
		plist.Items[0].Name,
		&corev1.PodLogOptions{Container: p.Name},
	).Stream(ctx)
	if err != nil {
		return fmt.Errorf("Unable to fetch <%s> container logs: %w", p.Name, err)
	}

	defer lr.Close()
	ps.LastErrorMsg, err = errparse.Parse(lr, p.Name)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to extract <%s> error message", p.Name))
	}
	return nil
}

// getClusterIssues returns the ClusterIssues from scanIDs
func (r *ClusterScanReconciler) getClusterIssues(ctx context.Context, scanIDs ...string) ([]v1alpha1.ClusterIssue, error) {
	log := ctrllog.FromContext(ctx, v1alpha1.LabelScanID, scanIDs)

	if len(scanIDs) <= 0 {
		return nil, nil
	}
	req, err := labels.NewRequirement(v1alpha1.LabelScanID, selection.In, scanIDs)
	if err != nil {
		log.Error(err, "failed to build label selector")
		return nil, err
	}
	sel := labels.NewSelector().Add(*req)
	list := &v1alpha1.ClusterIssueList{}
	if err := r.List(ctx, list, client.MatchingLabelsSelector{Selector: sel}); err != nil {
		log.Error(err, "failed to list ClusterIssues")
		return nil, err
	}
	log.V(1).Info(fmt.Sprintf("%d ClusterIssues found", len(list.Items)))
	return list.Items, nil
}

// getLastJob returns the last Job from a CronJob
func (r *ClusterScanReconciler) getLastJob(ctx context.Context, cronJob *batchv1.CronJob) (*batchv1.Job, error) {
	log := ctrllog.FromContext(ctx, "CronJob", cronJob.Name)

	jobList := &batchv1.JobList{}
	if err := r.List(ctx, jobList, client.MatchingFields{jobOwnerKey: cronJob.Name}); err != nil {
		log.Error(err, "failed to list Jobs")
		return nil, err
	}
	log.V(1).Info(fmt.Sprintf("found %d Jobs", len(jobList.Items)))
	sort.Slice(jobList.Items, func(i, j int) bool {
		if jobList.Items[i].Status.StartTime == nil {
			return jobList.Items[j].Status.StartTime != nil
		}
		return jobList.Items[j].Status.StartTime.Before(jobList.Items[i].Status.StartTime)
	})

	if len(jobList.Items) > 0 {
		return &jobList.Items[0], nil
	}
	return nil, nil
}

// setControllerReference add Cluster as owner (controller) of ClusterScan, add a label and update
func (r *ClusterScanReconciler) setControllerReference(ctx context.Context, clusterscan *v1alpha1.ClusterScan, cluster *v1alpha1.Cluster) error {
	if metav1.IsControlledBy(clusterscan, cluster) && clusterscan.Labels != nil && clusterscan.Labels[v1alpha1.LabelCluster] == cluster.Name {
		return nil
	}
	log := ctrllog.FromContext(ctx)
	if err := ctrl.SetControllerReference(cluster, clusterscan, r.Scheme); err != nil {
		log.Error(err, "failed to set Cluster as Owner of ClusterScan")
		return err
	}
	if clusterscan.Labels == nil {
		clusterscan.Labels = map[string]string{}
	}
	clusterscan.Labels[v1alpha1.LabelCluster] = cluster.Name
	if err := r.Update(ctx, clusterscan); err != nil {
		log.Error(err, "failed to update ClusterScan")
		return err
	}
	log.Info("ClusterScan updated")
	return nil
}

func (r *ClusterScanReconciler) defaultPlugins() []v1alpha1.PluginReference {
	p := make([]v1alpha1.PluginReference, 0, len(r.DefaultPluginsNames))
	for _, name := range r.DefaultPluginsNames {
		p = append(p, v1alpha1.PluginReference{
			Name:      name,
			Namespace: r.DefaultPluginsNamespace,
		})
	}
	return p
}

// applyRBAC Create or Update a ServiceAccount (with ClusterScan as Owner) and append it to ClusterRoleBinding
func (r *ClusterScanReconciler) applyRBAC(ctx context.Context, clusterscan *v1alpha1.ClusterScan) error {
	log := ctrllog.FromContext(ctx)

	crb := &rbacv1.ClusterRoleBinding{}
	crbKey := client.ObjectKey{Name: r.ClusterRoleBindingName}
	if err := r.Get(ctx, crbKey, crb); err != nil {
		log.Error(err, fmt.Sprintf("failed to fetch ClusterRoleBinding %s", r.ClusterRoleBindingName))
		clusterscan.SetReadyStatus(false, "ClusterRoleBindingFetchError", err.Error())
		return err
	}

	sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: r.ServiceAccountName, Namespace: clusterscan.Namespace}}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, sa, func() error {
		return controllerutil.SetOwnerReference(clusterscan, sa, r.Scheme)
	})
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to apply ServiceAccount %s", r.ServiceAccountName))
		clusterscan.SetReadyStatus(false, "ServiceAccountApplyError", err.Error())
		return err
	}
	if res != controllerutil.OperationResultNone {
		msg := fmt.Sprintf("ServiceAccount %s has been %s", r.ServiceAccountName, res)
		log.Info(msg)
		r.Recorder.Event(clusterscan, corev1.EventTypeNormal, "ServiceAccountConfigured", msg)
	}
	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      sa.Name,
		Namespace: sa.Namespace,
	}
	exists := false
	for _, s := range crb.Subjects {
		if s.Kind == subject.Kind && s.Namespace == subject.Namespace && s.Name == subject.Name {
			exists = true
			break
		}
	}
	if !exists {
		crb.Subjects = append(crb.Subjects, subject)
		if err := r.Update(ctx, crb); err != nil {
			log.Error(err, fmt.Sprintf("failed to update ClusterRoleBinding %s", crb.Name))
			clusterscan.SetReadyStatus(false, "ClusterRoleBindingUpdateError", err.Error())
			return err
		}
		msg := fmt.Sprintf("ClusterRoleBinding %s has been updated", crb.Name)
		log.Info(msg)
		r.Recorder.Event(clusterscan, corev1.EventTypeNormal, "ClusterRoleBindingConfigured", msg)
	} else {
		log.V(1).Info(fmt.Sprintf("ServiceAccount %s/%s already exists in ClusterRoleBinding %s subjects", sa.Namespace, sa.Name, crb.Name))
	}
	return nil
}

// getFinishedStatus return true if Complete or Failed condition is True
func getFinishedStatus(job *batchv1.Job) (bool, batchv1.JobConditionType, *metav1.Time) {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type, &c.LastTransitionTime
		}
	}
	return false, "", nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.Job{}, jobOwnerKey, func(rawObj client.Object) []string {
		job := rawObj.(*batchv1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != batchv1.SchemeGroupVersion.String() || owner.Kind != "CronJob" {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterScan{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
