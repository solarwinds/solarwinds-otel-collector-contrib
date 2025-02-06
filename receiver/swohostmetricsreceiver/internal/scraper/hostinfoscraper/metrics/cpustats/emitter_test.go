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

package cpustats

import (
	"errors"
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/cpustats"
	"github.com/stretchr/testify/assert"
)

func fakeCPUStatsContainer(err error) cpustats.Container {
	return cpustats.Container{
		WorkDetails: map[string][]cpustats.WorkDetail{
			cpustats.FieldTypeCPUTime: {
				{AttrName: "mode", AttrValue: cpustats.UserMode, Value: 123234},
				{AttrName: "mode", AttrValue: cpustats.SystemMode, Value: 232234},
			},
		},
		Error: err,
	}
}

type cpuStatsProviderMock struct {
	err error
}

func (p *cpuStatsProviderMock) Provide() <-chan cpustats.Container {
	ch := make(chan cpustats.Container, 3)
	ch <- fakeCPUStatsContainer(p.err)

	close(ch)
	return ch
}

func Test_Initialize_NotFailing(t *testing.T) {
	sut := createEmitter(MetricNameCPUTime, &cpuStatsProviderMock{nil})
	err := sut.Init()
	assert.NoError(t, err, "emitter initialization should not fail")
}

func TestEmitter_GetEmittingFunction_ReturnsEmptySliceAndErrorWhenEmitFails(t *testing.T) {
	sut := createEmitter(MetricNameCPUTime, &cpuStatsProviderMock{errors.New("provider returned error")})

	res := sut.Emit()
	assert.Equal(t, res.Error, fmt.Errorf("provider returned error"))
	assert.Equal(t, 0, res.Data.Len(), "metric slice should be empty on error")
}

func Test_GetEmittingFunction_ProvidesFunctionCapableOfMetricEmitting(t *testing.T) {
	sut := createEmitter(MetricNameCPUTime, &cpuStatsProviderMock{nil})

	res := sut.Emit()
	assert.Nil(t, res.Error, fmt.Sprintf("emit function returned error: %s", res.Error))

	assert.Equal(t, res.Data.Len(), 1, "expected number of metrics is 1")
	metric := res.Data.At(0)
	assert.Equal(t, MetricNameCPUTime, metric.Name())
	assert.Equal(t, description(MetricNameCPUTime), metric.Description())

	dps := metric.Gauge().DataPoints()
	assert.Equal(t, 2, dps.Len(), "expected datapoint count is 2")

	dp1 := dps.At(0)
	attr1 := dp1.Attributes()
	m1 := make(map[string]any)
	m1["mode"] = cpustats.UserMode
	assert.EqualValues(t, m1, attr1.AsRaw())
	assert.Equal(t, float64(123234), dp1.DoubleValue())

	dp2 := dps.At(1)
	attr2 := dp2.Attributes()
	m2 := make(map[string]any)
	m2["mode"] = cpustats.SystemMode
	assert.EqualValues(t, m2, attr2.AsRaw())
	assert.Equal(t, float64(232234), dp2.DoubleValue())
}
