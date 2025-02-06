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
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/installedupdates"
	testinghelper "github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/scraper/assetscraper/metrics"
	"github.com/stretchr/testify/require"
)

const (
	expectedNumberOfMetrics    int    = 1
	expectedMetricName         string = "swo.asset.installedupdates"
	expectedMetricDescription  string = "carries attributes describing installed OS updates"
	expectedMetricUnit         string = ""
	expectedNumberOfDatapoints int    = 1
	expectedDatapointValue     int64  = 0
	expectedNumberOfAttributes int    = 5
	expectedCaption            string = "some hotfix link"
	expectedDescription        string = "some hotfix"
	expectedHotFixID           string = "some hotfix id"
	expectedInstalledOn        string = "3/21/2023"
	expectedInstalledBy        string = "someone"
	attrCaption                string = "installedupdate.caption"
	attrDescription            string = "installedupdate.description"
	attrHotFixID               string = "installedupdate.hotfixid"
	attrInstalledOn            string = "installedupdate.installedon"
	attrInstalledBy            string = "installedupdate.installedby"
)

func Test_Functional(t *testing.T) {
	t.Skip("this test should be run manually")

	sut := NewEmitter()
	err := sut.Init()
	require.NoErrorf(t, err, "Initialize should not return an error")

	r := sut.Emit()
	require.NoErrorf(t, r.Error, "Emit should not return an error")
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

	testinghelper.RequireAttribute(t, attributes, attrCaption, expectedCaption)
	testinghelper.RequireAttribute(t, attributes, attrDescription, expectedDescription)
	testinghelper.RequireAttribute(t, attributes, attrHotFixID, expectedHotFixID)
	testinghelper.RequireAttribute(t, attributes, attrInstalledOn, expectedInstalledOn)
	testinghelper.RequireAttribute(t, attributes, attrInstalledBy, expectedInstalledBy)
}

// stubs.
type InstalledUpdatesProviderStub struct{}

var _ (installedupdates.Provider) = (*InstalledUpdatesProviderStub)(nil)

// GetUpdates implements basictypes.InstalledUpdatesProvider.
func (*InstalledUpdatesProviderStub) GetUpdates() ([]installedupdates.InstalledUpdate, error) {
	return []installedupdates.InstalledUpdate{{
		Caption:     expectedCaption,
		Description: expectedDescription,
		HotFixID:    expectedHotFixID,
		InstalledOn: expectedInstalledOn,
		InstalledBy: expectedInstalledBy,
	}}, nil
}
