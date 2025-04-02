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

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
)

func Test_createScraperRuntime_failsOnNoScheduledScopeEmitters(t *testing.T) {
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	scraperDescriptor := &Descriptor{
		Type:             testingScraper,
		ScopeDescriptors: make(map[string]scope.Descriptor),
	}
	enabledMetrics := &metric.Enabled{
		Metrics: map[string]*struct{}{
			"testing.metric.1": {},
		},
	}

	rt, err := createScraperRuntime(scraperDescriptor, enabledMetrics)

	assert.Error(t, err, "on empty scraper descriptor runtime creation fails")
	assert.ErrorContains(t, err, "no scheduled scope emitters for scraper 'testing_scraper'")
	assert.Nil(t, rt, "runtime must be nil")
}

func Test_createScraperRuntime_succeedsOnEnabledMetrics(t *testing.T) {
	tm1 := "testing.metric.1"
	tm2 := "testing.metric.2"
	tm3 := "testing.metric.3"
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	scraperDescriptor := &Descriptor{
		Type: testingScraper,
		ScopeDescriptors: map[string]scope.Descriptor{
			"scope1": {
				ScopeName: "scope1",
				MetricDescriptors: map[string]metric.Descriptor{
					tm1: {
						Create: CreateMetricEmitter1a,
					},
					tm2: {
						Create: CreateMetricEmitter1b,
					},
				},
			},
			"scope2": {
				ScopeName: "scope2",
				MetricDescriptors: map[string]metric.Descriptor{
					tm3: {
						Create: CreateMetricEmitter2a,
					},
				},
			},
		},
	}
	enabledMetrics := &metric.Enabled{
		Metrics: map[string]*struct{}{
			tm1: {},
			tm2: {},
		},
	}
	rt, err := createScraperRuntime(scraperDescriptor, enabledMetrics)

	assert.NoError(t, err, "creation must succeed")
	assert.NotNil(t, rt, "runtime must be provided")
	assert.Equal(t, 1, len(rt.ScopeEmitters), "number of scope emitters must be the same")
	_, found := rt.ScopeEmitters["scope1"]
	assert.True(t, found, "scope emitter for scope1 must exists")
	_, found = rt.ScopeEmitters["scope2"]
	assert.False(t, found, "scope emitter for scope2 must not exists")
}
