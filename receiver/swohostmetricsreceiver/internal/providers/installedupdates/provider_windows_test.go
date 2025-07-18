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
	"fmt"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/tools/wmi"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ProvidesInstalledUpdates(t *testing.T) {
	expectedUpdates := []InstalledUpdate{
		{
			Caption:     "https://support.microsoft.com/help/111111",
			Description: "Security Update",
			HotFixID:    "KB111111",
			InstalledBy: "NT AUTHORITY\\SYSTEM",
			InstalledOn: "2021-3-21",
		},
		{
			Caption:     "",
			Description: "Update",
			HotFixID:    "KB222222",
			InstalledBy: "NT AUTHORITY\\SYSTEM",
			InstalledOn: "2022-3-21",
		},
		{
			Caption:     "http://support.microsoft.com/?kbid=333333",
			Description: "Update",
			HotFixID:    "KB333333",
			InstalledBy: "",
			InstalledOn: "2023-3-21",
		},
	}
	updatesMock := []Win32_QuickFixEngineering{
		{
			Caption:     "https://support.microsoft.com/help/111111",
			Description: "Security Update",
			HotFixID:    "KB111111",
			InstalledBy: "NT AUTHORITY\\SYSTEM",
			InstalledOn: "3/21/2021",
		},
		{
			Caption:     "",
			Description: "Update",
			HotFixID:    "KB222222",
			InstalledBy: "NT AUTHORITY\\SYSTEM",
			InstalledOn: "3/21/2022",
		},
		{
			Caption:     "http://support.microsoft.com/?kbid=333333",
			Description: "Update",
			HotFixID:    "KB333333",
			InstalledBy: "",
			InstalledOn: "3/21/2023",
		},
	}

	sut := &windowsProvider{
		wmi: wmi.CreateWmiExecutorMock([]interface{}{&updatesMock}, nil),
	}

	installedUpdates, err := sut.GetUpdates()
	require.NoError(t, err)
	require.Equal(t, expectedUpdates, installedUpdates)
}

func Test_ProvidesEmptyUpdatesOnFailure(t *testing.T) {
	expectedUpdates := []InstalledUpdate{}

	sut := &windowsProvider{
		wmi: wmi.CreateWmiExecutorMock(nil, map[interface{}]error{
			&[]Win32_QuickFixEngineering{}: fmt.Errorf("no updates for you"),
		}),
	}

	installedUpdates, err := sut.GetUpdates()
	require.Error(t, err)
	require.Equal(t, expectedUpdates, installedUpdates)
}
