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

package cpu

import (
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	cpuProvider "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/cpu"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
)

const (
	Name        = "swo.hardwareinventory.cpu"
	Description = "Current CPU clock speed"
	Unit        = "MHz"
)

type emitter struct {
	cpuProvider providers.Provider[cpuProvider.Container]
}

var _ metric.Emitter = (*emitter)(nil)

func NewEmitter(logger *zap.Logger) metric.Emitter {
	return createCPUEmitter(
		cpuProvider.CreateProvider(logger),
	)
}

func createCPUEmitter(
	cpuProvider providers.Provider[cpuProvider.Container],
) metric.Emitter {
	return &emitter{
		cpuProvider: cpuProvider,
	}
}

// Emit implements metric.Emitter.
func (e *emitter) Emit() *metric.Result {
	data := <-e.cpuProvider.Provide()
	if data.Error != nil {
		return &metric.Result{Data: pmetric.NewMetricSlice(), Error: data.Error}
	}

	ms, err := e.constructMetricSlice(data.Processors)
	return &metric.Result{Data: ms, Error: err}
}

// Init implements metric.Emitter.
func (e *emitter) Init() error {
	return nil
}

// Name implements metric.Emitter.
func (*emitter) Name() string {
	return Name
}

func (e *emitter) constructMetricSlice(processors []cpuProvider.Processor) (pmetric.MetricSlice, error) {
	ms := pmetric.NewMetricSlice()
	ms.EnsureCapacity(1)

	m := ms.AppendEmpty()
	m.SetName(Name)
	m.SetDescription(Description)
	m.SetUnit(Unit)

	s := m.SetEmptyGauge()
	s.DataPoints().EnsureCapacity(len(processors))

	for _, processor := range processors {
		dp := s.DataPoints().AppendEmpty()

		now := time.Now()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(now))

		attrs := generateAttributes(processor)
		err := dp.Attributes().FromRaw(convertToMapAny(attrs))
		if err != nil {
			return pmetric.NewMetricSlice(),
				fmt.Errorf(
					"storing attributes %v into datapoint failed: %w",
					attrs,
					err,
				)
		}
		dp.SetIntValue(int64(processor.Speed))
	}

	return ms, nil
}

const (
	name         = `processor.name`
	manufacturer = `processor.manufacturer`
	model        = `processor.model`
	stepping     = `processor.stepping`
	cores        = `processor.cores`
	threads      = `processor.threads`
)

func generateAttributes(processor cpuProvider.Processor) shared.Attributes {
	m := make(shared.Attributes, 7)
	if processor.Name != "" {
		m[name] = processor.Name
	}
	if processor.Manufacturer != "" {
		m[manufacturer] = processor.Manufacturer
	}
	if processor.Model != "" {
		m[model] = processor.Model
	}
	if processor.Stepping != "" {
		m[stepping] = processor.Stepping
	}
	if processor.Cores != 0 {
		m[cores] = strconv.FormatUint(uint64(processor.Cores), 10)
	}
	if processor.Threads != 0 {
		m[threads] = strconv.FormatUint(uint64(processor.Threads), 10)
	}
	return m
}

func convertToMapAny(attrs shared.Attributes) map[string]any {
	otelAttrs := make(map[string]any, len(attrs))
	for k, v := range attrs {
		otelAttrs[k] = v
	}
	return otelAttrs
}
