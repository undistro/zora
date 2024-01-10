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

package plugins

import (
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/undistro/zora/api/zora/v1alpha1"
	"github.com/undistro/zora/pkg/kubeconfig"
)

const (
	checksVolumeName           = "custom-checks"
	checksPath                 = "/custom-checks"
	workerContainerName        = "worker"
	kubeconfigVolumeName       = "kubeconfig"
	kubeconfigMountPath        = "/etc/zora"
	kubeconfigFile             = "kubeconfig.yml"
	resultsVolumeName          = "results"
	resultsDir                 = "/tmp/zora/results"
	labelClusterScan           = "zora.undistro.io/cluster-scan"
	labelPlugin                = "zora.undistro.io/plugin"
	annotationDefaultContainer = "kubectl.kubernetes.io/default-container"
)

var (
	// commonEnv environment variables to be used in worker and plugin containers
	commonEnv = []corev1.EnvVar{
		{
			Name:  "DONE_DIR",
			Value: resultsDir,
		},
		{
			Name:  "DONE_FILE",
			Value: filepath.Join(resultsDir, "done"),
		},
		{
			Name:  "ERROR_FILE",
			Value: filepath.Join(resultsDir, "error"),
		},
	}
	// commonVolumeMounts represents the volume mounts to be used in worker and plugin containers
	commonVolumeMounts = []corev1.VolumeMount{
		{
			Name:      resultsVolumeName,
			MountPath: resultsDir,
		},
	}
	// pluginVolumes represents the volume mounts to be used in plugin container
	kubeconfigVolumeMount = corev1.VolumeMount{
		Name:      kubeconfigVolumeName,
		ReadOnly:  true,
		MountPath: kubeconfigMountPath,
	}

	// customChecksVolume represents the volume mount to be used in the init container
	customChecksVolume = corev1.VolumeMount{Name: checksVolumeName, MountPath: checksPath}
)

func NewCronJob(name, namespace string) *batchv1.CronJob {
	return &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
}

type CronJobMutator struct {
	Scheme             *runtime.Scheme
	Existing           *batchv1.CronJob
	Plugin             *v1alpha1.Plugin
	PluginRef          v1alpha1.PluginReference
	ClusterScan        *v1alpha1.ClusterScan
	KubeconfigSecret   *corev1.Secret
	WorkerImage        string
	ServiceAccountName string
	Suspend            bool
	KubexnsImage       string
	ChecksConfigMap    string
	ClusterUID         types.UID
}

// Mutate returns a function which mutates the existing CronJob into it's desired state.
func (r *CronJobMutator) Mutate() error {
	if r.Existing.ObjectMeta.Labels == nil {
		r.Existing.ObjectMeta.Labels = make(map[string]string)
	}
	r.Existing.ObjectMeta.Labels[labelClusterScan] = r.ClusterScan.Name
	r.Existing.ObjectMeta.Labels[labelPlugin] = r.Plugin.Name
	r.Existing.Spec.Schedule = r.ClusterScan.Spec.Schedule
	r.Existing.Spec.ConcurrencyPolicy = batchv1.ForbidConcurrent
	r.Existing.Spec.SuccessfulJobsHistoryLimit = r.ClusterScan.Spec.SuccessfulScansHistoryLimit
	r.Existing.Spec.FailedJobsHistoryLimit = r.ClusterScan.Spec.FailedScansHistoryLimit

	r.Existing.Spec.Suspend = &r.Suspend
	if !r.Suspend {
		r.Existing.Spec.Suspend = r.ClusterScan.Spec.Suspend
	}
	r.Existing.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	r.Existing.Spec.JobTemplate.Spec.BackoffLimit = pointer.Int32(0)
	r.Existing.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = r.ServiceAccountName
	r.Existing.Spec.JobTemplate.Spec.Template.Annotations = map[string]string{annotationDefaultContainer: r.Plugin.Name}
	r.Existing.Spec.JobTemplate.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name:         resultsVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		},
		{
			Name:         checksVolumeName,
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
		},
	}
	if r.KubeconfigSecret != nil {
		//nolint:lll
		r.Existing.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(r.Existing.Spec.JobTemplate.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: kubeconfigVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  r.KubeconfigSecret.Name,
					DefaultMode: pointer.Int32(0644),
					Items:       []corev1.KeyToPath{{Key: kubeconfig.SecretField, Path: kubeconfigFile}},
				},
			},
		})
	}

	if pointer.BoolDeref(r.Plugin.Spec.MountCustomChecksVolume, false) {
		initContainer := r.initContainer()
		if len(r.Existing.Spec.JobTemplate.Spec.Template.Spec.InitContainers) == 0 {
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.InitContainers = []corev1.Container{initContainer}
		} else {
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].Name = initContainer.Name
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].Image = initContainer.Image
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].Env = initContainer.Env
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].VolumeMounts = initContainer.VolumeMounts
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].ImagePullPolicy = initContainer.ImagePullPolicy
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].Resources = initContainer.Resources
		}
	}

	desiredContainers := map[string]corev1.Container{
		workerContainerName: r.workerContainer(),
		r.Plugin.Name:       r.pluginContainer(),
	}

	foundContainers := 0
	for i, container := range r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers {
		desired, found := desiredContainers[container.Name]
		if found {
			foundContainers++
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Name = desired.Name
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Image = desired.Image
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Command = desired.Command
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Args = desired.Args
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].EnvFrom = desired.EnvFrom
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Env = desired.Env
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Resources = desired.Resources
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].ImagePullPolicy = desired.ImagePullPolicy
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].SecurityContext = desired.SecurityContext
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].VolumeMounts = desired.VolumeMounts
		}
	}
	if foundContainers != len(desiredContainers) {
		containers := make([]corev1.Container, 0, len(desiredContainers))
		for _, c := range desiredContainers {
			containers = append(containers, c)
		}
		r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers = containers
	}

	return ctrl.SetControllerReference(r.ClusterScan, r.Existing, r.Scheme)
}

// workerContainer returns a Container for Worker
func (r *CronJobMutator) workerContainer() corev1.Container {
	return corev1.Container{
		Name:            workerContainerName,
		Image:           r.WorkerImage,
		Env:             r.workerEnv(),
		EnvFrom:         r.Plugin.Spec.EnvFrom,
		Resources:       r.Plugin.Spec.Resources,
		VolumeMounts:    commonVolumeMounts,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             pointer.Bool(true),
			AllowPrivilegeEscalation: pointer.Bool(false),
		},
	}
}

// pluginContainer returns a Container for Plugin
func (r *CronJobMutator) pluginContainer() corev1.Container {
	c := corev1.Container{
		Name:            r.Plugin.Name,
		Image:           r.Plugin.Spec.Image,
		Command:         r.Plugin.Spec.Command,
		Args:            r.Plugin.Spec.Args,
		EnvFrom:         r.Plugin.Spec.EnvFrom,
		Env:             r.pluginEnv(),
		Resources:       r.Plugin.Spec.Resources,
		ImagePullPolicy: r.Plugin.Spec.GetImagePullPolicy(),
		SecurityContext: r.Plugin.Spec.SecurityContext,
		VolumeMounts:    commonVolumeMounts,
	}
	if r.KubeconfigSecret != nil {
		c.VolumeMounts = append(c.VolumeMounts, kubeconfigVolumeMount)
	}
	if pointer.BoolDeref(r.Plugin.Spec.MountCustomChecksVolume, false) {
		c.VolumeMounts = append(c.VolumeMounts, customChecksVolume)
	}
	return c
}

// initContainer returns an init container to mount custom checks
func (r *CronJobMutator) initContainer() corev1.Container {
	return corev1.Container{
		Name:  "checks",
		Image: r.KubexnsImage,
		Env: []corev1.EnvVar{
			{Name: "DIR", Value: checksPath},
			{Name: "CONFIGMAPS", Value: r.ChecksConfigMap},
			{Name: "IGNORE_NOT_FOUND", Value: "true"},
		},
		VolumeMounts:    []corev1.VolumeMount{customChecksVolume},
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources:       r.Plugin.Spec.Resources,
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             pointer.Bool(true),
			AllowPrivilegeEscalation: pointer.Bool(false),
		},
	}
}

// pluginEnv returns a list of environment variables to set in the Plugin container
func (r *CronJobMutator) pluginEnv() []corev1.EnvVar {
	p := append(r.Plugin.Spec.Env, r.PluginRef.Env...)
	p = append(p, commonEnv...)
	p = append(p,
		corev1.EnvVar{
			Name:  "CRONJOB_NAMESPACE",
			Value: r.Existing.ObjectMeta.Namespace,
		},
		corev1.EnvVar{
			Name:  "CRONJOB_NAME",
			Value: r.Existing.ObjectMeta.Name,
		},
	)
	if r.KubeconfigSecret != nil {
		p = append(p, corev1.EnvVar{
			Name:  "KUBECONFIG",
			Value: filepath.Join(kubeconfigMountPath, kubeconfigFile),
		})
	}
	if pointer.BoolDeref(r.Plugin.Spec.MountCustomChecksVolume, false) {
		p = append(p, corev1.EnvVar{Name: "CUSTOM_CHECKS_PATH", Value: checksVolumeName})
	}
	return p
}

// workerEnv returns a list of environment variables to set in the Worker container
func (r *CronJobMutator) workerEnv() []corev1.EnvVar {
	p := append(commonEnv, r.Plugin.Spec.Env...)
	p = append(p,
		corev1.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: r.ClusterScan.Spec.ClusterRef.Name,
		},
		corev1.EnvVar{
			Name:  "CLUSTER_UID",
			Value: string(r.ClusterUID),
		},
		corev1.EnvVar{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace", APIVersion: "v1"},
			},
		},
		corev1.EnvVar{
			Name:  "PLUGIN_NAME",
			Value: r.Plugin.Name,
		},
		corev1.EnvVar{
			Name:  "PLUGIN_TYPE",
			Value: r.Plugin.Spec.Type,
		},
		corev1.EnvVar{
			Name: "JOB_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['job-name']", APIVersion: "v1"},
			},
		},
		corev1.EnvVar{
			Name: "JOB_UID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['controller-uid']", APIVersion: "v1"},
			},
		},
		corev1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name", APIVersion: "v1"},
			},
		},
	)
	return p
}
