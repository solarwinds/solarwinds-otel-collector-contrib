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

//go:build !integration

package cpu

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers/cpu"
	"github.com/stretchr/testify/assert"
)

func Test_Initialize_NotFailing(t *testing.T) {
	sut := NewEmitter()
	err := sut.Init()
	assert.NoError(t, err, "emitter initialization should not fail")
}

type cpuProviderMock struct {
	err error
}

func (p *cpuProviderMock) Provide() <-chan cpu.Container {
	ch := make(chan cpu.Container, 3)
	ch <- cpu.Container{
		Processors: []cpu.Processor{
			{
				Name:         "Processor 1",
				Manufacturer: "Manufacturer 1",
				Speed:        123,
				Cores:        4,
				Threads:      12,
				Model:        "78",
				Stepping:     "2",
			},
			{
				Name:         "Processor 2",
				Manufacturer: "Manufacturer 2",
				Speed:        456,
				Cores:        1,
				Threads:      2,
				Model:        "76",
				Stepping:     "1",
			},
		},
		Error: p.err,
	}
	close(ch)
	return ch
}

var _ providers.Provider[cpu.Container] = (*cpuProviderMock)(nil)

func TestEmitter_GetEmittingFunction_ReturnsEmptySliceAndErrorWhenEmitFails(t *testing.T) {
	cpuProvider := &cpuProviderMock{
		err: fmt.Errorf("provider returned error"),
	}
	sut := createCPUEmitter(cpuProvider)

	res := sut.Emit()
	assert.Equal(t, res.Error, fmt.Errorf("provider returned error"))
	assert.Equal(t, 0, res.Data.Len(), "metric slice should be empty on error")
}

func Test_GetEmittingFunction_ProvidesFunctionCapableOfMetricEmitting(t *testing.T) {
	cpuProvider := &cpuProviderMock{}
	sut := createCPUEmitter(cpuProvider)

	res := sut.Emit()
	assert.Nil(t, res.Error, fmt.Sprintf("emit function returned error: %s", res.Error))

	assert.Equal(t, res.Data.Len(), 1, "expected number of metrics is 1")
	metric := res.Data.At(0)
	assert.Equal(t, "swo.hardwareinventory.cpu", metric.Name())
	assert.Equal(t, "Current CPU clock speed", metric.Description())
	assert.Equal(t, "MHz", metric.Unit())

	dps := metric.Gauge().DataPoints()
	assert.Equal(t, 2, dps.Len(), "expected datapoint (processor) count is 2")

	attr1 := dps.At(0).Attributes()
	m1 := make(map[string]any)
	m1["processor.name"] = "Processor 1"
	m1["processor.manufacturer"] = "Manufacturer 1"
	m1["processor.cores"] = "4"
	m1["processor.threads"] = "12"
	m1["processor.model"] = "78"
	m1["processor.stepping"] = "2"
	assert.EqualValues(t, m1, attr1.AsRaw())

	attr2 := dps.At(1).Attributes()
	m2 := make(map[string]any)
	m2["processor.name"] = "Processor 2"
	m2["processor.manufacturer"] = "Manufacturer 2"
	m2["processor.cores"] = "1"
	m2["processor.threads"] = "2"
	m2["processor.model"] = "76"
	m2["processor.stepping"] = "1"
	assert.EqualValues(t, m2, attr2.AsRaw())
}
