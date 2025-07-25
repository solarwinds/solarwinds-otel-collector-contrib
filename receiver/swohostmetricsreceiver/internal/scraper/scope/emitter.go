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

package scope

import (
	"errors"
	"fmt"
	"sync"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// emitter struct with Name and MetricEmitters.
type emitter struct {
	Name           string
	MetricEmitters map[string]types.MetricEmitterInterface
}

// Emitter interface for scope emitter.
type Emitter interface {
	Init() error
	Emit() Result
}

// Result contains merged results from all metric emitters,
// with scope Name and possible Error from emits.
type Result struct {
	Name    string
	Metrics pmetric.MetricSlice
	Error   error
}

var _ Emitter = (*emitter)(nil)

// NewScopeEmitter creates scope emitter object in case at least one metric contained
// in this scope is enabled in the scraper config (metrics). If metric is enabled the metric emitter
// is added to MetricEmitters to be later initialized and called for emit result.
// If there is no metric emitter to be used the function returns nil.
func NewScopeEmitter(scopeName string, metricEmitters map[string]types.MetricEmitterInterface, metrics map[string]types.MetricSettingsConfig) Emitter {
	se := createScopeEmitter(scopeName)
	for metricName, metricEmitter := range metricEmitters {
		setting, found := metrics[metricName]
		if !found {
			continue
		}
		if setting.Enabled {
			se.MetricEmitters[metricName] = metricEmitter
		}
	}

	if len(se.MetricEmitters) == 0 {
		return nil
	}
	return se
}

// createScopeEmitter will return new empty emitter.
func createScopeEmitter(scopeName string) *emitter {
	return &emitter{
		Name:           scopeName,
		MetricEmitters: make(map[string]types.MetricEmitterInterface),
	}
}

// Init initializes all scheduled metric emitters.
func (s *emitter) Init() error {
	errCh := make(chan error, len(s.MetricEmitters))

	var wg sync.WaitGroup
	wg.Add(len(s.MetricEmitters))

	// launch initialization routines
	for _, e := range s.MetricEmitters {
		go func(e types.MetricEmitterInterface) {
			defer wg.Done()

			err := e.Initialize()
			errCh <- err
		}(e)
	}

	wg.Wait()
	close(errCh)

	var err error
	for e := range errCh {
		err = errors.Join(err, e)
	}

	if err != nil {
		message := fmt.Sprintf(
			"emitter initialization failed with following errors %v",
			err,
		)
		return errors.New(message)
	}

	return nil
}

// Emit calls emit on all scheduled metric emitters. Resulting data from
// metric emitters are merged and returned as Result. Result contains scope name,
// possible errors and the metrics are propagated as one pmetric.MetricSlice built from
// obtained pmetric.MetricSlice of each metric emitter.
func (s *emitter) Emit() Result {
	resCh := s.spinMetricEmitters()
	data := s.assemblyMetricEmittersResults(resCh)

	return data
}

func (s *emitter) spinMetricEmitters() chan metric.Result {
	metricsCh := make(chan metric.Result, len(s.MetricEmitters))
	defer close(metricsCh)

	var wg sync.WaitGroup
	wg.Add(len(s.MetricEmitters))

	for _, e := range s.MetricEmitters {
		go func(me types.MetricEmitterInterface) {
			defer wg.Done()

			emit := me.GetEmittingFunction()
			data, err := emit()

			metricsCh <- metric.Result{
				MetricSlice: data,
				Error:       err,
			}
		}(e)
	}
	wg.Wait()
	return metricsCh
}

func (s *emitter) assemblyMetricEmittersResults(ch chan metric.Result) Result {
	metricSlice := pmetric.NewMetricSlice()
	metricSlice.EnsureCapacity(len(s.MetricEmitters))

	var err error

	for e := range ch {
		if e.Error != nil {
			err = errors.Join(e.Error)
			continue
		}
		e.MetricSlice.MoveAndAppendTo(metricSlice)
	}

	return Result{
		Name:    s.Name,
		Metrics: metricSlice,
		Error:   err,
	}
}
