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

import "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"

const (
	// Scraper name definition.
	scraperName = "test_scraper"

	// Scope name definitions.
	scope1 = "otelcol/swohostmetricsreceiver/test_scraper/scope1"
	scope2 = "otelcol/swohostmetricsreceiver/test_scraper/scope2"

	// Metric name definitions for scope 1.
	scope1metric1 = "swo.test_scraper.scope1.metric1"
	scope1metric2 = "swo.test_scraper.scope1.metric2"
	scope1metric3 = "swo.test_scraper.scope1.metric3"

	// Metric name definitions for scope 2.
	scope2metric1 = "swo.test_scraper.scope2.metric1"
	scope2metric2 = "swo.test_scraper.scope2.metric2"
)

func CreateMetricEmitter1a() metric.Emitter {
	return metric.CreateMetricEmitterMockV2(scope1metric1, 1, 101)
}

func CreateMetricEmitter1b() metric.Emitter {
	return metric.CreateMetricEmitterMockV2(scope1metric2, 2, 102)
}

func CreateMetricEmitter1c() metric.Emitter {
	return metric.CreateMetricEmitterMockV2(scope1metric3, 3, 103)
}

func CreateMetricEmitter2a() metric.Emitter {
	return metric.CreateMetricEmitterMockV2(scope2metric1, 2, 201)
}

func CreateMetricEmitter2b() metric.Emitter {
	return metric.CreateMetricEmitterMockV2(scope2metric2, 4, 202)
}
