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

package processesscraper

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	fscraper "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/processesscraper/internal/metadata"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/processesscraper/metrics/count"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"go.opentelemetry.io/collector/scraper"
	"go.uber.org/zap"
)

type Scraper struct {
	fscraper.Manager
	config *types.ScraperConfig
}

var _ fscraper.Scraper = (*Scraper)(nil)

func NewScraper(
	scraperConfig *types.ScraperConfig,
	settings scraper.Settings,
) (*Scraper, error) {
	config := ToMetadataConfig(scraperConfig)
	mb := metadata.NewMetricsBuilder(config, settings)
	descriptor := &fscraper.Descriptor{
		Type: metadata.Type,
		ScopeDescriptors: map[string]scope.Descriptor{
			metadata.ScopeName: {
				ScopeName: metadata.ScopeName,
				MetricDescriptors: map[string]metric.Descriptor{
					metadata.MetricsInfo.SwoSystemProcessesCount.Name: {
						Create: func(*zap.Logger) metric.Emitter { return count.NewEmitter(mb, settings.Logger) }},
				},
			},
		},
	}

	managerConfig := &fscraper.ManagerConfig{ScraperConfig: scraperConfig}

	s := &Scraper{
		Manager: fscraper.NewScraperManager(settings.Logger),
		config:  scraperConfig,
	}

	if err := s.Init(descriptor, managerConfig); err != nil {
		return nil, err
	}

	return s, nil
}
