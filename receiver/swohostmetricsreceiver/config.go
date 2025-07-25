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
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

/*
Expected config example

swohostmetrics:
	collection_interval: <duration>
	scrapers:
		hostinfo:
			...
		another:
			...
*/

// ReceiverConfig defines SWO host metrics configuration.
type ReceiverConfig struct {
	// common receiver settings.
	scraperhelper.ControllerConfig `mapstructure:",squash"`

	// available scrapers for receiver.
	Scrapers map[string]component.Config `mapstructure:"-"`
}

var (
	_ component.Config    = (*ReceiverConfig)(nil) // Type check against Config
	_ confmap.Unmarshaler = (*ReceiverConfig)(nil) // Type check against Unmarshaller
)

// Unmarshal implements confmap.Unmarshaler.
func (receiverConfig *ReceiverConfig) Unmarshal(rawConfig *confmap.Conf) error {
	if receiverConfig == nil {
		return fmt.Errorf("receiverConfig function receiver is nil")
	}

	if rawConfig == nil {
		return fmt.Errorf("raw configuration object is nil")
	}

	const logErrorInclude = ": %w"
	// try to unmarshall raw config into receiver config
	err := rawConfig.Unmarshal(receiverConfig, confmap.WithIgnoreUnused())
	if err != nil {
		message := "Config unmarshalling failed"
		return fmt.Errorf(message+logErrorInclude, err)
	}

	// loading scrapers config section
	scrapersSectionConfigMap, err := rawConfig.Sub("scrapers")
	if err != nil {
		message := "Failed to fetch scrapers section from config"
		return fmt.Errorf(message+logErrorInclude, err)
	}

	// processing scrapers
	scraperMap := scrapersSectionConfigMap.ToStringMap()
	receiverConfig.Scrapers = make(map[string]component.Config, len(scraperMap))
	for scraperName := range scraperMap {
		scraperFactory, err := GetScraperFactory(scraperName)
		if err != nil {
			message := fmt.Sprintf("Scraper factory for scraper %s was not found", scraperName)
			return fmt.Errorf(message+logErrorInclude, err)
		}

		// loads scraper config with default values
		scraperConfig := scraperFactory.CreateDefaultConfig()
		// extracting scraper config from configuration map
		scraperSectionConfigMap, err := scrapersSectionConfigMap.Sub(scraperName)
		if err != nil {
			message := fmt.Sprintf("Scraper configuration for scraper %s can not be fetched", scraperName)
			return fmt.Errorf(message+logErrorInclude, err)
		}

		// unmarshal it into scraper configuration struct
		err = scraperSectionConfigMap.Unmarshal(scraperConfig, confmap.WithIgnoreUnused())
		if err != nil {
			message := fmt.Sprintf("Umarshalling config for scraper %s failed", scraperName)
			return fmt.Errorf(message+logErrorInclude, err)
		}

		// set up unmarshalled config for given scraper
		receiverConfig.Scrapers[scraperName] = scraperConfig
	}

	return nil
}
