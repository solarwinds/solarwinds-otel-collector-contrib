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

package scope

import (
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/stretchr/testify/assert"
)

func createArtificialMetricEmitter() metric.Emitter {
	return metric.CreateMetricEmitterMockV2("testing.metric", 0, 0)
}

func Test_TraverseThroughScopeDescriptors_onNoMatchNoEmittersAreReturned(t *testing.T) {
	scopeDescriptors := map[string]Descriptor{
		"scope1": {
			ScopeName: "scope1",
			MetricDescriptors: map[string]metric.Descriptor{
				"scope1.metric.1": {
					Create: createArtificialMetricEmitter,
				},
			},
		},
	}
	enabledMetrics := &metric.Enabled{
		Metrics: map[string]*struct{}{
			"scope2.metric.2": {},
		},
	}
	ses := TraverseThroughScopeDescriptors(scopeDescriptors, enabledMetrics)

	assert.NotNil(t, ses, "even empty scope emitters but never nil")
	assert.Zero(t, len(ses), "no match no allocated emitters")
}

func Test_TraverseThroughScopeDescriptors_onMatchEmittersAreReturned(t *testing.T) {
	expectedEmitterScopeName := "scope1"
	scopeDescriptors := map[string]Descriptor{
		expectedEmitterScopeName: {
			ScopeName: expectedEmitterScopeName,
			MetricDescriptors: map[string]metric.Descriptor{
				"scope1.metric.1": {
					Create: createArtificialMetricEmitter,
				},
			},
		},
	}
	enabledMetrics := &metric.Enabled{
		Metrics: map[string]*struct{}{
			"scope1.metric.1": {},
		},
	}
	ses := TraverseThroughScopeDescriptors(scopeDescriptors, enabledMetrics)

	assert.Equal(t, 1, len(ses), "match must produce scope emitter")
	_, found := ses["scope1"]
	assert.True(t, found, "scope emitter for scope1 must exists")
}

func Test_TraverseThroughScopeDescriptors_whenCustomAllocatorIsRequiredItIsUsed(t *testing.T) {
	scopeDescriptors := map[string]Descriptor{
		"scope1": {
			ScopeName: "scope1",
			MetricDescriptors: map[string]metric.Descriptor{
				"scope1.metric.1": {
					Create: createArtificialMetricEmitter,
				},
			},
			// Custom allocator is used. Produces special named scope emitter.
			Create: createArtificialScopeEmitter,
		},
	}
	enabledMetrics := &metric.Enabled{
		Metrics: map[string]*struct{}{
			"scope1.metric.1": {},
		},
	}
	ses := TraverseThroughScopeDescriptors(scopeDescriptors, enabledMetrics)

	assert.Equal(t, 1, len(ses), "match must produce scope emitter")
	_, found := ses["testing/scope"]
	assert.True(t, found, "emitter with overiden scope name must exists")
}

func createArtificialScopeEmitter(string, map[string]metric.Emitter) Emitter {
	return CreateCustomScopeEmitter("testing/scope", make(map[string]metric.Emitter))
}
