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

package installedupdates

import (
	"fmt"
	"time"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/installedupdates"
	metricshelper "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/assetscraper/metrics"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
)

const (
	Name        = "swo.asset.installedupdates"
	description = "carries attributes describing installed OS updates"
	unit        = ""
)

type Emitter struct {
	provider      installedupdates.Provider
	startTimeNano time.Time
	logger        *zap.Logger
}

var _ metric.Emitter = (*Emitter)(nil)

func NewEmitter(logger *zap.Logger) metric.Emitter {
	return createEmitter(
		installedupdates.NewProvider(logger),
		logger,
	)
}

func createEmitter(
	provider installedupdates.Provider,
	logger *zap.Logger,
) metric.Emitter {
	return &Emitter{
		provider: provider,
		logger:   logger,
	}
}

// Emit implements metric.Emitter.
func (emitter *Emitter) Emit() *metric.Result {
	ms, err := emitter.populateMetric()
	if err != nil {
		return &metric.Result{
			Data:  pmetric.NewMetricSlice(),
			Error: fmt.Errorf("failed to populate metric %s: %w", Name, err),
		}
	}

	return &metric.Result{
		Data:  ms,
		Error: nil,
	}
}

// Init implements metric.Emitter.
func (emitter *Emitter) Init() error {
	emitter.startTimeNano = time.Now()
	return nil
}

// Name implements metric.Emitter.
func (emitter *Emitter) Name() string {
	return Name
}

func (emitter *Emitter) populateMetric() (pmetric.MetricSlice, error) {
	installedUpdates, err := emitter.provider.GetUpdates()
	if err != nil {
		return pmetric.NewMetricSlice(), fmt.Errorf("failed to obtain installed updates: %w", err)
	}

	// Nothing to be sent up. No error.
	if len(installedUpdates) == 0 {
		emitter.logger.Debug("no installed update was obtained from installupdate metric emitter")
		return pmetric.NewMetricSlice(), nil
	}

	metadata := metricshelper.MetricMetadata{
		Name:        Name,
		Description: description,
		Unit:        unit,
	}
	metricSlice := metricshelper.ConstructMetricBase(metadata)
	metric := metricSlice.At(0)

	dataPoints := metricshelper.PrepareEmptySum(metric, len(installedUpdates))

	for _, update := range installedUpdates {
		dataPoint := metricshelper.AppendNumberDataPoint(dataPoints, emitter.startTimeNano)

		rawAttributes := getAttributes(update)
		_ = dataPoint.Attributes().FromRaw(rawAttributes)
	}

	return metricSlice, nil
}
