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

package firewall

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/firewall"
	"github.com/stretchr/testify/assert"
)

func Test_Emit_ProduceMetricOnProvidedProfiles(t *testing.T) {
	expectedFp := []firewall.Profile{
		{Name: "kokoha", Enabled: 1},
		{Name: "picin", Enabled: 0},
	}

	fc := firewall.Container{
		FirewallProfiles: expectedFp,
		Error:            nil,
	}

	sut := emitter{
		FirewallProvider: CreateFirewallProviderMock(fc),
	}
	err := sut.Init()
	assert.Nil(t, err, "init must not fail")
	er := sut.Emit()
	assert.Nil(t, err, "emit must not fail")

	expectedMetricCount := 1
	actualMetricCount := er.Data.Len()
	assert.Equal(t, expectedMetricCount, actualMetricCount)

	m := er.Data.At(0)
	expectedDpCount := 2
	actualDpCount := m.Gauge().DataPoints().Len()
	assert.Equal(t, expectedDpCount, actualDpCount)

	fp1 := produceFirewallProfileFromDataPoint(t, m.Gauge().DataPoints().At(0))
	fp2 := produceFirewallProfileFromDataPoint(t, m.Gauge().DataPoints().At(1))
	actualFp := []firewall.Profile{fp1, fp2}
	assert.ElementsMatch(t, expectedFp, actualFp, "data points must match to provided firewall profiles")
}

func Test_Emit_ProduceEmptyMetricOnProvidedError(t *testing.T) {
	fc := firewall.Container{
		FirewallProfiles: nil,
		Error:            fmt.Errorf("error in provider happened"),
	}

	sut := emitter{
		FirewallProvider: CreateFirewallProviderMock(fc),
	}
	err := sut.Init()
	assert.Nil(t, err, "init must not fail")
	er := sut.Emit()
	assert.NotNil(t, er.Error, "emit must fail")

	expectedMetricCount := 0
	actualMetricCount := er.Data.Len()
	assert.Equal(t, expectedMetricCount, actualMetricCount)
}

func Test_Emit_ProduceLimitedMetricOnCorruptedFirewallProfile(t *testing.T) {
	providedFp := []firewall.Profile{
		{Name: "kokoha", Enabled: 1},
		{Name: "", Enabled: 0}, // this profile will be erased from metric slice
	}

	fc := firewall.Container{
		FirewallProfiles: providedFp,
		Error:            nil,
	}

	sut := emitter{
		FirewallProvider: CreateFirewallProviderMock(fc),
	}
	err := sut.Init()
	assert.Nil(t, err, "init must not fail")
	er := sut.Emit()
	assert.Nil(t, err, "emit must not fail")

	expectedMetricCount := 1
	actualMetricCount := er.Data.Len()
	assert.Equal(t, expectedMetricCount, actualMetricCount)

	m := er.Data.At(0)
	expectedDpCount := 1
	actualDpCount := m.Gauge().DataPoints().Len()
	assert.Equal(t, expectedDpCount, actualDpCount)

	fp1 := produceFirewallProfileFromDataPoint(t, m.Gauge().DataPoints().At(0))
	actualFp := []firewall.Profile{fp1}
	expectedFp := []firewall.Profile{providedFp[0]}
	assert.ElementsMatch(t, expectedFp, actualFp, "data points must match to provided firewall profiles")
}
