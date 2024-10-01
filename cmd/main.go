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

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/undistro/zora/pkg/crds"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"go.uber.org/zap/zapcore"
	apiextensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	zorav1alpha1 "github.com/undistro/zora/api/zora/v1alpha1"
	zorav1alpha2 "github.com/undistro/zora/api/zora/v1alpha2"
	zoracontroller "github.com/undistro/zora/internal/controller/zora"
	"github.com/undistro/zora/internal/saas"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(zorav1alpha1.AddToScheme(scheme))
	utilruntime.Must(zorav1alpha2.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var defaultPluginsNamespace string
	var defaultPluginsNames string
	var workerImage string
	var cronJobClusterRoleBinding string
	var cronJobServiceAccount string
	var cronJobAnnotations string
	var saasWorkspaceID string
	var saasServer string
	var version string
	var checksConfigMapNamespace string
	var checksConfigMapName string
	var kubexnsImage string
	var kubexnsPullPolicy string
	var trivyPVC string
	var updateCRDs bool
	var injectConversion bool
	var caPath string
	var webhookServiceName string
	var webhookServiceNamespace string
	var webhookServicePath string
	var tokenPath string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.StringVar(&defaultPluginsNamespace, "default-plugins-namespace", "zora-system", "The namespace of default plugins")
	flag.StringVar(&defaultPluginsNames, "default-plugins-names", "marvin,popeye", "Comma separated list of default plugins")
	flag.StringVar(&workerImage, "worker-image", "ghcr.io/undistro/zora/worker:latest", "Docker image name of Worker container")
	flag.StringVar(&cronJobClusterRoleBinding, "cronjob-clusterrolebinding-name", "zora-plugins-rolebinding", "Name of ClusterRoleBinding to append CronJob ServiceAccounts")
	flag.StringVar(&cronJobServiceAccount, "cronjob-serviceaccount-name", "zora-plugins", "Name of ServiceAccount to be configured, appended to ClusterRoleBinding and used by CronJobs")
	flag.StringVar(&cronJobAnnotations, "cronjob-serviceaccount-annotations", "annotaion1=value1,annotation2=value2", "Annotations to be applied to the CronJob Service Account")
	flag.StringVar(&saasWorkspaceID, "saas-workspace-id", "", "Your workspace ID in Zora SaaS")
	flag.StringVar(&saasServer, "saas-server", "http://localhost:3003", "Address for Zora's saas server")
	flag.StringVar(&version, "version", "0.10.1", "Zora version")
	flag.StringVar(&checksConfigMapNamespace, "checks-configmap-namespace", "zora-system", "Namespace of custom checks ConfigMap")
	flag.StringVar(&checksConfigMapName, "checks-configmap-name", "zora-custom-checks", "Name of custom checks ConfigMap")
	flag.StringVar(&kubexnsImage, "kubexns-image", "ghcr.io/undistro/kubexns:latest", "kubexns image")
	flag.StringVar(&kubexnsPullPolicy, "kubexns-pull-policy", "Always", "kubexns image pull policy")
	flag.StringVar(&trivyPVC, "trivy-db-pvc", "", "PersistentVolumeClaim name for Trivy DB")
	flag.BoolVar(&updateCRDs, "update-crds", false,
		"If set to true, operator will update Zora CRDs if needed")
	flag.BoolVar(&injectConversion, "inject-conversion", false,
		"If set to true, operator will inject webhook conversion in annotated CRDs")
	flag.StringVar(&caPath, "ca-path", "/tmp/k8s-webhook-server/serving-certs/ca.crt",
		"Path of CA file to be injected in CRDs")
	flag.StringVar(&webhookServiceName, "webhook-service-name", "zora-webhook",
		"Webhook service name")
	flag.StringVar(&webhookServiceNamespace, "webhook-service-namespace", "zora-system",
		"Webhook service namespace")
	flag.StringVar(&webhookServicePath, "webhook-service-path", "/convert",
		"URL path for webhook conversion")
	flag.StringVar(&tokenPath, "token-path", "/tmp/jwt-tokens/tokens",
		"URL path for authorization tokens")

	done := make(chan struct{})
	defer close(done)

	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.TimeEncoderOfLayout(time.RFC3339),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}
	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	restConfig := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			TLSOpts:       tlsOpts,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e0f4eef4.zora.undistro.io",
		WebhookServer:          webhookServer,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	var onClusterUpdate, onClusterDelete saas.ClusterHook
	var onClusterScanUpdate, onClusterScanDelete saas.ClusterScanHook
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	if saasWorkspaceID != "" {
		saasClient, err := saas.NewClient(saasServer, version, saasWorkspaceID, client, tokenPath, done)
		if err != nil {
			setupLog.Error(err, "unable to create SaaS client", "workspaceID", saasWorkspaceID)
			os.Exit(1)
		}
		setupLog.Info("registering SaaS hooks on reconcilers", "workspaceID", saasWorkspaceID)
		onClusterUpdate = saas.UpdateClusterHook(saasClient)
		onClusterDelete = saas.DeleteClusterHook(saasClient)
		onClusterScanUpdate = saas.UpdateClusterScanHook(saasClient, mgr.GetClient())
		onClusterScanDelete = saas.DeleteClusterScanHook(saasClient, mgr.GetClient())
	}

	if err = (&zoracontroller.ClusterReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("cluster-controller"),
		Config:   restConfig,
		OnUpdate: onClusterUpdate,
		OnDelete: onClusterDelete,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}

	annotations, err := annotations(cronJobAnnotations)
	if err != nil {
		setupLog.Error(err, "unable to parse annotations")
		os.Exit(1)
	}
	if err = (&zoracontroller.ClusterScanReconciler{
		Client:                  mgr.GetClient(),
		K8sClient:               kubernetes.NewForConfigOrDie(restConfig),
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("clusterscan-controller"),
		DefaultPluginsNamespace: defaultPluginsNamespace,
		DefaultPluginsNames:     strings.Split(defaultPluginsNames, ","),
		WorkerImage:             workerImage,
		ClusterRoleBindingName:  cronJobClusterRoleBinding,
		ServiceAccountName:      cronJobServiceAccount,
		Annotations:             annotations,
		OnUpdate:                onClusterScanUpdate,
		OnDelete:                onClusterScanDelete,
		KubexnsImage:            kubexnsImage,
		KubexnsPullPolicy:       kubexnsPullPolicy,
		TrivyPVC:                trivyPVC,
		ChecksConfigMap:         fmt.Sprintf("%s/%s", checksConfigMapNamespace, checksConfigMapName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterScan")
		os.Exit(1)
	}
	if err = (&zoracontroller.CustomCheckReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		ConfigMapNamespace: checksConfigMapNamespace,
		ConfigMapName:      checksConfigMapName,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CustomCheck")
		os.Exit(1)
	}
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&zorav1alpha1.VulnerabilityReport{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "VulnerabilityReport")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
	ctx := ctrl.SetupSignalHandler()

	if updateCRDs || injectConversion {
		extClient := apiextensionsv1client.NewForConfigOrDie(restConfig)
		copts := crds.NewConversionOptions(injectConversion, webhookServiceName, webhookServiceNamespace, webhookServicePath, caPath)
		if err := crds.Update(ctrllog.IntoContext(ctx, setupLog), extClient, *copts); err != nil {
			setupLog.Error(err, "unable to update CRDs")
			os.Exit(1)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func annotations(cronJobAnnotations string) (map[string]string, error) {
	cronJobAnnotations = strings.Trim(cronJobAnnotations, " ")
	if len(cronJobAnnotations) == 0 {
		return nil, nil
	}
	annotations := map[string]string{}
	for _, annotation := range strings.Split(cronJobAnnotations, ",") {
		index := strings.Index(annotation, "=")
		if index == -1 || index == len(annotation) {
			return nil, fmt.Errorf("Could not parse annotation %s", annotation)
		}
		key := annotation[:index]
		value := annotation[index+1:]
		annotations[key] = value
	}
	return annotations, nil
}
