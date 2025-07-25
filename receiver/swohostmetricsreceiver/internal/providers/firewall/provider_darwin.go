// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
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

package firewall

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"go.uber.org/zap"
)

type provider struct {
	logger *zap.Logger
}

var _ (providers.Provider[Container]) = (*provider)(nil)

func CreateFirewallProvider(logger *zap.Logger) providers.Provider[Container] {
	return &provider{
		logger: logger,
	}
}

// Provide implements providers.Provider.
func (p *provider) Provide() <-chan Container {
	ch := make(chan Container)
	go func() {
		defer close(ch)
		p.logger.Warn("This provider is not supported on Darwin")
	}()
	return ch
}
