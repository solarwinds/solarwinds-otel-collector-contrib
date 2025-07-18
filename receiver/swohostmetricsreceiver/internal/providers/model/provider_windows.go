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
	"sync"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/synchronization"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/tools/wmi"
	"go.uber.org/zap"
)

type computerSystem struct {
	Model        string
	Manufacturer string
}

type bios struct {
	SerialNumber string
}

type provider struct {
	wmi wmi.Executor
}

// Win32_ComputerSystem represents actual Computer System WMI Object
// with subset of fields required for scraping.
type Win32_ComputerSystem struct {
	Model        string
	Manufacturer string
}

// Win32_BIOS represents actual BIOS WMI Object
// with subset of fields required for scraping.
type Win32_BIOS struct {
	SerialNumber string
}

var _ providers.Provider[Model] = (*provider)(nil)

func CreateModelProvider() providers.Provider[Model] {
	return &provider{
		wmi: wmi.NewExecutor(),
	}
}

// Provide implements Provider.
func (p *provider) Provide() <-chan Model {
	ch := make(chan Model)
	go p.provideInternal(ch)
	return ch
}

func (p *provider) provideInternal(ch chan Model) {
	defer close(ch)

	var wg sync.WaitGroup

	// spin workers
	wg.Add(2)
	csCh := p.loadComputerSystem()
	bCh := p.loadBios()

	// synchronization channel for breaking endless cycle
	terminationCh := synchronization.ActivateSupervisingRoutine(&wg)

	var model Model
loop:
	for {
		select {
		case cs, ok := <-csCh:
			if ok {
				model.Manufacturer = cs.Manufacturer
				model.Name = cs.Model
			} else {
				csCh = nil // remove it from select choices
				wg.Done()
			}
		case b, ok := <-bCh:
			if ok {
				model.SerialNumber = b.SerialNumber
			} else {
				bCh = nil // remove it from select choices
				wg.Done()
			}
		case <-terminationCh:
			break loop
		}
	}

	zap.L().Debug(fmt.Sprintf("Model provider result: %+v", model))

	ch <- model
}

func (p *provider) loadComputerSystem() chan computerSystem {
	ch := make(chan computerSystem)
	go func() {
		defer close(ch)
		result, err := wmi.QuerySingleResult[Win32_ComputerSystem](p.wmi)
		if err == nil {
			ch <- computerSystem{Model: result.Model, Manufacturer: result.Manufacturer}
		}
	}()
	return ch
}

func (p *provider) loadBios() chan bios {
	ch := make(chan bios)
	go func() {
		defer close(ch)
		result, err := wmi.QuerySingleResult[Win32_BIOS](p.wmi)
		if err == nil {
			ch <- bios{SerialNumber: result.SerialNumber}
		}
	}()
	return ch
}
