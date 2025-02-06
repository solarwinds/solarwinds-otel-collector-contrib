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

package uptime

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/hostdetails"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/osdetails"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/domain"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/infostat"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/language"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/model"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/timezone"
	uptimeprovider "github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/uptime"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/synchronization"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
)

const (
	MetricName        = "swo.hostinfo.uptime"
	MetricDescription = "Host uptime in seconds"
	MetricUnit        = "s"
)

type emitter struct {
	startTime                      time.Time
	HostdetailsAttributesGenerator shared.AttributesGenerator
	OsdetailsAttributesGenerator   shared.AttributesGenerator
	UptimeProvider                 uptimeprovider.Provider
}

var _ metric.Emitter = (*emitter)(nil)

var errOverflowInt64 = errors.New("number overflows int64")

// safeUint64ToInt64 converts uint64 to int64 with
// bounds checking.
func safeUint64ToInt64(i uint64) (int64, error) {
	if i > math.MaxInt64 {
		return 0, errOverflowInt64
	}

	return int64(i), nil
}

func NewEmitter() metric.Emitter {
	return createUptimeEmitter(
		hostdetails.CreateHostDetailsAttributesGenerator(
			hostdetails.CreateDomainAttributesGenerator(
				domain.CreateDomainProvider(),
			),
			hostdetails.CreateModelAttributesGenerator(
				model.CreateModelProvider(),
			),
			hostdetails.CreateTimeZoneAttributesGenerator(
				timezone.CreateTimeZoneProvider(),
			),
		),
		osdetails.CreateOsDetailsAttributesGenerator(
			osdetails.CreateInfoStatAttributesGenerator(
				infostat.CreateInfoStatProvider(),
			),
			osdetails.CreateLanguageAttributesGenerator(
				language.CreateLanguageProvider(),
			),
		),
		uptimeprovider.CreateUptimeProvider(
			uptimeprovider.CreateUptimeWrapper(),
		),
	)
}

func createUptimeEmitter(
	hostdetailsAttributesGenerator shared.AttributesGenerator,
	osdetailsAttributesGenerator shared.AttributesGenerator,
	uptimeProvider uptimeprovider.Provider,
) metric.Emitter {
	return &emitter{
		HostdetailsAttributesGenerator: hostdetailsAttributesGenerator,
		OsdetailsAttributesGenerator:   osdetailsAttributesGenerator,
		UptimeProvider:                 uptimeProvider,
	}
}

// Emit implements metric.Emitter.
func (e *emitter) Emit() *metric.Result {
	uptime, atts, err := e.getMetricData()
	if err != nil {
		return &metric.Result{
			Data:  pmetric.NewMetricSlice(),
			Error: fmt.Errorf("uptime calculation failed: %w", err),
		}
	}

	ms, err := e.constructMetricSlice(uptime, atts)
	return &metric.Result{
		Data:  ms,
		Error: err,
	}
}

// Init implements metric.Emitter.
func (e *emitter) Init() error {
	e.startTime = time.Now()

	return nil
}

// Name implements metric.Emitter.
func (e *emitter) Name() string {
	return MetricName
}

func (e *emitter) getMetricData() (
	uptime int64,
	atts shared.Attributes,
	err error,
) {
	const sourcesCount = 3

	var wg sync.WaitGroup
	wg.Add(sourcesCount)

	hostCh := e.HostdetailsAttributesGenerator.Generate()
	osCh := e.OsdetailsAttributesGenerator.Generate()
	uptimeCh := e.UptimeProvider.GetUptime()

	terminationCh := synchronization.ActivateSupervisingRoutine(&wg)

	atts = make(shared.Attributes)

loop:
	for {
		select {
		case hAtts, open := <-hostCh:
			if !open {
				hostCh = nil
				wg.Done()
			} else {
				mergeAttributes(atts, hAtts)
			}
		case osAtts, open := <-osCh:
			if !open {
				osCh = nil
				wg.Done()
			} else {
				mergeAttributes(atts, osAtts)
			}
		case uVal, open := <-uptimeCh:
			if !open {
				uptimeCh = nil
				wg.Done()
			} else {
				uptime, err = calculateUptime(uVal)
			}
		case <-terminationCh:
			break loop
		}
	}

	return uptime, atts, err
}

func mergeAttributes(base shared.Attributes, increment shared.Attributes) {
	for k, v := range increment {
		base[k] = v
	}
}

func calculateUptime(uptime uptimeprovider.Uptime) (int64, error) {
	if uptime.Error != nil {
		message := fmt.Sprintf("Failed to receive uptime from %s emitter", MetricName)
		zap.L().Error(message, zap.Error(uptime.Error))
		return 0, fmt.Errorf("%s: %w", message, uptime.Error)
	}

	uptimeInt64, err := safeUint64ToInt64(uptime.Uptime)
	if err != nil {
		return 0, err
	}

	if zap.L().Core().Enabled(zap.DebugLevel) {
		days := uptimeInt64 / (60 * 60 * 24)
		hours := (uptimeInt64 - (days * 60 * 60 * 24)) / (60 * 60)
		minutes := (uptimeInt64 - (days * 60 * 60 * 24) - (hours * 60 * 60)) / 60
		seconds := (uptimeInt64 - (days * 60 * 60 * 24) - (hours * 60 * 60) - (minutes * 60))
		zap.L().Debug(
			"UptimeMetricEmitter: uptime received",
			zap.Int64("uptime", uptimeInt64),
			zap.Int64("uptime_days", days),
			zap.Int64("uptime_hours", hours),
			zap.Int64("uptime_minutes", minutes),
			zap.Int64("uptime_seconds", seconds),
		)
	}

	return uptimeInt64, nil
}

func (e *emitter) constructMetricSlice(
	dpValue int64,
	atts shared.Attributes,
) (pmetric.MetricSlice, error) {
	ms := pmetric.NewMetricSlice()
	ms.EnsureCapacity(1)

	m := ms.AppendEmpty()
	m.SetName(MetricName)
	m.SetDescription(MetricDescription)
	m.SetUnit(MetricUnit)

	s := m.SetEmptySum()
	// metric is cumulative => counter
	s.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	// metric is monotonic => For cumulative monotonic sums, this means the reader SHOULD expect values
	// that are not less than the previous value.
	s.SetIsMonotonic(true)
	s.DataPoints().EnsureCapacity(1)

	dp := s.DataPoints().AppendEmpty()
	dp.SetFlags(pmetric.DefaultDataPointFlags)
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(e.startTime))
	now := time.Now()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(now))
	dp.SetIntValue(dpValue)

	otelAtts := make(map[string]any, len(atts))
	for k, v := range atts {
		otelAtts[k] = v
	}
	err := dp.Attributes().FromRaw(otelAtts)
	if err != nil {
		return pmetric.NewMetricSlice(),
			fmt.Errorf(
				"storing attributes into %v dapapoint failed: %w",
				atts,
				err,
			)
	}

	return ms, nil
}
