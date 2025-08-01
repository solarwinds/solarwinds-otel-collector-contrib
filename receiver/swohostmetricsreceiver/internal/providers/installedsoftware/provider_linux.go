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

package installedsoftware

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/installedsoftware/discovery"
	"go.uber.org/zap"
)

type linuxProvider struct {
	logger *zap.Logger
}

var _ Provider = (*linuxProvider)(nil)

func NewInstalledSoftwareProvider(logger *zap.Logger) Provider {
	return createInstalledSoftwareProvider(
		map[string]linuxInstalledSoftwareContainer{
			"rpm": {
				Discovery: discovery.NewRpmDiscovery(),
				Provider:  NewRpmProvider(logger),
			},
			"dpkg": {
				Discovery: discovery.NewDpkgDiscovery(),
				Provider:  NewDpkgProvider(logger),
			},
		},
		getDefaultProvider(logger),
		logger,
	)
}

func createInstalledSoftwareProvider(
	discoverableProviders map[string]linuxInstalledSoftwareContainer,
	fallbackProvider Provider,
	logger *zap.Logger,
) Provider {
	provider := discoverProvider(
		discoverableProviders,
		fallbackProvider,
		logger,
	)
	return provider
}

func getDefaultProvider(logger *zap.Logger) Provider {
	return &linuxProvider{
		logger: logger,
	}
}

// GetSoftware implements Provider.
func (p *linuxProvider) GetSoftware() ([]InstalledSoftware, error) {
	p.logger.Debug("unable to provide installed software via linuxProvider")
	return make([]InstalledSoftware, 0), nil
}

type linuxInstalledSoftwareContainer struct {
	Discovery discovery.Discovery
	Provider  Provider
}

func discoverProvider(
	discoverableProviders map[string]linuxInstalledSoftwareContainer,
	fallbackProvider Provider,
	logger *zap.Logger,
) Provider {
	// go through providers and select the most prioritized one
	for _, container := range discoverableProviders {
		// discovery successful => use its provider
		if container.Discovery.Discover() {
			return container.Provider
		}
	}

	logger.Warn("default installed software provider for linux will be used")
	return fallbackProvider
}
