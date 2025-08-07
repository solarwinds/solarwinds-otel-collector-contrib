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

package timezone

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/wmi"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_Provider_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := CreateTimeZoneProvider(zap.NewNop())
	result := <-sut.Provide()
	fmt.Printf("Result: %+v\n", result)
}

func Test_Provide_ProvidesCompleteDataAndChannelIsClosedAfterDelivery(t *testing.T) {
	expectedTimeZone := TimeZone{
		Bias:         30,
		Caption:      "(UTC+11:00) Some Time Zone",
		StandardName: "Central Antarctic Standard Time",
	}

	timeZoneMock := []Win32_TimeZone{
		{
			Bias:         30,
			Caption:      "(UTC+11:00) Some Time Zone",
			StandardName: "Central Antarctic Standard Time",
		},
	}

	sut := provider{
		wmi:    wmi.CreateWmiExecutorMock([]interface{}{&timeZoneMock}, nil),
		logger: zap.NewNop(),
	}

	ch := sut.Provide()
	actualTimeZone := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedTimeZone, actualTimeZone)
	assert.False(t, open, "channel must be closed")
}

func Test_Provide_FailsAndProvidesEmptyObjectAndChannelIsClosedAfterDelivery(t *testing.T) {
	expectedTimeZone := TimeZone{
		Bias:         0,
		Caption:      "",
		StandardName: "",
	}

	errors := map[interface{}]error{
		&[]Win32_TimeZone{}: fmt.Errorf("some timezone error"),
	}
	sut := provider{
		wmi:    wmi.CreateWmiExecutorMock(nil, errors),
		logger: zap.NewNop(),
	}

	ch := sut.Provide()
	actualDomain := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedTimeZone, actualDomain)
	assert.False(t, open, "channel must be closed")
}
