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

package installedsoftware

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers/installedsoftware"
	testinghelper "github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/assetscraper/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FunctionalTest_Emit(t *testing.T) {
	t.Skip("this test must be run manually only")

	emitter := NewEmitter()
	err := emitter.Init()
	if err != nil {
		assert.Fail(t, "emitter initialization failed")
	}

	r := emitter.Emit()
	fmt.Println(r.Data)
	fmt.Println(r.Error)
}

const (
	expectedNumberOfMetrics    int    = 1
	expectedMetricName         string = "swo.asset.installedsoftware"
	expectedMetricDescription  string = "carries attributes describing installed software"
	expectedMetricUnit         string = ""
	expectedNumberOfDatapoints int    = 1
	expectedDatapointValue     int64  = 0
	expectedNumberOfAttributes int    = 4
	expectedName               string = "swi swo"
	expectedPublisher          string = "swi"
	expectedVersion            string = "0.0.1"
	expectedDate               string = "2023-11-11"
	attrName                   string = "installedsoftware.name"
	attrPublisher              string = "installedsoftware.publisher"
	attrVersion                string = "installedsoftware.version"
	attrDate                   string = "installedsoftware.date"
)

func Test_MetricsAreEmitted(t *testing.T) {
	sut := createInstalledSoftwareEmitter(&InstalledSoftwareProviderStub{})
	err := sut.Init()
	require.NoErrorf(t, err, "Initialize should not return an error")

	r := sut.Emit()
	require.NoErrorf(t, err, "Emit should not return an error")
	require.Equalf(t, expectedNumberOfMetrics, r.Data.Len(), "Number of metrics in the slice should be %d", expectedNumberOfMetrics)

	metric := r.Data.At(0)
	require.Equalf(t, expectedMetricName, metric.Name(), "The metric name should be %s", expectedMetricName)
	require.Equalf(t, expectedMetricDescription, metric.Description(), "The metric description should be %s", expectedMetricDescription)
	require.Equalf(t, expectedMetricUnit, metric.Unit(), "The metric unit should be %s", expectedMetricUnit)

	datapoints := metric.Sum().DataPoints()
	require.Equalf(t, expectedNumberOfDatapoints, datapoints.Len(), "The number of datapoints in the metric should be %d", expectedNumberOfDatapoints)

	datapoint := datapoints.At(0)
	require.Equalf(t, expectedDatapointValue, datapoint.IntValue(), "The datapoint value should be %d", expectedDatapointValue)

	attributes := datapoint.Attributes()
	require.Equalf(t, expectedNumberOfAttributes, attributes.Len(), "The number of attributes in the datapoint should be %d", expectedNumberOfAttributes)

	testinghelper.RequireAttribute(t, attributes, attrName, expectedName)
	testinghelper.RequireAttribute(t, attributes, attrPublisher, expectedPublisher)
	testinghelper.RequireAttribute(t, attributes, attrVersion, expectedVersion)
	testinghelper.RequireAttribute(t, attributes, attrDate, expectedDate)
}

// Stubs.
type InstalledSoftwareProviderStub struct{}

var _ (installedsoftware.Provider) = (*InstalledSoftwareProviderStub)(nil)

func (*InstalledSoftwareProviderStub) GetSoftware() ([]installedsoftware.InstalledSoftware, error) {
	return []installedsoftware.InstalledSoftware{{
		Name:        expectedName,
		Publisher:   expectedPublisher,
		Version:     expectedVersion,
		InstallDate: expectedDate,
	}}, nil
}
