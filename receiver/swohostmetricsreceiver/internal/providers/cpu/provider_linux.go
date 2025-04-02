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
	"strconv"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers"
)

type cpuExecutor interface {
	Info() ([]cpu.InfoStat, error)
}

type executor struct{}

func (e *executor) Info() ([]cpu.InfoStat, error) {
	return cpu.Info()
}

var _ cpuExecutor = (*executor)(nil)

type provider struct {
	executor cpuExecutor
}

func CreateProvider() providers.Provider[Container] {
	return &provider{
		executor: &executor{},
	}
}

// Provide implements providers.Provider.
func (p *provider) Provide() <-chan Container {
	ch := make(chan Container)
	go func() {
		defer close(ch)

		infoStats, err := p.executor.Info()
		if err != nil {
			ch <- Container{Error: err}
			return
		}

		processors := make(map[string]Processor, len(infoStats))
		cores := make(map[CoreKey]bool, len(infoStats))
		for _, infoStat := range infoStats {
			processor, exists := processors[infoStat.PhysicalID]
			if !exists {
				processor = Processor{
					Name:         infoStat.ModelName,
					Manufacturer: infoStat.VendorID,
					Speed:        infoStat.Mhz,
					Cores:        0,
					Threads:      0,
					Model:        infoStat.Model,
					Stepping:     strconv.FormatInt(int64(infoStat.Stepping), 10),
				}
			}

			// InfoStat for Linux is per virtual CPU (thread), so cores and threads have to be counted manually
			processor.Threads++
			coreKey := CoreKey{infoStat.PhysicalID, infoStat.CoreID}
			if _, exists := cores[coreKey]; !exists {
				cores[coreKey] = true
				processor.Cores++
			}

			processors[infoStat.PhysicalID] = processor
		}

		ch <- Container{Processors: toSlice(processors)}
	}()
	return ch
}

func toSlice[T any](m map[string]T) []T {
	list := make([]T, 0, len(m))
	for _, v := range m {
		list = append(list, v)
	}
	return list
}

type CoreKey struct {
	ProcessorID string
	CoreID      string
}
