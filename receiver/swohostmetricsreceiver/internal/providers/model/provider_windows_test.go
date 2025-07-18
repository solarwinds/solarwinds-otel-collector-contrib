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

package model

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/tools/wmi"

	"github.com/stretchr/testify/assert"
)

func Test_Provider_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := CreateModelProvider()
	result := <-sut.Provide()
	fmt.Printf("Result: %+v\n", result)
}

func Test_Provide_ProvidesCompleteData(t *testing.T) {
	expectedModel := Model{
		SerialNumber: "NCC-1701E",
		Manufacturer: "Starfleet Yards",
		Name:         "USS Enterprise",
	}

	biosMock := []Win32_BIOS{
		{
			SerialNumber: "NCC-1701E",
		},
	}
	computerSystemMock := []Win32_ComputerSystem{
		{
			Model:        "USS Enterprise",
			Manufacturer: "Starfleet Yards",
		},
	}

	processActWithEvaluation(t, []interface{}{&biosMock, &computerSystemMock}, nil, expectedModel)
}

func Test_Provide_ProvidePartialModelOnBiosCommandFailure(t *testing.T) {
	expectedModel := Model{
		SerialNumber: "",
		Manufacturer: "Starfleet Yards",
		Name:         "USS Enterprise",
	}

	computerSystemMock := []Win32_ComputerSystem{
		{
			Model:        "USS Enterprise",
			Manufacturer: "Starfleet Yards",
		},
	}
	errors := map[interface{}]error{
		&[]Win32_BIOS{}: fmt.Errorf("BIOS is not available"),
	}

	processActWithEvaluation(t, []interface{}{&computerSystemMock}, errors, expectedModel)
}

func Test_Provide_ProvidePartialModelOnComputerSystemCommandFailure(t *testing.T) {
	expectedModel := Model{
		SerialNumber: "NCC-1701E",
		Manufacturer: "",
		Name:         "",
	}

	biosMock := []Win32_BIOS{
		{
			SerialNumber: "NCC-1701E",
		},
	}
	errors := map[interface{}]error{
		&[]Win32_ComputerSystem{}: fmt.Errorf("computer system is not available"),
	}

	processActWithEvaluation(t, []interface{}{&biosMock}, errors, expectedModel)
}

func Test_Provide_ProvideEmptyObjectOnBothFailures(t *testing.T) {
	expectedModel := Model{
		SerialNumber: "",
		Manufacturer: "",
		Name:         "",
	}

	errors := map[interface{}]error{
		&[]Win32_ComputerSystem{}: fmt.Errorf("computer system is not available"),
		&[]Win32_BIOS{}:           fmt.Errorf("BIOS is not available"),
	}

	processActWithEvaluation(t, nil, errors, expectedModel)
}

func processActWithEvaluation(
	t *testing.T,
	mocks []interface{},
	errors map[interface{}]error,
	expectedModel Model,
) {
	sut := provider{
		wmi.CreateWmiExecutorMock(mocks, errors),
	}

	ch := sut.Provide()
	actualModel := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedModel, actualModel)
	assert.False(t, open, "channel must be closed")
}
