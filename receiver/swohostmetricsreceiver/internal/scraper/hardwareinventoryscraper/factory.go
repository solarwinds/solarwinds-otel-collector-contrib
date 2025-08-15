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
	"context"

	fscraper "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/scraper"
)

type factory struct{}

var _ types.MetricsScraperFactory = (*factory)(nil)

func NewFactory() types.MetricsScraperFactory {
	return new(factory)
}

// Type implements types.MetricsScraperFactory.
func (f *factory) Type() component.Type {
	return scraperType
}

// CreateDefaultConfig implements types.ScraperFactory.
func (*factory) CreateDefaultConfig() component.Config {
	return CreateDefaultConfig()
}

// CreateScraper implements types.ScraperFactory.
func (*factory) CreateMetrics(
	_ context.Context,
	set scraper.Settings,
	cfg component.Config,
) (scraper.Metrics, error) {
	return fscraper.CreateScraper[Config, Scraper](
		ScraperType(),
		cfg,
		NewHardwareInventoryScraper,
		set.Logger,
	)
}
