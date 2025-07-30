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

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func Test_TraverseThroughMetricDescriptors_onNoMatchEmptyEmitterMapIsReturned(t *testing.T) {
	descriptors := map[string]Descriptor{
		"testing.metric.1": {
			Create: createArtificialMetricEmitter,
		},
	}
	enabledMetrics := &Enabled{
		Metrics: map[string]*struct{}{
			"testing.metric.2": {},
		},
	}
	mes := TraverseThroughMetricDescriptors(descriptors, enabledMetrics, zap.NewNop())

	assert.Zero(t, len(mes), "on no match no emitters are created")
}

func Test_TraverseThroughMetricDescriptors_onMatchOnlyMetricEmittersAreCreated(t *testing.T) {
	expectedMetricName := "testing.metric.1"
	descriptors := map[string]Descriptor{
		expectedMetricName: {
			Create: createArtificialMetricEmitter,
		},
	}
	enabledMetrics := &Enabled{
		Metrics: map[string]*struct{}{
			expectedMetricName: {},
		},
	}
	mes := TraverseThroughMetricDescriptors(descriptors, enabledMetrics, zap.NewNop())

	assert.Equal(t, 1, len(mes), "on match metric emitter is created")
	_, found := mes[expectedMetricName]
	assert.True(t, found, "emitter for matching metric must be created")
}

func createArtificialMetricEmitter(_ *zap.Logger) Emitter {
	return CreateMetricEmitterMockV2("testing.metric.1", 0, 0)
}
