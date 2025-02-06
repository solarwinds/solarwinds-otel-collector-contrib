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
	"go.opentelemetry.io/collector/component"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/hardwareinventoryscraper/metrics/cpu"
)

const (
	cpuScopeName = "otelcol/swohostmetricsreceiver/hardwareinventory/cpu"
)

//nolint:gochecknoglobals // Private, read-only.
var scraperType component.Type = component.MustNewType("hardwareinventory")

func ScraperType() component.Type {
	return scraperType
}

type Scraper struct {
	scraper.Manager
	config *Config
}

var _ scraper.Scraper = (*Scraper)(nil)

func NewHardwareInventoryScraper(
	config *Config,
) (*Scraper, error) {
	descriptor := &scraper.Descriptor{
		Type: ScraperType(),
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
		Manager: scraper.NewScraperManager(),
		config:  config,
	}

	if err := s.Init(descriptor, managerConfig); err != nil {
		return nil, err
	}

	return s, nil
}
