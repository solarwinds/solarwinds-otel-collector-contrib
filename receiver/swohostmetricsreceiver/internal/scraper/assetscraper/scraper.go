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

package assetscraper

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/assetscraper/metrics/installedsoftware"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/assetscraper/metrics/installedupdates"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
	"go.uber.org/zap"
)

const (
	scopeMetricsName = "otelcol/swohostmetricsreceiver/asset"
)

type AssetScraper struct {
	scraper.Manager
	config *Config
}

var _ scraper.Scraper = (*AssetScraper)(nil)

func NewAssetScraper(
	config *Config,
	logger *zap.Logger,
) (*AssetScraper, error) {
	descriptor := &scraper.Descriptor{
		Type: ScraperType(),
		ScopeDescriptors: map[string]scope.Descriptor{
			scopeMetricsName: {
				ScopeName: scopeMetricsName,
				MetricDescriptors: map[string]metric.Descriptor{
					installedsoftware.Name: {Create: installedsoftware.NewEmitter},
					installedupdates.Name:  {Create: installedupdates.NewEmitter},
				},
			},
		},
	}

	managerConfig := &scraper.ManagerConfig{
		ScraperConfig:           &config.ScraperConfig,
		DelayedProcessingConfig: &config.DelayedProcessingConfig,
	}

	s := &AssetScraper{
		Manager: scraper.NewScraperManager(logger),
		config:  config,
	}

	if err := s.Init(descriptor, managerConfig); err != nil {
		return nil, err
	}

	return s, nil
}
