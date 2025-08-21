// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package swok8sworkloadstatusprocessor // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadstatusprocessor"

import (
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	k8sconfig.APIConfig `mapstructure:",squash"`
	WatchSyncPeriod     time.Duration `mapstructure:"watch_sync_period"`
}

func (c *Config) getK8sClient() (kubernetes.Interface, error) {
	return initKubeClient(c.APIConfig)
}
