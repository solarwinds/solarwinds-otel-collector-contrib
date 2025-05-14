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

import "go.opentelemetry.io/collector/pdata/pmetric"

func CreateMetricEmitterMockV2(
	name string,
	dpCount int,
	dpValue int,
) Emitter {
	return &metricEmitterMock{
		name:        name,
		metricSlice: createMetricSlice(name, dpCount, dpValue),
	}
}

func createMetricSlice(
	metricName string,
	dpCount int,
	dpValue int,
) pmetric.MetricSlice {
	ms := pmetric.NewMetricSlice()
	m := ms.AppendEmpty()
	m.SetName(metricName)

	s := m.SetEmptySum()

	for i := 0; i < dpCount; i++ {
		dp := s.DataPoints().AppendEmpty()
		dp.SetIntValue(int64(dpValue))
	}

	return ms
}

type metricEmitterMock struct {
	name        string
	metricSlice pmetric.MetricSlice
}

// Emit implements MetricEmitter.
func (m *metricEmitterMock) Emit() *Result {
	return &Result{Data: m.metricSlice, Error: nil}
}

// Init implements MetricEmitter.
func (m *metricEmitterMock) Init() error {
	return nil
}

// Name implements MetricEmitter.
func (m *metricEmitterMock) Name() string {
	return m.name
}
