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
	"fmt"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/cpustats"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

const (
	MetricNameCPUTime      = "os.cpu.time_microseconds"
	MetricNameProcs        = "os.cpu.procs"
	MetricNameCurrentProcs = "os.cpu.current_procs"
	MetricNameIntr         = "os.cpu.intr"
	MetricNameCtxt         = "os.cpu.ctxt"
	MetricNameNumCores     = "os.cpu.numcores"
)

type emitter struct {
	metricName       string
	cpuStatsProvider providers.Provider[cpustats.Container]
}

var _ metric.Emitter = (*emitter)(nil)

func NewEmitter(metricName string) metric.EmitterCreateFunc {
	return func(logger *zap.Logger) metric.Emitter {
		return createEmitter(metricName, cpustats.CreateProvider(logger))
	}
}

func createEmitter(metricName string, cpuStatsProvider providers.Provider[cpustats.Container]) metric.Emitter {
	return &emitter{
		metricName:       metricName,
		cpuStatsProvider: cpuStatsProvider,
	}
}

func (e *emitter) Emit() *metric.Result {
	data := <-e.cpuStatsProvider.Provide()
	if data.Error != nil {
		return &metric.Result{Data: pmetric.NewMetricSlice(), Error: data.Error}
	}

	ms, err := e.constructMetricSlice(data)
	return &metric.Result{Data: ms, Error: err}
}

func (e *emitter) Init() error {
	return nil
}

func (e *emitter) Name() string {
	return e.metricName
}

func (e *emitter) constructMetricSlice(data cpustats.Container) (pmetric.MetricSlice, error) {
	fieldType := getMetricFieldMapping(e.metricName)
	if fieldType == "" {
		// unknown type
		return pmetric.NewMetricSlice(), nil
	}

	attrs, ok := data.WorkDetails[fieldType]
	if !ok {
		// data for metric not available
		return pmetric.NewMetricSlice(), nil
	}

	now := time.Now()

	ms := pmetric.NewMetricSlice()
	ms.EnsureCapacity(1)

	m := ms.AppendEmpty()
	m.SetName(e.metricName)
	m.SetDescription(description(e.metricName))

	g := m.SetEmptyGauge()
	g.DataPoints().EnsureCapacity(len(attrs))

	for _, attr := range attrs {
		dp := g.DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(now))

		if attr.AttrName != "" {
			sharedAttrs := shared.Attributes{
				attr.AttrName: attr.AttrValue,
			}

			err := dp.Attributes().FromRaw(convertToMapAny(sharedAttrs))
			if err != nil {
				return pmetric.NewMetricSlice(),
					fmt.Errorf(
						"storing attributes %v into datapoint failed: %w",
						sharedAttrs,
						err,
					)
			}
		}

		dp.SetDoubleValue(attr.Value)
	}

	return ms, nil
}

func convertToMapAny(attrs shared.Attributes) map[string]any {
	otelAttrs := make(map[string]any, len(attrs))
	for k, v := range attrs {
		otelAttrs[k] = v
	}
	return otelAttrs
}

func getMetricFieldMapping(metricName string) string {
	switch metricName {
	case MetricNameCPUTime:
		return cpustats.FieldTypeCPUTime
	case MetricNameProcs:
		return cpustats.FieldTypeProcesses
	case MetricNameCurrentProcs:
		return cpustats.FieldTypeCurrentProcs
	case MetricNameIntr:
		return cpustats.FieldTypeCtxt
	case MetricNameCtxt:
		return cpustats.FieldTypeCtxt
	case MetricNameNumCores:
		return cpustats.FieldTypeNumCores
	default:
		return ""
	}
}

func description(metricName string) string {
	cpuStatsDescription := map[string]string{
		MetricNameCPUTime:      "Amount of time CPU has spent performing work",
		MetricNameProcs:        "Total number of processes and threads created",
		MetricNameCurrentProcs: "Number of processes currently running on CPUs",
		MetricNameIntr:         "Count of interrupts serviced since boot time",
		MetricNameCtxt:         "Total number of context switches across all CPUs",
		MetricNameNumCores:     "Number of CPU cores",
	}
	return cpuStatsDescription[metricName]
}
