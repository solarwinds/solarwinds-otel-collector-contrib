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
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
)

func NewMetricEmitterS1M1() metric.Emitter {
	return metric.CreateMetricEmitterMockV2(scope1metric1, 1, 101)
}

func NewMetricEmitterS1M2() metric.Emitter {
	return metric.CreateMetricEmitterMockV2(scope1metric2, 2, 102)
}

func NewMetricEmitterS2M1() metric.Emitter {
	return metric.CreateMetricEmitterMockV2(scope2metric1, 1, 201)
}
