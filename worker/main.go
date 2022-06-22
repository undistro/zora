package main

import (
	"time"

	"github.com/getupio-undistro/zora/worker/run"
	"go.uber.org/zap/zapcore"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var log = ctrl.Log.WithName("worker")

func main() {
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.TimeEncoderOfLayout(time.RFC3339),
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	log.Info("Starting worker")
	if err := run.Run(); err != nil {
		log.Info("Worker crashed")
		panic(err)
	}
	log.Info("Worker finished successfully")
	log.Info("Stopping worker")
}
