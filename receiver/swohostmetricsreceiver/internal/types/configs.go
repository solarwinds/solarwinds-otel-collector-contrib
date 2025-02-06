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

package types

/*
Expected config example

swohostmetrics:
	collection_interval: <duration>
	root_path: <string>
	scrapers:
		hostinfo:
			metrics:
				swo.hostinfo.uptime:
					enabled: true
				swo.hostinfo.whatever:
					enabled: false
		another:
			metrics:
				swo.another.whatever:
					enabled:true
		...
*/

// Scraper metrics configuration.
type MetricSettingsConfig struct {
	// flag which indicates if metric is enabled or disabled for scraping
	Enabled bool `mapstructure:"enabled"`
}

// Scraper configuration.
type ScraperConfig struct {
	// scraper provides metrics scraping.
	Metrics map[string]MetricSettingsConfig `mapstructure:"metrics"`
}

// CreateDefaultScraperConfig creates scraper config usable as general
// configuration struct for common/usual use.
func CreateDefaultScraperConfig() *ScraperConfig {
	return &ScraperConfig{
		Metrics: map[string]MetricSettingsConfig{},
	}
}
