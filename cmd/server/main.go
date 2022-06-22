package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/go-chi/chi/v5"
	"github.com/ory/graceful"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/getupio-undistro/zora/handlers"
	"github.com/getupio-undistro/zora/pkg/clientset/versioned"
)

var log = ctrl.Log.WithName("server")

func main() {
	var port int
	flag.IntVar(&port, "port", 3000, "server port")
	opts := zap.Options{Development: true, TimeEncoder: zapcore.TimeEncoderOfLayout(time.RFC3339)}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	config := ctrl.GetConfigOrDie()
	client, err := versioned.NewForConfig(config)
	if err != nil {
		log.Error(err, "failed to create Clusters clientset")
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Get("/api/v1/clusters", handlers.ClusterListHandler(client, logger))
	r.Get("/api/v1/issues", handlers.IssueListHandler(client, logger))
	r.Get("/api/v1/namespaces/{namespace}/clusters/{clusterName}", handlers.ClusterHandler(client, logger))
	r.Get("/health", handlers.Health)

	server := graceful.WithDefaults(&http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: r,
	})

	log.Info(fmt.Sprintf("starting the server at %s", server.Addr))
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		log.Error(err, "failed to gracefully shutdown")
		os.Exit(1)
	}
	log.Info("server was shutdown gracefully")
}
