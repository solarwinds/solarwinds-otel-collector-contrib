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

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func Test_Example_HowToFillScopeMetrics(t *testing.T) {
	// product of metric emitter
	ms := pmetric.NewMetricSlice()
	m := ms.AppendEmpty()
	m.SetName("kokoha.metric")
	m.SetDescription("This is mighty kokoha metric")
	s := m.SetEmptySum()
	s.DataPoints().EnsureCapacity(2)
	s.DataPoints().AppendEmpty().SetIntValue(1701)
	s.DataPoints().AppendEmpty().SetIntValue(1702)

	// scope metric emitter product
	rm := pmetric.NewResourceMetrics()
	sm := rm.ScopeMetrics().AppendEmpty()
	ms.MoveAndAppendTo(sm.Metrics())

	assert.Equal(t, 1, rm.ScopeMetrics().At(0).Metrics().Len(), "There must be exactly one metric")
	assert.Equal(t, "kokoha.metric", rm.ScopeMetrics().At(0).Metrics().At(0).Name(), "Metric name must be the same")
	assert.Equal(t, 2, rm.ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().Len(), "Number of data points must fit")
	assert.Equal(t, int64(1701), rm.ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).IntValue(), "Value must be the same")
}

func Test_ScopeEmitter_EmptyMetrics_ReturnsEmptyScope(t *testing.T) {
	// Create an empty scope metrics slice
	emptyScopeData := pmetric.NewScopeMetricsSlice()

	// Create scope emitter mock that returns empty data
	scopeEmitter := CreateEmitterMock(
		&Result{
			Data:  emptyScopeData,
			Error: nil,
		},
		nil,
		"test.scope",
	)

	// Initialize and emit
	err := scopeEmitter.Init()
	assert.NoError(t, err, "scope emitter initialization should succeed")

	result := scopeEmitter.Emit()
	assert.NoError(t, result.Error, "scope emitter should not return error for empty metrics")
	assert.Equal(t, 0, result.Data.Len(), "scope metrics slice should be empty")
}

func Test_ScopeEmitter_AllEmptyMetrics_ReturnsEmptyScope(t *testing.T) {
	// Create an empty scope metrics slice
	emptyScopeData := pmetric.NewScopeMetricsSlice()

	// Create scope emitter mock that returns empty data
	scopeEmitter := CreateEmitterMock(
		&Result{
			Data:  emptyScopeData,
			Error: nil,
		},
		nil,
		"test.scope.all.empty",
	)

	// Initialize and emit
	err := scopeEmitter.Init()
	assert.NoError(t, err, "scope emitter initialization should succeed")

	result := scopeEmitter.Emit()
	assert.NoError(t, result.Error, "scope emitter should not return error when all metrics are empty")
	assert.Equal(t, 0, result.Data.Len(), "scope metrics slice should be empty when all metrics are empty")
}
