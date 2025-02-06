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
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper/metrics/uptime"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/types"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

func Test_SpecificMetricIsEnabledByDefault(t *testing.T) {
	enabledMetric := uptime.MetricName

	sut := NewFactory()
	defaultConfig := sut.CreateDefaultConfig().(*types.ScraperConfig)
	require.Truef(t, defaultConfig.Metrics[enabledMetric].Enabled, enabledMetric+" is disabled by default, but should be enabled.")
}

func Test_ScraperIsSuccessfullyCreated(t *testing.T) {
	config := &types.ScraperConfig{
		Metrics: map[string]types.MetricSettingsConfig{
			uptime.MetricName: {Enabled: true},
		},
	}
	receiverConfig := receiver.Settings{}

	sut := NewFactory()
	scraper, err := sut.CreateScraper(context.TODO(), receiverConfig, config)

	require.NoErrorf(t, err, "Scraper should be created without any error")
	require.Equalf(t, component.MustNewType("hostinfo"), scraper.ID().Type(), "Scraper type should be 'hostinfo'")
}
