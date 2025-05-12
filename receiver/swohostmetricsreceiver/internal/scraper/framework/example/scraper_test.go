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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	frameworkscraper "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
)

func Test_Example_HowToConfigureScraper(t *testing.T) {
	scraperType := scraper.RequireComponentType(t, "artificialScraper")

	// Describes one scraper.
	scraperDefinition := &frameworkscraper.Descriptor{
		Type: scraperType,
		ScopeDescriptors: map[string]scope.Descriptor{
			scope1: {
				ScopeName: scope1,
				MetricDescriptors: map[string]metric.Descriptor{
					scope1metric1: {
						Create: frameworkscraper.CreateMetricEmitter1a,
					},
					scope1metric2: {
						Create: frameworkscraper.CreateMetricEmitter1b,
					},
					scope1metric3: {
						Create: frameworkscraper.CreateMetricEmitter1c,
					},
				},
				// Without custom scope alocator - most of the cases.
				// Default one is expect to be taken.
			},
			scope2: {
				ScopeName: scope2,
				// Allocator callback for creation of custom scope emitter
				// for this particullar scope - scope2.
				Create: scope.CreateCustomScopeEmitter,
				MetricDescriptors: map[string]metric.Descriptor{
					scope2metric1: {
						Create: frameworkscraper.CreateMetricEmitter2a,
					},
				},
			},
		},
	}

	assert.Equal(t, "artificialScraper", scraperDefinition.Type.String(), "Scraper must have a name")
	assert.Equal(t, 2, len(scraperDefinition.ScopeDescriptors), "scraper must have two scopes")
	assert.Equal(t,
		3,
		len(scraperDefinition.ScopeDescriptors[scope1].MetricDescriptors),
		"scope scope1 must have three metrics",
	)
	assert.Nil(t, scraperDefinition.ScopeDescriptors[scope1].Create,
		"scope scope1 must have no custom allocator",
	)
	assert.Equal(t,
		1,
		len(scraperDefinition.ScopeDescriptors[scope2].MetricDescriptors),
		"scope scope2 must have one metric",
	)
	assert.NotNil(t, scraperDefinition.ScopeDescriptors[scope2].Create,
		"scope scope2 must have no custom allocator",
	)
}

func Test_OverallScraperTest(t *testing.T) {
	ctx := context.Background()

	// Scraper configuration coming from outside - from otel collector.
	// Contains enabled/disabled metrics based on YAML config.
	usedConfig := &ScraperConfig{
		DelayedProcessingConfig: types.DelayedProcessingConfig{},
		ScraperConfig: types.ScraperConfig{
			Metrics: map[string]types.MetricSettingsConfig{
				scope1metric1: {Enabled: true},
				scope1metric2: {Enabled: false},
				scope2metric1: {Enabled: true},
			},
		},
	}

	// Creating scraper in factory.
	sut, err := CreateScraperImplicitly(ctx, receiver.Settings{}, usedConfig)
	assert.Nil(t, err, "creating scraper must be errorless")

	// Initializes scraper.
	err = sut.Start(ctx, nil)
	assert.Nil(t, err, "starting scraper must be errorless")

	// Scraping.
	metrics, err := sut.ScrapeMetrics(ctx)
	assert.Nil(t, err, "scraping must be errorless")

	// Datapoints: 2, metrics: 2
	assert.Equal(t, 2, metrics.DataPointCount(), "2 datapoints must be available")
	assert.Equal(t, 2, metrics.MetricCount(), "2 metrics must be available")

	sms := metrics.ResourceMetrics().At(0).ScopeMetrics()
	assert.Equal(t, 2, sms.Len(), "metrics must have 2 scopes")

	// Since processing is fully parallel, we can not be sure in which order metric are
	// composed.
	evaluateScopeMetrics(t, sms.At(0))
	evaluateScopeMetrics(t, sms.At(1))
}

func evaluateScopeMetrics(t *testing.T, sm pmetric.ScopeMetrics) {
	switch sm.Scope().Name() {
	case scope1:
		evaluateScope1Metrics(t, sm.Metrics())
	case scope2:
		evaluateScope2Metrics(t, sm.Metrics())
	default:
		m := fmt.Sprintf("there is a scope '%s', which is not expected", sm.Scope().Name())
		assert.Fail(t, m)
	}
}

func evaluateScope1Metrics(t *testing.T, ms pmetric.MetricSlice) {
	assert.Equal(t, 1, ms.Len(), "there must be only one metric")
	assert.Equal(t, scope1metric1, ms.At(0).Name(), "name must the same")
	assert.Equal(t, 1, ms.At(0).Sum().DataPoints().Len(), "there must be only one data point")
}

func evaluateScope2Metrics(t *testing.T, ms pmetric.MetricSlice) {
	assert.Equal(t, 1, ms.Len(), "there must be only one metric")
	assert.Equal(t, scope2metric1, ms.At(0).Name(), "name must the same")
	assert.Equal(t, 1, ms.At(0).Sum().DataPoints().Len(), "there must be only one data point")
}
