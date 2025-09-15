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

package hardwareinventoryscraper

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hardwareinventoryscraper/internal/metadata"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hardwareinventoryscraper/metrics/cpu"
)

const (
	cpuScopeName = "otelcol/swohostmetricsreceiver/hardwareinventory/cpu"
)

type Scraper struct {
	scraper.Manager
	config *Config
}

var _ scraper.Scraper = (*Scraper)(nil)

func NewHardwareInventoryScraper(
	config *Config,
	logger *zap.Logger,
) (*Scraper, error) {
	descriptor := &scraper.Descriptor{
		Type: metadata.Type,
		ScopeDescriptors: map[string]scope.Descriptor{
			cpuScopeName: {
				ScopeName: cpuScopeName,
				MetricDescriptors: map[string]metric.Descriptor{
					cpu.Name: {Create: cpu.NewEmitter},
				},
			},
		},
	}

	managerConfig := &scraper.ManagerConfig{
		ScraperConfig:           &config.ScraperConfig,
		DelayedProcessingConfig: &config.DelayedProcessingConfig,
	}

	s := &Scraper{
		Manager: scraper.NewScraperManager(logger),
		config:  config,
	}

	if err := s.Init(descriptor, managerConfig); err != nil {
		return nil, err
	}

	return s, nil
}
