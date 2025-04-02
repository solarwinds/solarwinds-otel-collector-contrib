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

package hostdetails

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers/timezone"
	"github.com/stretchr/testify/assert"
)

func Test_TimezoneAttributesGenerator_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := CreateTimeZoneAttributesGenerator(timezone.CreateTimeZoneProvider())
	result := <-sut.Generate()

	fmt.Printf("Result %+v\n", result)
}

func Test_Generate_TimeZoneIsProvided_AttributesAreGenerated(t *testing.T) {
	timeZoneMock := timezone.TimeZone{
		Bias:         120,
		Caption:      "(UTC+02:00) City",
		StandardName: "Not Europe Standard Time",
	}
	expectedAttributes := shared.Attributes{
		"hostdetails.timezone.bias":         "120",
		"hostdetails.timezone.caption":      "(UTC+02:00) City",
		"hostdetails.timezone.standardname": "Not Europe Standard Time",
	}
	sut := CreateTimeZoneAttributesGenerator(
		shared.CreateProviderMock(timeZoneMock), // send valid data
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, expectedAttributes, actualAttributes, "expected attributes are not the same as actual")
}

func Test_Generate_TimeZoneDataUnavailable_AttributesAreNotGenerated(t *testing.T) {
	sut := CreateTimeZoneAttributesGenerator(shared.CreateEmptyProviderMock[timezone.TimeZone]()) // no data, closing the channel only

	actualAttributes := <-sut.Generate()

	assert.Equal(t, shared.Attributes(nil), actualAttributes, "expected atrributes are not nil")
}
