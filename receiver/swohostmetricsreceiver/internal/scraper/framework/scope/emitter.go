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

	"go.opentelemetry.io/collector/pdata/pmetric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/version"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
)

// Result of scope emitter. Data are provided on success.
// Error is provided on failure.
type Result struct {
	Data  pmetric.ScopeMetricsSlice
	Error error
}

// Emitter interface prescribes how scope emitter needs to
// implemented.
type Emitter interface {
	// Init initializes scope emitter. Error is returned on
	// failure, otherwise nil is returned.
	Init() error
	// Emit emits scope emitter pointer to result.
	Emit() *Result
	// Name returns scope name of emitter.
	Name() string
}

type metricEmitterEmitResult struct {
	Result      *metric.Result
	EmitterName string
}

type metricEmitterInitResult struct {
	Name  string
	Error error
}

// scopeEmitter is general implementation for scope emitter.
// It manages processing of metric emitters in his ownership.
type emitter struct {
	scopeName      string
	metricEmitters map[string]metric.Emitter
	logger         *zap.Logger
}

var _ Emitter = (*emitter)(nil)

func CreateDefaultScopeEmitter(
	scopeName string,
	metricEmitters map[string]metric.Emitter,
	logger *zap.Logger,
) Emitter {
	return &emitter{
		scopeName:      scopeName,
		metricEmitters: metricEmitters,
		logger:         logger,
	}
}

// Emit implements ScopeEmitter.
func (s *emitter) Emit() *Result {
	if err := s.checkIfMetricEmittersAreRegistered(); err != nil {
		return createErrorScopeResult(err)
	}

	// Processes emit for all emitters.
	scopeMetric, err := s.processMetricEmitters()
	if err != nil {
		return createErrorScopeResult(err)
	}

	s.logger.Debug("emit of scope emitter finished successfully", zap.String("scopeName", s.scopeName))
	return &Result{scopeMetric, nil}
}

func (s *emitter) processMetricEmitters() (pmetric.ScopeMetricsSlice, error) {
	meCount := len(s.metricEmitters)

	// Making barrier.
	emitWg := new(sync.WaitGroup)
	emitWg.Add(meCount)

	// Making channel, wide enough for accommodate
	// all emitters
	rCh := make(chan *metricEmitterEmitResult, meCount)

	for meName, me := range s.metricEmitters {
		go processMetricEmitter(emitWg, rCh, meName, me, s.logger)
	}

	// Wait until all emitters are done.
	emitWg.Wait()
	// Close channel, there is no more data to send.
	close(rCh)

	// Process & evaluate emitted results.
	ms, err := processEmittedResults(rCh, s.logger)
	if err != nil {
		return pmetric.NewScopeMetricsSlice(), err
	}

	// Assembly scope metric slice.
	sms, err := s.assemblyScopeMetricSlice(ms)
	if err != nil {
		return pmetric.NewScopeMetricsSlice(), err
	}

	return sms, nil
}

func processMetricEmitter(
	emitWg *sync.WaitGroup,
	rCh chan *metricEmitterEmitResult,
	meName string,
	me metric.Emitter,
	logger *zap.Logger,
) {
	defer emitWg.Done()

	logger.Debug("emitting of metric emitter", zap.String("metric", meName))
	mr := me.Emit()
	rCh <- &metricEmitterEmitResult{mr, meName}
}

func processEmittedResults(
	rCh chan *metricEmitterEmitResult,
	logger *zap.Logger,
) ([]pmetric.MetricSlice, error) {
	errs := make([]error, 0)
	mrs := make([]pmetric.MetricSlice, 0)

	// Collect results.
	for r := range rCh {
		if r.Result.Error != nil {
			message := fmt.Sprintf(
				"emit for metric emitter '%s' failed",
				r.EmitterName,
			)
			logger.Error(message, zap.Error(r.Result.Error))
			errs = append(errs, fmt.Errorf("%s: %w", message, r.Result.Error))
		} else {
			mrs = append(mrs, r.Result.Data)
		}
	}

	// Evaluate errors, then pack into single one.
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return mrs, nil
}

func (s *emitter) assemblyScopeMetricSlice(
	ms []pmetric.MetricSlice,
) (pmetric.ScopeMetricsSlice, error) {
	if len(ms) == 0 {
		message := fmt.Sprintf(
			"no metric slices available for scope emitter '%s'",
			s.scopeName,
		)
		s.logger.Error(message)
		return pmetric.NewScopeMetricsSlice(), errors.New(message)
	}

	sms := pmetric.NewScopeMetricsSlice()

	// Configure scope metric attributes.
	sm := sms.AppendEmpty()
	sm.SetSchemaUrl(semconv.SchemaURL)
	sm.Scope().SetName(s.scopeName)
	sm.Scope().SetVersion(version.Version)

	// Setup metrics into scope metric slice.
	for _, m := range ms {
		m.MoveAndAppendTo(sm.Metrics())
	}

	s.logger.Debug(
		"assembled scope metric finished successfully",
		zap.String("scope_name", s.scopeName),
		zap.Int("metric_count", sm.Metrics().Len()),
	)
	return sms, nil
}

func createErrorScopeResult(err error) *Result {
	return &Result{
		Data:  pmetric.ScopeMetricsSlice{},
		Error: err,
	}
}

// Init implements ScopeEmitter.
func (s *emitter) Init() error {
	if err := s.checkIfMetricEmittersAreRegistered(); err != nil {
		return err
	}

	if err := s.initializeMetricEmitters(); err != nil {
		message := fmt.Sprintf(
			"initialization of metric emitters in scope emitter '%s' failed",
			s.scopeName,
		)
		s.logger.Error(message, zap.Error(err))
		return fmt.Errorf("%s: %w", message, err)
	}

	s.logger.Debug("scope emitter initialized successfully", zap.String("scope_name", s.scopeName))
	return nil
}

func (s *emitter) checkIfMetricEmittersAreRegistered() error {
	// Check if at least some metric emitter is registered.
	// If not, there is no reason to run Init() or even has
	// this emitter created.
	if len(s.metricEmitters) == 0 {
		message := fmt.Sprintf(
			"scope emitter for '%s' is initialized and has no registered metric emitters",
			s.scopeName,
		)
		s.logger.Error(message)
		return fmt.Errorf("%s", message)
	}

	return nil
}

func (s *emitter) initializeMetricEmitters() error {
	meCount := len(s.metricEmitters)

	// Making barrier.
	initWg := new(sync.WaitGroup)
	initWg.Add(meCount)

	rCh := make(chan *metricEmitterInitResult, meCount)

	for name, me := range s.metricEmitters {
		go initMetricEmitter(initWg, rCh, name, me, s.logger)
	}

	// Wait until all emitters are initialized.
	initWg.Wait()
	// Close channel when all emitents are done.
	close(rCh)

	return evaluateInitResults(rCh, s.logger)
}

func evaluateInitResults(rCh chan *metricEmitterInitResult, logger *zap.Logger) error {
	errs := make([]error, 0)

	// Collect results.
	for r := range rCh {
		if r.Error != nil {
			message := fmt.Sprintf(
				"initialization for metric emitter '%s' failed",
				r.Name,
			)
			logger.Error(message, zap.Error(r.Error))
			errs = append(errs, fmt.Errorf("%s: %w", message, r.Error))
		}
	}

	// Evaluate, then pack into single error.
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func initMetricEmitter(
	wg *sync.WaitGroup,
	rCh chan *metricEmitterInitResult,
	meName string,
	me metric.Emitter,
	logger *zap.Logger,
) {
	defer wg.Done()

	logger.Debug("initializing of metric emitter", zap.String("metric", meName))

	err := me.Init()
	rCh <- &metricEmitterInitResult{meName, err}
}

// Name implements ScopeEmitter.
func (s *emitter) Name() string {
	return s.scopeName
}
