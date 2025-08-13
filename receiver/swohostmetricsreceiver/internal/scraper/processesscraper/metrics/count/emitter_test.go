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

package count

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/processesscraper/internal/metadata"
	"go.opentelemetry.io/collector/scraper"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/processescount"
)

const (
	count int64 = 1701
)

type ProcessesCountProviderMock struct{}

var _ providers.Provider[processescount.ProcessesCount] = (*ProcessesCountProviderMock)(nil)

// GetCount implements count.Provider.
func (*ProcessesCountProviderMock) Provide() <-chan processescount.ProcessesCount {
	ch := make(chan processescount.ProcessesCount, 1)
	ch <- processescount.ProcessesCount{
		Count: count,
		Error: nil,
	}
	close(ch)
	return ch
}

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

	sut := NewEmitter(
		metadata.NewMetricsBuilder(metadata.DefaultMetricsBuilderConfig(), scraper.Settings{}),
		zap.NewNop(),
	)

	err := sut.Init()
	assert.Nil(t, err)

	er := sut.Emit()
	assert.Nil(t, er.Error)

	fmt.Printf("Result: %+v\n", er.Data)
}

func Test_Initialize_NotFailing(t *testing.T) {
	sut := NewEmitter(
		metadata.NewMetricsBuilder(metadata.DefaultMetricsBuilderConfig(), scraper.Settings{}),
		zap.NewNop(),
	)
	err := sut.Init()
	require.NoError(t, err)
}

func Test_GetEmittingFunction_ProvidesFunctionCapableOfMetricsEmitting(t *testing.T) {
	provider := &ProcessesCountProviderMock{}

	sut := createEmitter(provider, metadata.NewMetricsBuilder(metadata.DefaultMetricsBuilderConfig(), scraper.Settings{}), zap.NewNop())

	er := sut.Emit()
	if er.Error != nil {
		t.Fatalf("Emitter must not fail. Error:[%s]", er.Error.Error())
	}

	metricCount := er.Data.Len()
	assert.Equal(t, metricCount, 1, "Expected number of metrics is 1")

	metric := er.Data.At(0)
	assert.Equal(t, metric.Name(), metadata.MetricsInfo.SwoSystemProcessesCount.Name)

	datapointsCount := metric.Sum().DataPoints().Len()
	assert.Equal(t, datapointsCount, 1, "Metric count is differne than expected")

	datapoint := metric.Sum().DataPoints().At(0)
	assert.Equal(t, datapoint.IntValue(), count, "Count value is different than expected")
}
