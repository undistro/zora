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
	"time"

	"github.com/undistro/zora/worker/run"
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
	if err := run.Run(log); err != nil {
		log.Info("Worker crashed")
		panic(err)
	}
	log.Info("Worker finished successfully")
	log.Info("Stopping worker")
}
