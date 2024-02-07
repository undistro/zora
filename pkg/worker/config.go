// Copyright 2023 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worker

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v9"
)

// config represents worker configuration
type config struct {
	DoneFile     string        `env:"DONE_FILE" envDefault:"/tmp/zora/results/done"`
	ErrorFile    string        `env:"ERROR_FILE" envDefault:"/tmp/zora/results/error"`
	PluginName   string        `env:"PLUGIN_NAME,required"`
	PluginType   string        `env:"PLUGIN_TYPE,required"`
	ClusterName  string        `env:"CLUSTER_NAME,required"`
	ClusterUID   string        `env:"CLUSTER_UID,required"`
	Namespace    string        `env:"NAMESPACE,required"`
	JobName      string        `env:"JOB_NAME,required"`
	JobUID       string        `env:"JOB_UID,required"`
	PodName      string        `env:"POD_NAME,required"`
	WaitInterval time.Duration `env:"WAIT_INTERVAL" envDefault:"1s"`

	suffix string
}

// configFromEnv returns a config from environment variables
func configFromEnv() (*config, error) {
	cfg := &config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if i := strings.LastIndex(cfg.PodName, "-"); i > 0 && len(cfg.PodName) > i+1 {
		cfg.suffix = cfg.PodName[i+1:]
	}
	return cfg, nil
}
