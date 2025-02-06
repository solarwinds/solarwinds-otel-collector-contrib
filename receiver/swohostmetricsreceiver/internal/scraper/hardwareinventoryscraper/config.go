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
	"time"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/hardwareinventoryscraper/metrics/cpu"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/types"
	"go.opentelemetry.io/collector/component"
)

// Config for Hardware Inventory scraper.
type Config struct {
	types.DelayedProcessingConfig `mapstructure:",squash"`
	types.ScraperConfig           `mapstructure:",squash"`
}

// Config implements component.Config interface.
var _ component.Config = (*Config)(nil)

func CreateDefaultConfig() component.Config {
	return &Config{
		DelayedProcessingConfig: types.DelayedProcessingConfig{
			CollectionInterval: 90 * time.Second,
		},
		ScraperConfig: types.ScraperConfig{
			Metrics: map[string]types.MetricSettingsConfig{
				cpu.Name: {
					Enabled: true,
				},
			},
		},
	}
}
