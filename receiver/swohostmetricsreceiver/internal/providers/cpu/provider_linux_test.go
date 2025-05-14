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

package cpu

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/stretchr/testify/assert"
)

type cpuExecutorMock struct {
	err     error
	isEmpty bool
}

func (c cpuExecutorMock) Info() ([]cpu.InfoStat, error) {
	if c.isEmpty {
		return []cpu.InfoStat{}, nil
	}

	// processors are differentiated by PhysicalID
	// cores within processor are grouped by PhysicalId
	// threads are grouped based on PhysicalId and CoreID
	processors := []cpu.InfoStat{
		// processor with PhysicalId processor1 should have 3 cores and 4 threads
		// core 1 -> 2 threads; core 2 -> 1 thread; core 3 -> 1 thread
		{
			ModelName:  "Processor Name 1",
			VendorID:   "Vendor 1",
			Model:      "Model 1",
			Stepping:   6,
			PhysicalID: "processor1",
			CoreID:     "core1",
			Mhz:        123456,
		},
		{
			ModelName:  "Processor Name 1",
			VendorID:   "Vendor 1",
			Model:      "Model 1",
			Stepping:   6,
			PhysicalID: "processor1",
			CoreID:     "core1",
			Mhz:        123456,
		},
		{
			ModelName:  "Processor Name 1",
			VendorID:   "Vendor 1",
			Model:      "Model 1",
			Stepping:   6,
			PhysicalID: "processor1",
			CoreID:     "core2",
			Mhz:        123456,
		},
		{
			ModelName:  "Processor Name 1",
			VendorID:   "Vendor 1",
			Model:      "Model 1",
			Stepping:   6,
			PhysicalID: "processor1",
			CoreID:     "core3",
			Mhz:        123456,
		},
		// processor2 should have two cores and two threads
		// core 1 -> 1 thread; core 2 -> 1 thread
		{
			ModelName:  "Processor Name 2",
			VendorID:   "Vendor 2",
			Model:      "Model 789",
			Stepping:   4,
			PhysicalID: "processor2",
			CoreID:     "core1",
			Mhz:        478,
		},
		{
			ModelName:  "Processor Name 2",
			VendorID:   "Vendor 2",
			Model:      "Model 789",
			Stepping:   4,
			PhysicalID: "processor2",
			CoreID:     "core2",
			Mhz:        478,
		},
	}

	return processors, c.err
}

var _ cpuExecutor = (*cpuExecutorMock)(nil)

func Test_Generate_AttributesAreGenerated(t *testing.T) {
	sut := provider{
		executor: cpuExecutorMock{},
	}
	ch := sut.Provide()
	container, ok := <-ch
	assert.NotNil(t, ch)
	assert.True(t, ok, "channel should be closed")
	assert.Nil(t, container.Error, "error should be nil")

	processors := container.Processors
	assert.Equal(t, 2, len(processors), "expected count of processors: 2")

	expectedProcessors := map[string]Processor{
		"Processor Name 1": {
			Name:         "Processor Name 1",
			Manufacturer: "Vendor 1",
			Model:        "Model 1",
			Stepping:     "6",
			Cores:        uint32(3),
			Threads:      uint32(4),
			Speed:        float64(123456),
		},
		"Processor Name 2": {
			Name:         "Processor Name 2",
			Manufacturer: "Vendor 2",
			Model:        "Model 789",
			Stepping:     "4",
			Cores:        uint32(2),
			Threads:      uint32(2),
			Speed:        float64(478),
		},
	}

	for _, actual := range processors {
		expected, found := expectedProcessors[actual.Name]
		assert.True(t, found, "actual processor was not found in expected processors")
		assert.EqualValues(t, expected, actual)
	}
}

func Test_Generate_ReturnsEmptyChannelWhenNothingIsProvided(t *testing.T) {
	sut := provider{
		executor: cpuExecutorMock{
			isEmpty: true,
		},
	}
	ch := sut.Provide()
	container, ok := <-ch
	assert.True(t, ok, "channel should be closed")
	assert.Nil(t, container.Error, "error should be nil, when provider sends no data")
	assert.Equal(t, 0, len(container.Processors))
}

func Test_Generate_ReturnsErrorWhenProviderFails(t *testing.T) {
	sut := provider{
		executor: cpuExecutorMock{err: fmt.Errorf("CPU provider error")},
	}
	ch := sut.Provide()
	container, ok := <-ch
	assert.True(t, ok, "channel should be closed even when provider fails")
	assert.Error(t, container.Error, "provider should return error")
	assert.Equal(t, 0, len(container.Processors), "processors should be empty on error")
}
