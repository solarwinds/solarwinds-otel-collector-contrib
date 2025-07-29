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
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.uber.org/zap"

	fscraper "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
)

// CreateScraperExplicitely represents the way how to create scraper with
// a possibility to run some additional code.
func CreateScraperExplicitly(
	_ context.Context,
	_ receiver.Settings,
	cfg component.Config,
	logger *zap.Logger,
) (scraper.Metrics, error) {
	// Create scraper directly through allocating callback.
	exampleScraper, err := NewExemplaryScraper(cfg.(*ScraperConfig), logger)
	if err != nil {
		return nil, fmt.Errorf("scraper '%s' creation failed: %w", ScraperType(), err)
	}

	// In case there is a need to add some more code to be run.
	//
	// Create an OTEL scraper.
	otelScraper, err := scraper.NewMetrics(
		exampleScraper.Scrape,
		scraper.WithStart(exampleScraper.Start),
		scraper.WithShutdown(exampleScraper.Shutdown),
	)

	return otelScraper, err
}

func CreateScraperImplicitly(
	_ context.Context,
	_ receiver.Settings,
	cfg component.Config,
	logger *zap.Logger,
) (scraper.Metrics, error) {
	return fscraper.CreateScraper[ScraperConfig, ExemplaryScraper](
		ScraperType(),
		cfg,
		NewExemplaryScraper,
		logger,
	)
}
