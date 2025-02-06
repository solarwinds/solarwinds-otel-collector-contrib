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

package lastloggeduser

import (
	"fmt"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/loggedusers"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers"
)

const (
	MetricName        = "swo.hostinfo.user.lastLogged"
	MetricDescription = "Last logged-in users on given host"
	MetricUnit        = "User"
)

type emitter struct {
	startTime     int64
	UsersProvider providers.Provider[loggedusers.Data]
}

var _ metric.Emitter = (*emitter)(nil)

func NewEmitter() metric.Emitter {
	return createMetricEmitter(
		loggedusers.CreateProvider(),
	)
}

func createMetricEmitter(
	usersProvider providers.Provider[loggedusers.Data],
) metric.Emitter {
	return &emitter{
		UsersProvider: usersProvider,
	}
}

// Emit implements metric.Emitter.
func (e *emitter) Emit() *metric.Result {
	data := <-e.UsersProvider.Provide()
	if data.Error != nil {
		return &metric.Result{
			Data:  pmetric.NewMetricSlice(),
			Error: data.Error,
		}
	}

	ms, err := e.constructMetricSlice(data.Users)
	return &metric.Result{
		Data:  ms,
		Error: err,
	}
}

// Init implements metric.Emitter.
func (e *emitter) Init() error {
	e.startTime = time.Now().UnixNano()
	return nil
}

// Name implements metric.Emitter.
func (*emitter) Name() string {
	return MetricName
}

func (e *emitter) constructMetricSlice(users []loggedusers.User) (pmetric.MetricSlice, error) {
	ms := pmetric.NewMetricSlice()
	ms.EnsureCapacity(1)

	m := ms.AppendEmpty()
	m.SetName(MetricName)
	m.SetDescription(MetricDescription)
	m.SetUnit(MetricUnit)

	s := m.SetEmptyGauge()
	s.DataPoints().EnsureCapacity(len(users))

	for _, user := range users {
		dp := s.DataPoints().AppendEmpty()

		now := time.Now()
		dp.SetTimestamp(pcommon.NewTimestampFromTime(now))

		attrs := generateAttributes(user)
		otelAtts := make(map[string]any, len(attrs))
		for k, v := range attrs {
			otelAtts[k] = v
		}
		err := dp.Attributes().FromRaw(otelAtts)
		if err != nil {
			return pmetric.NewMetricSlice(),
				fmt.Errorf(
					"storing attributes into %v dapapoint failed: %w",
					attrs,
					err,
				)
		}
		dp.SetIntValue(1)
	}

	return ms, nil
}

const (
	name        = `user.name`
	displayName = `user.displayname`
	tty         = "user.tty"
)

func generateAttributes(user loggedusers.User) shared.Attributes {
	m := make(shared.Attributes, 2)
	if len(user.Name) > 0 {
		m[name] = user.Name
	}
	if len(user.DisplayName) > 0 {
		m[displayName] = user.DisplayName
	}
	if len(user.TTY) > 0 {
		m[tty] = user.TTY
	}
	return m
}
