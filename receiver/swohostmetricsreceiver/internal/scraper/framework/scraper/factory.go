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

package scraper

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/scraper"
	"go.uber.org/zap"
)

// CreateScraper creates scraper in implicit way. Function packs
// all required checks and allocation into single call to be minimalistic
// in usage.
func CreateScraper[TConfig component.Config, TScraper Scraper](
	scraperName component.Type,
	config component.Config,
	sAllocator func(*TConfig) (*TScraper, error),
) (scraper.Metrics, error) {
	sc, err := sAllocator(config.(*TConfig))
	if err != nil {
		m := fmt.Sprintf("scraper '%s' creation failed", scraperName)
		zap.L().Error(m, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", m, err)
	}

	otelScraper, err := scraper.NewMetrics(
		(*sc).Scrape,
		scraper.WithStart((*sc).Start),
		scraper.WithShutdown((*sc).Shutdown),
	)
	if err != nil {
		m := fmt.Sprintf("new metrics scraper '%s' creation failed", scraperName)
		zap.L().Error(m, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", m, err)
	}

	return otelScraper, nil
}
