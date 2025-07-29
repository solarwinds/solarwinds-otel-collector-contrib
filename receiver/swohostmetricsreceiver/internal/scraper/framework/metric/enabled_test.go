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

package metric

import (
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"github.com/stretchr/testify/assert"
)

const scraperName = "testing_scraper"

func Test_GetEnabledMetrics_failsOnEmptyScraperConfig(t *testing.T) {
	config := &types.ScraperConfig{}
	en, err := GetEnabledMetrics(scraperName, config)

	assert.Error(t, err, "on empty scraper config must fail")
	assert.ErrorContains(t, err, "no configured metrics for scraper 'testing_scraper'")
	assert.Nil(t, en, "enabled metrics must ne nil on failure")
}

func Test_GetEnabledMetrics_onlyEnabledMetricsAreReturned(t *testing.T) {
	expectedEnabledMetric := "testing.metric.1"
	config := &types.ScraperConfig{
		Metrics: map[string]types.MetricSettingsConfig{
			expectedEnabledMetric: {
				Enabled: true,
			},
			"testing.metric.2": {
				Enabled: false,
			},
		},
	}
	en, err := GetEnabledMetrics(scraperName, config)

	assert.NoError(t, err, "GetEnabledMetrics call must not fail")
	assert.NotNil(t, en, "enabled metrics must exists")
	assert.Equal(t, 1, len(en.Metrics), "there must be exactly one metric")
	_, found := en.Metrics[expectedEnabledMetric]
	assert.True(t, found, "desired metric must be found among enabled ones")
}

func Test_GetEnabledMetrics_failsOnNoEnabledMetrics(t *testing.T) {
	config := &types.ScraperConfig{
		Metrics: map[string]types.MetricSettingsConfig{
			"testing.metric.1": {
				Enabled: false,
			},
			"testing.metric.2": {
				Enabled: false,
			},
		},
	}
	en, err := GetEnabledMetrics(scraperName, config)

	assert.Error(t, err, "with no enabled metric call must fail")
	assert.Nil(t, en, "enable metrics must not exists")
}
