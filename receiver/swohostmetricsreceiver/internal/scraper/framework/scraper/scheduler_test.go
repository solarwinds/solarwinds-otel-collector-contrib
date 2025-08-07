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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"

	"go.uber.org/zap"
)

func Test_HowSchedulerWorks(t *testing.T) {
	// Scraper config.
	scraperConfig := &types.ScraperConfig{
		Metrics: map[string]types.MetricSettingsConfig{
			scope1metric1: {Enabled: true},
			// Will be removed from final runtime => not enabled.
			scope1metric2: {Enabled: false},
			// Will be removed from final runtime => not enabled.
			scope1metric3: {Enabled: false},
			scope2metric1: {Enabled: true},
			scope2metric2: {Enabled: true},
		},
	}

	testingScraper := scraper.RequireComponentType(t, scraperName)

	scraperDefinition := &Descriptor{
		Type: testingScraper,
		ScopeDescriptors: map[string]scope.Descriptor{
			scope1: {
				ScopeName: scope1,
				MetricDescriptors: map[string]metric.Descriptor{
					scope1metric1: {
						Create: CreateMetricEmitter1a,
					},
					scope1metric2: {
						Create: CreateMetricEmitter1b,
					},
					scope1metric3: {
						Create: CreateMetricEmitter1c,
					},
				},
				// Without custom scope allocator - most of the cases.
				// Default one is expect to be taken.
			},
			scope2: {
				ScopeName: scope2,
				// Allocator callback for creation of custom scope emitter
				// for this particular scope - scope2. In case we need to
				// create something with totally custom implementation.
				Create: scope.CreateCustomScopeEmitter,
				MetricDescriptors: map[string]metric.Descriptor{
					scope2metric1: {
						Create: CreateMetricEmitter2a,
					},
					scope2metric2: {
						Create: CreateMetricEmitter2b,
					},
				},
			},
		},
	}

	sut := NewScraperScheduler()
	runtime, err := sut.Schedule(scraperDefinition, scraperConfig, zap.NewNop())

	assert.Nil(t, err, "functionality must be errorless")
	assert.Equal(t, 2, len(runtime.ScopeEmitters), "there must be two scpe emitters ready for use")
}
