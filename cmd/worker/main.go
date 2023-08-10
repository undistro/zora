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
	"context"
	"flag"
	"os"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/undistro/zora/pkg/worker"
)

func main() {
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.TimeEncoderOfLayout(time.RFC3339),
	}
	opts.BindFlags(flag.CommandLine)
	log := zap.New(zap.UseFlagOptions(&opts)).WithName("worker")
	ctx := logr.NewContext(context.Background(), log)

	log.Info("starting worker")
	if err := worker.Run(ctx); err != nil {
		log.Error(err, "failed to run worker")
		os.Exit(1)
	}
	log.Info("worker finished successfully")
}
