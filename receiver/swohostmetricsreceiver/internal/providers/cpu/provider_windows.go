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
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/wmi"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"go.uber.org/zap"
)

// cpuInfoExecutor abstracts gopsutil cpu.Info() for testability.
type cpuInfoExecutor interface {
	Info() ([]cpu.InfoStat, error)
}

type cpuInfoExecutorImpl struct{}

func (c *cpuInfoExecutorImpl) Info() ([]cpu.InfoStat, error) {
	return cpu.Info()
}

var _ cpuInfoExecutor = (*cpuInfoExecutorImpl)(nil)

type provider struct {
	wmi      wmi.Executor
	cpuInfo  cpuInfoExecutor
	countsFn func(bool) (int, error)
	logger   *zap.Logger
}

var _ providers.Provider[Container] = (*provider)(nil)

func CreateProvider(logger *zap.Logger) providers.Provider[Container] {
	return &provider{
		wmi:      wmi.NewExecutor(),
		cpuInfo:  &cpuInfoExecutorImpl{},
		countsFn: cpu.Counts,
		logger:   logger,
	}
}

// Provide implements providers.Provider.
func (p *provider) Provide() <-chan Container {
	ch := make(chan Container)
	go func() {
		defer close(ch)

		infos, err := p.cpuInfo.Info()
		if err != nil {
			ch <- Container{Error: err}
			return
		}
		if len(infos) == 0 {
			ch <- Container{}
			return
		}

		wmiResults, err := wmi.QueryResult[[]Win32_Processor](p.wmi)
		if err != nil {
			ch <- Container{Error: err}
			return
		}

		totalLogical, err := p.countsFn(true)
		if err != nil {
			ch <- Container{Error: err}
			return
		}
		if totalLogical <= 0 {
			ch <- Container{Error: fmt.Errorf("cpu counts returned non-positive value: %d", totalLogical)}
			return
		}

		threadsPerSocket := uint32(totalLogical) / uint32(len(infos))

		// WHY: WMI rows are uniform for server hardware — all sockets expose identical
		// descriptive fields (Name, Manufacturer, Speed, Cores). Using the first row
		// for every socket is intentional; a zero-value descriptor is used when WMI
		// returns no rows, producing empty-but-valid Processor entries.
		var descriptor Win32_Processor
		if len(wmiResults) > 0 {
			descriptor = wmiResults[0]
		}

		processors := make([]Processor, 0, len(infos))
		for i := range infos {
			p := buildProcessor(i, descriptor, threadsPerSocket)
			processors = append(processors, p)
		}

		ch <- Container{Processors: processors}
	}()
	return ch
}

func buildProcessor(index int, descriptor Win32_Processor, threadsPerSocket uint32) Processor {
	proc := Processor{
		DeviceID: fmt.Sprintf("CPU%d", index),
		Threads:  threadsPerSocket,
		Cores:    descriptor.NumberOfCores,
	}
	proc = applyDescriptiveFields(proc, descriptor)
	return amendFromCaption(proc, descriptor.Caption)
}

func applyDescriptiveFields(proc Processor, d Win32_Processor) Processor {
	proc.Name = d.Name
	proc.Manufacturer = d.Manufacturer
	proc.Speed = float64(d.CurrentClockSpeed)
	proc.Stepping = d.Stepping
	return proc
}

func amendFromCaption(processor Processor, caption string) Processor {
	fields := strings.Fields(caption)
	for i, field := range fields {
		if strings.EqualFold(field, "Model") && len(fields) > i+1 {
			processor.Model = fields[i+1]
		}
		if processor.Stepping == "" && strings.EqualFold(field, "Stepping") && len(fields) > i+1 {
			processor.Stepping = fields[i+1]
		}
	}
	return processor
}

// Win32_Processor represents actual Processor WMI Object
// with subset of fields required for scraping.
type Win32_Processor struct {
	Name              string
	Manufacturer      string
	CurrentClockSpeed uint32
	NumberOfCores     uint32
	Stepping          string
	Caption           string
}
