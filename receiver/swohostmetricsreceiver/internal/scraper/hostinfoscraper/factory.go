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

package hostinfoscraper

import (
	"context"

	fscraper "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper/internal/metadata"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/scraper"
)

func NewFactory() scraper.Factory {
	return scraper.NewFactory(
		metadata.Type,
		CreateDefaultConfig,
		scraper.WithMetrics(createMetrics, metadata.MetricsStability))
}

func createMetrics(
	_ context.Context,
	set scraper.Settings,
	cfg component.Config,
) (scraper.Metrics, error) {
	return fscraper.CreateScraper[types.ScraperConfig, Scraper](
		metadata.Type,
		cfg,
		NewHostInfoScraper,
		set.Logger,
	)
}
