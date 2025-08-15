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

package swohostmetricsreceiver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/assetscraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hardwareinventoryscraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/processesscraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

const (
	stability = component.StabilityLevelDevelopment
)

//nolint:gochecknoglobals // Private, read-only.
var componentType component.Type

func init() {
	componentType = component.MustNewType("swohostmetrics")
}

func ComponentType() component.Type {
	return componentType
}

func scraperFactories() map[string]types.MetricsScraperFactory {
	return map[string]types.MetricsScraperFactory{
		assetscraper.ScraperType().String():             assetscraper.NewFactory(),
		hardwareinventoryscraper.ScraperType().String(): hardwareinventoryscraper.NewFactory(),
		hostinfoscraper.ScraperType().String():          hostinfoscraper.NewFactory(),
		processesscraper.NewFactory().Type().String():   processesscraper.NewFactory(),
	}
}

// Creates factory capable of creating swohostmetrics receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		ComponentType(),
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, stability),
	)
}

func createDefaultConfig() component.Config {
	return &ReceiverConfig{
		ControllerConfig: scraperhelper.ControllerConfig{
			CollectionInterval: 30 * time.Second,
		},
		Scrapers: map[string]component.Config{
			hostinfoscraper.ScraperType().String():          hostinfoscraper.CreateDefaultConfig(),
			assetscraper.ScraperType().String():             assetscraper.CreateDefaultConfig(),
			hardwareinventoryscraper.ScraperType().String(): hardwareinventoryscraper.CreateDefaultConfig(),
			processesscraper.NewFactory().Type().String():   processesscraper.NewFactory().CreateDefaultConfig(),
		},
	}
}

func createMetricsReceiver(
	ctx context.Context,
	settings receiver.Settings,
	config component.Config,
	metrics consumer.Metrics,
) (receiver.Metrics, error) {
	cfg := config.(*ReceiverConfig)

	// Way of creating receiver with multiple scrapers - here the single one is added
	scraperControllerOptions, err := createScraperControllerOptions(ctx, cfg, settings)
	if err != nil {
		return nil, err
	}

	receiver, err := scraperhelper.NewMetricsController(
		&cfg.ControllerConfig,
		settings,
		metrics,
		scraperControllerOptions...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create swohostmetrics receiver: %w", err)
	}

	return receiver, nil
}

func createScraperControllerOptions(
	ctx context.Context,
	receiverConfig *ReceiverConfig,
	settings receiver.Settings,
) ([]scraperhelper.ControllerOption, error) {
	scraperFactories := scraperFactories()
	scraperControllerOptions := make([]scraperhelper.ControllerOption, 0, len(scraperFactories))

	for scraperName, scraperFactory := range scraperFactories {
		// when config is not available it is not utilized in receiver
		// => skip it
		scraperConfig, found := receiverConfig.Scrapers[scraperName]
		if !found {
			continue
		}

		scraper, err := scraperFactory.CreateMetrics(
			ctx,
			scraper.Settings{
				ID:                component.NewID(scraperFactory.Type()),
				TelemetrySettings: settings.TelemetrySettings,
				BuildInfo:         settings.BuildInfo,
			},
			scraperConfig,
		)
		if err != nil {
			return nil, fmt.Errorf("creating scraper %s failed: %w", scraperName, err)
		}

		ct, err := component.NewType(scraperName)
		if err != nil {
			return nil, fmt.Errorf("invalid scraper key name: %s : %w", scraperName, err)
		}

		scraperControllerOptions = append(scraperControllerOptions, scraperhelper.AddScraper(ct, scraper))
	}

	return scraperControllerOptions, nil
}

// returns scraper factory for its creation or error if no such scraper can be
// provided.
func GetScraperFactory(scraperName string) (types.MetricsScraperFactory, error) {
	scraperFactory, found := scraperFactories()[scraperName]
	if !found {
		message := fmt.Sprintf("Scraper [%s] is unknown", scraperName)
		return nil, errors.New(message)
	}

	return scraperFactory, nil
}
