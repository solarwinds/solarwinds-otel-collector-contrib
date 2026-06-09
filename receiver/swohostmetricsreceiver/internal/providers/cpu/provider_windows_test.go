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

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/wmi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// cpuInfoExecutorMock implements cpuInfoExecutor for tests.
type cpuInfoExecutorMock struct {
	infos []cpu.InfoStat
	err   error
}

var _ cpuInfoExecutor = (*cpuInfoExecutorMock)(nil)

func (m *cpuInfoExecutorMock) Info() ([]cpu.InfoStat, error) {
	return m.infos, m.err
}

// Test helpers — build a provider with mocked dependencies.
func newTestProvider(infos []cpu.InfoStat, infoErr error, wmiResults []interface{}, wmiErr map[interface{}]error, countsFn func(bool) (int, error)) *provider {
	return &provider{
		wmi:      wmi.CreateWmiExecutorMock(wmiResults, wmiErr),
		cpuInfo:  &cpuInfoExecutorMock{infos: infos, err: infoErr},
		countsFn: countsFn,
		logger:   zap.NewNop(),
	}
}

func fixedCounts(n int) func(bool) (int, error) {
	return func(bool) (int, error) { return n, nil }
}

// Test_Provide_HappyPath — two gopsutil sockets, one WMI descriptor (NumberOfCores=48).
// Expects two Processor entries sharing same Name/Manufacturer, Cores=48 from WMI,
// Threads derived from total-logical/socket-count, DeviceID synthesized as CPU0/CPU1.
func Test_Provide_HappyPath_TwoSocketsOneWMIDescriptor(t *testing.T) {
	infos := []cpu.InfoStat{
		{PhysicalID: "0"},
		{PhysicalID: "1"},
	}
	// 96 logical threads / 2 sockets = 48 threads per socket.
	wmiOutput := []Win32_Processor{
		{
			Name:              "Intel Xeon Gold",
			Manufacturer:      "GenuineIntel",
			CurrentClockSpeed: 2600,
			NumberOfCores:     48,
			Stepping:          "5",
			Caption:           "Intel64 Family 6 Model 85 Stepping 5",
		},
	}

	sut := newTestProvider(infos, nil, []interface{}{&wmiOutput}, nil, fixedCounts(96))
	ch := sut.Provide()
	result := <-ch
	_, open := <-ch

	assert.False(t, open, "channel must be closed")
	assert.Nil(t, result.Error)
	assert.Len(t, result.Processors, 2)

	for i, p := range result.Processors {
		assert.Equal(t, fmt.Sprintf("CPU%d", i), p.DeviceID)
		assert.Equal(t, "Intel Xeon Gold", p.Name)
		assert.Equal(t, "GenuineIntel", p.Manufacturer)
		assert.Equal(t, uint32(48), p.Cores)
		assert.Equal(t, float64(2600), p.Speed)
		assert.Equal(t, "5", p.Stepping)
		assert.Equal(t, uint32(48), p.Threads)
	}
	assert.Equal(t, "85", result.Processors[0].Model)
}

// Test_Provide_WMIZeroResults — gopsutil returns 4 sockets, WMI returns 0 rows.
// Expects 4 Processor entries with DeviceID populated but empty descriptive fields.
func Test_Provide_WMIZeroResults_GracefulDegradation(t *testing.T) {
	infos := []cpu.InfoStat{
		{PhysicalID: "0"},
		{PhysicalID: "1"},
		{PhysicalID: "2"},
		{PhysicalID: "3"},
	}
	emptyWmi := []Win32_Processor{}

	sut := newTestProvider(infos, nil, []interface{}{&emptyWmi}, nil, fixedCounts(16))
	ch := sut.Provide()
	result := <-ch
	_, open := <-ch

	assert.False(t, open, "channel must be closed")
	assert.Nil(t, result.Error)
	assert.Len(t, result.Processors, 4)

	for i, p := range result.Processors {
		assert.Equal(t, fmt.Sprintf("CPU%d", i), p.DeviceID)
		assert.Empty(t, p.Name)
		assert.Empty(t, p.Manufacturer)
		assert.Equal(t, uint32(0), p.Cores)
	}
}

// Test_Provide_GopsutilError — Info() returns error → Container{Error}.
func Test_Provide_GopsutilInfoError_PropagatesError(t *testing.T) {
	infoErr := fmt.Errorf("SMBIOS unavailable")

	sut := newTestProvider(nil, infoErr, nil, nil, fixedCounts(4))
	ch := sut.Provide()
	result := <-ch
	_, open := <-ch

	assert.False(t, open, "channel must be closed")
	assert.ErrorIs(t, result.Error, infoErr)
	assert.Nil(t, result.Processors)
}

// Test_Provide_WMIError — WMI query returns error → Container{Error}.
func Test_Provide_WMIQueryError_PropagatesError(t *testing.T) {
	infos := []cpu.InfoStat{{PhysicalID: "0"}}
	wmiErr := fmt.Errorf("WMI access denied")

	sut := newTestProvider(infos, nil, nil, map[interface{}]error{
		&[]Win32_Processor{}: wmiErr,
	}, fixedCounts(4))
	ch := sut.Provide()
	result := <-ch
	_, open := <-ch

	assert.False(t, open, "channel must be closed")
	assert.NotNil(t, result.Error)
	assert.Nil(t, result.Processors)
}

// Test_Provide_EmptyInfos — Info() returns empty slice (no error) → empty container.
func Test_Provide_EmptyGopsutilInfos_ReturnsEmptyContainer(t *testing.T) {
	sut := newTestProvider([]cpu.InfoStat{}, nil, nil, nil, fixedCounts(4))
	ch := sut.Provide()
	result := <-ch
	_, open := <-ch

	assert.False(t, open, "channel must be closed")
	assert.Nil(t, result.Error)
	assert.Empty(t, result.Processors)
}

// Test_Provide_ZeroLogicalCounts — countsFn returns 0 → Container{Error}.
func Test_Provide_ZeroLogicalCounts_ReturnsError(t *testing.T) {
	infos := []cpu.InfoStat{{PhysicalID: "0"}}
	wmiOutput := []Win32_Processor{}

	sut := newTestProvider(infos, nil, []interface{}{&wmiOutput}, nil, fixedCounts(0))
	ch := sut.Provide()
	result := <-ch
	_, open := <-ch

	assert.False(t, open, "channel must be closed")
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "non-positive")
	assert.Nil(t, result.Processors)
}
