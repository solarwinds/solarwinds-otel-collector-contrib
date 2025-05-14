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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/assetscraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hardwareinventoryscraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
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
		},
	}
}

func createMetricsReceiver(
	ctx context.Context,
	settings receiver.Settings,
	config component.Config,
	metrics consumer.Metrics,
) (receiver.Metrics, error) {
	const logErrorInclude = ": %w"
	cfg := config.(*ReceiverConfig)

	// Way of creating receiver with multiple scrapers - here the single one is added
	scraperControllerOptions, err := createScraperControllerOptions(ctx, cfg)
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
		message := "Failed to create swohostmetrics receiver"
		zap.L().Error(message, zap.Error(err))
		return nil, fmt.Errorf(message+logErrorInclude, err)
	}

	return receiver, nil
}

func createScraperControllerOptions(
	ctx context.Context,
	receiverConfig *ReceiverConfig,
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
			scraper.Settings{},
			scraperConfig,
		)
		if err != nil {
			message := fmt.Sprintf("creating scraper %s failed", scraperName)
			zap.L().Error(message, zap.Error(err))
			return nil, fmt.Errorf(message+": %w", err)
		}

		ct, err := component.NewType(scraperName)
		if err != nil {
			return nil, fmt.Errorf("invalid scraper key name: %s", scraperName)
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
		zap.L().Error(message)
		return nil, errors.New(message)
	}

	return scraperFactory, nil
}
