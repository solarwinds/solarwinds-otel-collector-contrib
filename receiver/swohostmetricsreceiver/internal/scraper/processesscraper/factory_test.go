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

package processesscraper

import (
	"context"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/processesscraper/internal/metadata"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/scraper"
	"go.uber.org/zap"
)

func Test_SpecificMetricIsEnabledByDefault(t *testing.T) {
	sut := NewFactory()
	defaultConfig := sut.CreateDefaultConfig().(*metadata.MetricsBuilderConfig)
	require.Truef(t,
		defaultConfig.Metrics.SwoSystemProcessesCount.Enabled,
		"%s is disabled by default, but should be enabled.",
		metadata.MetricsInfo.SwoSystemProcessesCount.Name)
}

func Test_ScraperIsSuccessfullyCreated(t *testing.T) {
	config := &metadata.MetricsBuilderConfig{
		Metrics: metadata.MetricsConfig{
			SwoSystemProcessesCount: metadata.MetricConfig{Enabled: true},
		},
	}
	sConfig := scraper.Settings{
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
	}

	sut := NewFactory()
	_, err := sut.CreateMetrics(context.TODO(), sConfig, config)

	require.NoErrorf(t, err, "Scraper should be created without any error")
}
