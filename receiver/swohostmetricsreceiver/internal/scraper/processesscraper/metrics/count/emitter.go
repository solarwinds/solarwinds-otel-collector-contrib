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

package count

import (
	"fmt"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/processescount"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/processesscraper/internal/metadata"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/scraper"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
)

type emitter struct {
	processesCountProvider providers.Provider[processescount.ProcessesCount]
	metricsBuilder         *metadata.MetricsBuilder
	logger                 *zap.Logger
	metric.InitFunc
}

var _ metric.Emitter = (*emitter)(nil)

func NewEmitter(logger *zap.Logger) metric.Emitter {
	return createEmitter(
		processescount.Create(processescount.CreateWrapper()),
		metadata.NewMetricsBuilder(metadata.DefaultMetricsBuilderConfig(), scraper.Settings{}),
		logger,
	)
}

func createEmitter(
	processesCountProvider providers.Provider[processescount.ProcessesCount],
	metricsBuilder *metadata.MetricsBuilder,
	logger *zap.Logger,
) metric.Emitter {
	return &emitter{
		processesCountProvider: processesCountProvider,
		metricsBuilder:         metricsBuilder,
		logger:                 logger,
	}
}

func (e *emitter) Name() string {
	return metadata.MetricsInfo.SwoSystemProcessesCount.Name
}

func (e *emitter) Emit() *metric.Result {
	value, err := e.getMetricData()
	if err != nil {
		return &metric.Result{
			Data:  pmetric.NewMetricSlice(),
			Error: fmt.Errorf("count calculation failed: %w", err),
		}
	}

	e.metricsBuilder.RecordSwoSystemProcessesCountDataPoint(pcommon.NewTimestampFromTime(time.Now()), value)
	return &metric.Result{
		Data:  e.metricsBuilder.Emit().ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics(),
		Error: err,
	}
}

func (e *emitter) getMetricData() (int64, error) {
	pc := <-e.processesCountProvider.Provide()
	return pc.Count, pc.Error
}
