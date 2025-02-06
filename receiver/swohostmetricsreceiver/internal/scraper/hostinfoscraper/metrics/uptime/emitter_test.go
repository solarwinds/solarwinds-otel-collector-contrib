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

package uptime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/uptime"
)

const (
	presetUptime uint64 = 1701
)

type UptimeProviderMock struct{}

// GetUptime implements uptime.Provider.
func (*UptimeProviderMock) GetUptime() <-chan uptime.Uptime {
	ch := make(chan uptime.Uptime, 1)
	ch <- uptime.Uptime{
		Uptime: presetUptime,
		Error:  nil,
	}
	close(ch)
	return ch
}

var _ uptime.Provider = (*UptimeProviderMock)(nil)

type mock struct {
	Atts shared.Attributes
}

// Generate implements shared.AttributesGenerator.
func (m *mock) Generate() shared.AttributesChannel {
	ch := make(shared.AttributesChannel)
	go func() {
		ch <- m.Atts
		close(ch)
	}()

	return ch
}

var _ shared.AttributesGenerator = (*mock)(nil)

func CreateAttributesMock(
	atts shared.Attributes,
) shared.AttributesGenerator {
	return &mock{
		Atts: atts,
	}
}

func Test_Functional(t *testing.T) {
	t.Skip("This test should be run manually only")

	sut := NewEmitter()

	err := sut.Init()
	assert.Nil(t, err)

	er := sut.Emit()
	assert.Nil(t, er.Error)

	fmt.Printf("Result: %+v\n", er.Data)
}

func Test_Initialize_NotFailing(t *testing.T) {
	sut := NewEmitter()
	err := sut.Init()
	require.NoError(t, err)
}

func Test_GetEmittingFunction_ProvidesFunctionCapableOfMetricsEmitting(t *testing.T) {
	uptimeProvider := &UptimeProviderMock{}
	osDetailsAttributes := CreateAttributesMock(map[string]string{
		"osdetails.kokoha": "666",
	})
	hostDetailsAttributes := CreateAttributesMock(map[string]string{
		"hostdetails.kokoha": "777",
	})

	sut := createUptimeEmitter(hostDetailsAttributes, osDetailsAttributes, uptimeProvider)

	er := sut.Emit()
	if er.Error != nil {
		t.Fatalf("Emitter must not fail. Error:[%s]", er.Error.Error())
	}

	metricCount := er.Data.Len()
	assert.Equal(t, metricCount, 1, "Expected number of metrics is 1")

	metric := er.Data.At(0)
	assert.Equal(t, metric.Name(), MetricName)
	assert.Equal(t, metric.Description(), MetricDescription)
	assert.Equal(t, metric.Unit(), MetricUnit)

	datapointsCount := metric.Sum().DataPoints().Len()
	assert.Equal(t, datapointsCount, 1, "Metric count is differne than expected")

	datapoint := metric.Sum().DataPoints().At(0)
	assert.Equal(t, uint64(datapoint.IntValue()), presetUptime, "Uptime value is different than expected") //nolint:gosec // equals.
	assert.Equal(t, datapoint.Attributes().Len(), 2, "Count of attributes is different than expected")
	m := make(map[string]any)
	m["osdetails.kokoha"] = "666"
	m["hostdetails.kokoha"] = "777"
	assert.EqualValues(t, datapoint.Attributes().AsRaw(), m)
}
