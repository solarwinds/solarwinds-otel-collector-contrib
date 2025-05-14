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

package firewall

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/firewall"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
)

const (
	MetricName        = "swo.hostinfo.firewall"
	MetricDescription = "Windows firewall profile statuses. Value 1 means profile is enabled, value 0 means profile is disabled."
	MetricUnit        = "{status}"
)

const (
	akFirewallProfile = "firewall.profile.name"
	akCount           = 1
)

type emitter struct {
	startTime        time.Time
	FirewallProvider providers.Provider[firewall.Container]
}

var _ metric.Emitter = (*emitter)(nil)

func NewEmitter() metric.Emitter {
	return createEmitter(
		firewall.CreateFirewallProvider(),
	)
}

func createEmitter(
	fp providers.Provider[firewall.Container],
) metric.Emitter {
	return &emitter{
		FirewallProvider: fp,
	}
}

// Emit implements metric.Emitter.
func (e *emitter) Emit() *metric.Result {
	// This emitter is not supported on other systems than Windows
	if runtime.GOOS != "windows" {
		return &metric.Result{
			Data:  pmetric.NewMetricSlice(),
			Error: nil,
		}
	}
	// Get data from providers.
	p, err := e.getData()
	if err != nil {
		return &metric.Result{
			Data:  pmetric.NewMetricSlice(),
			Error: err,
		}
	}
	// Construct metric slice.
	ms := e.constructMetric(p)

	return &metric.Result{
		Data:  ms,
		Error: err,
	}
}

func (e *emitter) getData() ([]firewall.Profile, error) {
	ch := e.FirewallProvider.Provide()
	fc := <-ch // get data from channel and continue
	if fc.Error != nil {
		message := "getting data from firewall provider failed"
		err := fmt.Errorf("%s %w", message, fc.Error)
		zap.L().Error(message, zap.Error(fc.Error))
		return nil, err
	}
	return fc.FirewallProfiles, nil
}

func (e *emitter) constructMetric(
	fps []firewall.Profile,
) pmetric.MetricSlice {
	if len(fps) == 0 {
		zap.L().Warn("no firewall profiles provided")
		return pmetric.NewMetricSlice()
	}
	// Metric slice init.
	ms := pmetric.NewMetricSlice()
	ms.EnsureCapacity(1)
	// Metric init.
	m := ms.AppendEmpty()
	m.SetName(MetricName)
	m.SetDescription(MetricDescription)
	m.SetUnit(MetricUnit)
	// Populate metric data points with provided data
	e.populateMetric(fps, m)
	return ms
}

func (e *emitter) populateMetric(
	fps []firewall.Profile,
	m pmetric.Metric,
) {
	// Sum and data points
	g := m.SetEmptyGauge()
	dps := g.DataPoints()
	dps.EnsureCapacity(len(fps))
	// process datapoints
	var wg sync.WaitGroup
	wg.Add(len(fps))
	for _, fp := range fps {
		if !checkFirewallProfileValidity(fp) {
			wg.Done() // this profile will not be processed
			continue
		}
		dp := dps.AppendEmpty()
		go func(fp firewall.Profile, dp pmetric.NumberDataPoint) {
			defer wg.Done()
			e.populateDatapoint(fp, dp)
		}(fp, dp)
	}
	wg.Wait()
}

func checkFirewallProfileValidity(
	fp firewall.Profile,
) bool {
	return fp.Name != ""
}

func (e *emitter) populateDatapoint(
	fp firewall.Profile,
	dp pmetric.NumberDataPoint,
) {
	dp.SetIntValue(int64(fp.Enabled))
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(e.startTime))
	now := time.Now()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(now))
	atts := dp.Attributes()
	atts.EnsureCapacity(akCount)
	atts.PutStr(akFirewallProfile, fp.Name)
}

// Init implements metric.Emitter.
func (e *emitter) Init() error {
	e.startTime = time.Now()

	return nil
}

// Name implements metric.Emitter.
func (*emitter) Name() string {
	return MetricName
}
