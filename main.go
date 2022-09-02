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
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	zorav1alpha1 "github.com/getupio-undistro/zora/apis/zora/v1alpha1"
	zoracontrollers "github.com/getupio-undistro/zora/controllers/zora"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(zorav1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var defaultPluginsNamespace string
	var defaultPluginsNames string
	var workerImage string
	var cronJobClusterRoleBinding string
	var cronJobServiceAccount string
	var saasID string
	var saasServerAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&defaultPluginsNamespace, "default-plugins-namespace", "zora-system", "The namespace of default plugins")
	flag.StringVar(&defaultPluginsNames, "default-plugins-names", "popeye", "Comma separated list of default plugins")
	flag.StringVar(&workerImage, "worker-image", "registry.undistro.io/library/worker:v0.3.6", "Docker image name of Worker container")
	flag.StringVar(&cronJobClusterRoleBinding, "cronjob-clusterrolebinding-name", "zora-plugins", "Name of ClusterRoleBinding to append CronJob ServiceAccounts")
	flag.StringVar(&cronJobServiceAccount, "cronjob-serviceaccount-name", "zora-plugins", "Name of ServiceAccount to be configured, appended to ClusterRoleBinding and used by CronJobs")
	flag.StringVar(&saasID, "saas-id", "", "ID of Zora's service offering")
	flag.StringVar(&saasServerAddr, "saas-server-address", "localhost:8082", "Address for Zora's saas server")

	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.TimeEncoderOfLayout(time.RFC3339),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e0f4eef4.zora.undistro.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&zoracontrollers.ClusterReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("cluster-controller"),
		Config:   mgr.GetConfig(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}

	kcli, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "Failed to create Kubernetes clientset", "controller", "Cluster")
		os.Exit(1)
	}

	if err = (&zoracontrollers.ClusterScanReconciler{
		Client:                  mgr.GetClient(),
		K8sClient:               kcli,
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("clusterscan-controller"),
		DefaultPluginsNamespace: defaultPluginsNamespace,
		DefaultPluginsNames:     strings.Split(defaultPluginsNames, ","),
		WorkerImage:             workerImage,
		ClusterRoleBindingName:  cronJobClusterRoleBinding,
		ServiceAccountName:      cronJobServiceAccount,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterScan")
		os.Exit(1)
	}

	if err = (&zoracontrollers.SaasReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		HttpCli:    &http.Client{},
		ID:         saasID,
		ServerAddr: saasServerAddr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Saas")
		os.Exit(1)
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

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
