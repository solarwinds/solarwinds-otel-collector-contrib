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

package cpu

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/wmi"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_Provide_ProvidesCompleteDataAndChannelIsClosedAfterDelivery(t *testing.T) {
	wmiOutput := []Win32_Processor{
		{
			Name:                      "Proc 1",
			Manufacturer:              "Manufacturer 1",
			CurrentClockSpeed:         10,
			NumberOfCores:             4,
			NumberOfLogicalProcessors: 5,
			Stepping:                  "1",
			Caption:                   "Some Caption With No Extra Data",
		},
		{
			Name:                      "Proc 2",
			Manufacturer:              "Manufacturer 2",
			CurrentClockSpeed:         3,
			NumberOfCores:             2,
			NumberOfLogicalProcessors: 6,
			Stepping:                  "",
			Caption:                   "Some Caption With Model X and Stepping 50",
		},
	}
	expectedProcessors := []Processor{
		{
			Name:         "Proc 1",
			Manufacturer: "Manufacturer 1",
			Speed:        10,
			Cores:        4,
			Threads:      5,
			Model:        "",
			Stepping:     "1",
		},
		{
			Name:         "Proc 2",
			Manufacturer: "Manufacturer 2",
			Speed:        3,
			Cores:        2,
			Threads:      6,
			Model:        "X",
			Stepping:     "50",
		},
	}
	expectedModel := Container{
		Processors: expectedProcessors,
		Error:      nil,
	}

	sut := provider{
		wmi:    wmi.CreateWmiExecutorMock([]interface{}{&wmiOutput}, nil),
		logger: zap.NewNop(),
	}

	ch := sut.Provide()
	actualModel := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedModel, actualModel)
	assert.False(t, open, "channel must be closed")
}

func Test_Provide_FailsAndProvidesEmptyObjectAndChannelIsClosedAfterDelivery(t *testing.T) {
	err := fmt.Errorf("processor error")
	expectedModel := Container{
		Processors: nil,
		Error:      fmt.Errorf("wmi query failed: %w", err),
	}

	sut := provider{
		wmi: wmi.CreateWmiExecutorMock(nil, map[interface{}]error{
			&[]Win32_Processor{}: err,
		}),
		logger: zap.NewNop(),
	}

	ch := sut.Provide()
	actualModel := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedModel, actualModel)
	assert.False(t, open, "channel must be closed")
}
