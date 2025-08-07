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

package example

import (
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	frameworkscraper "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
)

const (
	// Scope name definitions.
	scope1 = "otelcol/swohostmetricsreceiver/exemplary_scraper/scope1"
	scope2 = "otelcol/swohostmetricsreceiver/exemplary_scraper/scope2"

	// Metric name definitions for scope 1.
	scope1metric1 = "swo.exemplary_scraper.scope1.metric1"
	scope1metric2 = "swo.exemplary_scraper.scope1.metric2"
	scope1metric3 = "swo.exemplary_scraper.scope1.metric3"

	// Metric name definitions for scope 2.
	scope2metric1 = "swo.exemplary_scraper.scope2.metric1"
)

//nolint:gochecknoglobals // Private, read-only.
var scraperType component.Type = component.MustNewType("exemplary_scraper")

func ScraperType() component.Type {
	return scraperType
}

// ExemplaryScraper exemplifies a typical use of
// the scraper framework.
type ExemplaryScraper struct {
	frameworkscraper.Manager
	config *ScraperConfig
}

var _ frameworkscraper.Scraper = (*ExemplaryScraper)(nil)

func NewExemplaryScraper(
	config *ScraperConfig,
	logger *zap.Logger,
) (*ExemplaryScraper, error) {
	s := &ExemplaryScraper{
		Manager: frameworkscraper.NewScraperManager(logger),
		config:  config,
	}

	// Scraper abilities. Declarative description of scopes and related metrics.
	scraperDescription := &frameworkscraper.Descriptor{
		Type: ScraperType(),
		ScopeDescriptors: map[string]scope.Descriptor{
			// Emits scope otelcol/swohostmetricsreceiver/exemplary-scraper/scope1
			scope1: {
				ScopeName: scope1,
				MetricDescriptors: map[string]metric.Descriptor{
					// Emits metric swo.exemplary-scraper.scope1.metric1.
					scope1metric1: {Create: NewMetricEmitterS1M1},
					// Emits metric swo.exemplary-scraper.scope1.metric2
					scope1metric2: {Create: NewMetricEmitterS1M2},
				},
			},
			// Emits scope otelcol/swohostmetricsreceiver/exemplary-scraper/scope1
			scope2: {
				ScopeName: scope2,
				MetricDescriptors: map[string]metric.Descriptor{
					// Emits metric swo.exemplary-scraper.scope2.metric1
					scope2metric1: {Create: NewMetricEmitterS2M1},
				},
			},
		},
	}

	// Translation from scraper specific config.
	smConfig := &frameworkscraper.ManagerConfig{
		ScraperConfig:           &config.ScraperConfig,
		DelayedProcessingConfig: &config.DelayedProcessingConfig,
	}

	err := s.Init(scraperDescription, smConfig)
	if err != nil {
		return nil, err
	}

	return s, nil
}
