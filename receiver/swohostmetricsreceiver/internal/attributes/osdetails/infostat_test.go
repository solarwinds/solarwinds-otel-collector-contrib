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

package osdetails

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/stretchr/testify/assert"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/infostat"
)

func Test_InfoStatAttributesGenerator_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := CreateInfoStatAttributesGenerator(infostat.CreateInfoStatProvider())
	result := <-sut.Generate()

	fmt.Printf("Result %+v\n", result)
}

func Test_Generate_InfoStatIsProvided_AttributesAreGenerated(t *testing.T) {
	providedInfoStat := infostat.InfoStat{
		Hostname:           "expected-Hostname",
		BootTime:           123,
		Os:                 "expected-Os",
		Platform:           "expected-Platform",
		PlatformFamily:     "expected-PlatformFamily",
		PlatformVersion:    "expected-PlatformVersion",
		KernelVersion:      "expected-KernelVersion",
		KernelArchitecture: "expected-KernelArchitecture",
		HostID:             "expected-HostID",
	}

	expectedAttributes := shared.Attributes{
		"osdetails.hostname":            providedInfoStat.Hostname,
		"osdetails.boottime":            strconv.FormatUint(providedInfoStat.BootTime, 10),
		"osdetails.os":                  providedInfoStat.Os,
		"osdetails.platform":            providedInfoStat.Platform,
		"osdetails.platform.family":     providedInfoStat.PlatformFamily,
		"osdetails.platform.version":    providedInfoStat.PlatformVersion,
		"osdetails.kernel.version":      providedInfoStat.KernelVersion,
		"osdetails.kernel.architecture": providedInfoStat.KernelArchitecture,
		"osdetails.host.id":             providedInfoStat.HostID,
	}

	sut := CreateInfoStatAttributesGenerator(
		shared.CreateProviderMock(providedInfoStat), // send valid data
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(
		t,
		expectedAttributes,
		actualAttributes,
		"expected attributes are not the same as actual",
	)
}

func Test_Generate_InfoStatDataUnavailable_AttributesAreNotGenerated(t *testing.T) {
	sut := CreateInfoStatAttributesGenerator(
		shared.CreateEmptyProviderMock[infostat.InfoStat](), // not sending any data just closing the channel
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(
		t,
		shared.Attributes(nil),
		actualAttributes,
		"expected attributes are not the same as actual",
	)
}
