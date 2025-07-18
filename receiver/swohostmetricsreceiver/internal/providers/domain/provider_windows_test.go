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

package domain

import (
	"fmt"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/tools/wmi"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Provider_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := CreateDomainProvider()
	result := <-sut.Provide()
	fmt.Printf("Result: %+v\n", result)
}

func Test_Provide_ProvidesCompleteDataAndChannelIsClosedAfterDelivery(t *testing.T) {
	wmiOutput := []Win32_ComputerSystem{
		{
			Name:        "SWI-D5ZRKQ2",
			Domain:      "swdev.local",
			DNSHostName: "SWI-D5ZRKQ2",
			DomainRole:  1,
		},
	}

	expectedDomain := Domain{
		Domain:     "swdev.local",
		FQDN:       "SWI-D5ZRKQ2.swdev.local",
		DomainRole: 1,
		Workgroup:  "",
	}

	sut := provider{
		wmi: wmi.CreateWmiExecutorMock([]interface{}{&wmiOutput}, nil),
	}

	ch := sut.Provide()
	actualDomain := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedDomain, actualDomain)
	assert.False(t, open, "channel must be closed")
}

func Test_Provide_FailsAndProvidesEmptyObjectAndChannelIsClosedAfterDelivery(t *testing.T) {
	expectedDomain := Domain{
		Domain:     "",
		FQDN:       "",
		DomainRole: 0,
		Workgroup:  "",
	}

	sut := provider{
		wmi: wmi.CreateWmiExecutorMock(nil, map[interface{}]error{
			&[]Win32_ComputerSystem{}: fmt.Errorf("kokoha error"),
		}),
	}

	ch := sut.Provide()
	actualDomain := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedDomain, actualDomain)
	assert.False(t, open, "channel must be closed")
}
