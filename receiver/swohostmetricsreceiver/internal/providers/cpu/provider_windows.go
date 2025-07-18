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
	"strings"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/tools/wmi"
)

type provider struct {
	wmi wmi.Executor
}

var _ providers.Provider[Container] = (*provider)(nil)

func CreateProvider() providers.Provider[Container] {
	return &provider{
		wmi: wmi.NewExecutor(),
	}
}

// Provide implements providers.Provider.
func (p *provider) Provide() <-chan Container {
	ch := make(chan Container)
	go func() {
		defer close(ch)
		results, err := wmi.QueryResult[[]Win32_Processor](p.wmi)
		if err != nil {
			ch <- Container{Error: err}
			return
		}
		processors := make([]Processor, 0, len(results))
		for _, result := range results {
			processors = append(processors, wmiProcessorToProcessor(result))
		}
		ch <- Container{Processors: processors}
	}()
	return ch
}

func wmiProcessorToProcessor(result Win32_Processor) Processor {
	processor := Processor{
		Name:         result.Name,
		Manufacturer: result.Manufacturer,
		Speed:        float64(result.CurrentClockSpeed),
		Cores:        result.NumberOfCores,
		Threads:      result.NumberOfLogicalProcessors,
		Stepping:     result.Stepping,
	}
	return amendFromCaption(processor, result.Caption)
}

func amendFromCaption(processor Processor, caption string) Processor {
	fields := strings.Fields(caption)
	for i, field := range fields {
		if strings.EqualFold(field, "Model") && len(fields) > i {
			processor.Model = fields[i+1]
		}
		if processor.Stepping == "" && strings.EqualFold(field, "Stepping") && len(fields) > i {
			processor.Stepping = fields[i+1]
		}

	}
	return processor
}

// Win32_Processor represents actual Processor WMI Object
// with subset of fields required for scraping.
type Win32_Processor struct {
	Name                      string
	Manufacturer              string
	CurrentClockSpeed         uint32
	NumberOfCores             uint32
	NumberOfLogicalProcessors uint32
	Stepping                  string
	Caption                   string
}
