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

package scraper

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/feature"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/synchronization"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
)

// ManagerConfig represents scraper manager configuration.
type ManagerConfig struct {
	// ScraperConfig is configuration for scraper as a nested property.
	*types.ScraperConfig
	// DelayedProcessingConfig for delayed processing feature.
	*types.DelayedProcessingConfig
}

// Manager prescribes how scraping manager needs to look like.
type Manager interface {
	// Init initialized scraper manager. It consumes scraper declarative
	// description and scraper manager configuration. Returns error on failure,
	// otherwise nil is returned.
	Init(*Descriptor, *ManagerConfig) error
	// Type returns the type of the scraper.
	Type() component.Type
	// Scraper prescription as nested interface, just to allow consumer to
	// omit implementation of scraping callbacks. It can be implemented or
	// re-used by built-in implementation.
	Scraper
}

type manager struct {
	// Feature manager.
	featureManager feature.Manager
	// Scraper scheduler for translation from scraper manager
	// config into scraper runtime.
	scheduler Scheduler
	// scraperRuntime represents current scheduled runtime for
	// scraping based on provided configuration from initialization.
	scraperRuntime *Runtime
	scraperType    component.Type
	config         *ManagerConfig

	logger *zap.Logger
}

var _ Manager = (*manager)(nil)

// Product targeted allocator for scraper manager.
func NewScraperManager(logger *zap.Logger) Manager {
	return createScraperManager(
		feature.NewFeatureManager(logger),
		NewScraperScheduler(),
		logger,
	)
}

func createScraperManager(
	featureManager feature.Manager,
	scheduler Scheduler,
	logger *zap.Logger,
) Manager {
	return &manager{
		featureManager: featureManager,
		scheduler:      scheduler,
		logger:         logger,
	}
}

func (s *manager) Init(
	descriptor *Descriptor,
	config *ManagerConfig,
) error {
	// Create feature management configuration.
	fmConfig := &feature.ManagerConfig{
		ScraperType:             descriptor.Type,
		DelayedProcessingConfig: config.DelayedProcessingConfig,
	}

	// Initialization of feature manager based on constructed
	// configuration.
	err := s.featureManager.Init(fmConfig)
	if err != nil {
		return fmt.Errorf("initialization of feature manager for scraper '%s' failed: %w", descriptor.Type, err)
	}

	// Scheduling scraper internals based on config.
	sRuntime, err := s.scheduler.Schedule(descriptor, config.ScraperConfig, s.logger)
	if err != nil {
		return fmt.Errorf("scheduling in scraper manager for scraper '%s' failed: %w", descriptor.Type, err)
	}

	s.logger.Debug(
		"initialization of scraper manager for scraper finished successfully",
		zap.String("scraper_type", descriptor.Type.String()),
	)

	s.scraperRuntime = sRuntime
	s.scraperType = descriptor.Type
	s.config = config
	return nil
}

// Scrape implements ScraperManager.
func (s *manager) Scrape(
	ctx context.Context,
) (pmetric.Metrics, error) {
	// Break processing when context is closed.
	if synchronization.IsContextClosed(ctx) {
		s.logger.Debug(
			"context is closed breaking processing of scrape for scraper",
			zap.String("scraper_type", s.scraperType.String()),
		)
		return pmetric.NewMetrics(), ctx.Err()
	}

	// Check if internals were properly initilized.
	methodName := "scrape"
	if err := s.checkRuntimeReadiness(methodName); err != nil {
		return pmetric.NewMetrics(), err
	}

	// Check if scraper is ready for processing.
	now := time.Now()
	if !s.featureManager.IsReady(now) {
		s.logger.Debug(
			"scraping skipped for processing",
			zap.String("scraper_type", s.scraperType.String()),
		)
		return pmetric.NewMetrics(), nil
	}

	// Process emit throu scope emitters.
	ch := s.processEmitOnScopeEmitters()

	// Receives and evaluates data from scope metric emitters.
	scopeMetrics, err := s.receiveAndEvaluateEmitResults(ch)
	if err != nil {
		return pmetric.NewMetrics(), err
	}

	// Assembly metrics to final shape.
	metrics, err := s.assemblyMetrics(scopeMetrics)
	if err != nil {
		return pmetric.NewMetrics(), err
	}

	// Update last processed scrape.
	s.featureManager.UpdateLastProcessedTime(now)

	s.logger.Debug(
		"scraping finished",
		zap.String("scraper_type", s.scraperType.String()),
		zap.Int("metric_count", metrics.MetricCount()),
		zap.Int("data_point_count", metrics.DataPointCount()),
	)
	return metrics, nil
}

func (s *manager) assemblyMetrics(
	scopeMetrics []pmetric.ScopeMetricsSlice,
) (pmetric.Metrics, error) {
	if len(scopeMetrics) == 0 {
		return pmetric.NewMetrics(), fmt.Errorf("no scope metrics were gathered for scraper '%s'", s.scraperType)
	}

	// Assembly resource metric.
	ms := pmetric.NewMetrics()
	rms := ms.ResourceMetrics().AppendEmpty()
	rms.SetSchemaUrl(semconv.SchemaURL)

	// Append scope metrics into resource metric.
	for _, sm := range scopeMetrics {
		sm.MoveAndAppendTo(rms.ScopeMetrics())
	}

	s.logger.Debug(
		"assembled resource metric for scraper",
		zap.String("scraper_type", s.scraperType.String()),
		zap.Int("scope_metrics_count", rms.ScopeMetrics().Len()),
	)
	return ms, nil
}

type emitResultChannel = chan *scope.Result

func (s *manager) processEmitOnScopeEmitters() emitResultChannel {
	// Creation of communication channel
	seCount := len(s.scraperRuntime.ScopeEmitters)
	ch := make(emitResultChannel, seCount)

	emittersWg := new(sync.WaitGroup)
	emittersWg.Add(seCount)

	for name, emitter := range s.scraperRuntime.ScopeEmitters {
		s.logger.Debug(
			"starting emit on scope emitter",
			zap.String("emitter_name", name),
		)

		go func(ch emitResultChannel, emitter scope.Emitter) {
			defer emittersWg.Done()

			res := emitter.Emit()
			ch <- res
		}(ch, emitter)
	}

	emittersWg.Wait()
	close(ch)

	return ch
}

func (s *manager) receiveAndEvaluateEmitResults(
	ch emitResultChannel,
) ([]pmetric.ScopeMetricsSlice, error) {
	errs := make([]error, 0)
	data := make([]pmetric.ScopeMetricsSlice, 0)

	// Receive init results.
	for sr := range ch {
		if sr.Error != nil {
			// Result was error.
			errs = append(errs, sr.Error)
		} else {
			// Result brought data.
			data = append(data, sr.Data)
		}
	}

	// Evaluate init results.
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return make([]pmetric.ScopeMetricsSlice, 0), fmt.Errorf("emit action on scope emitters for scraper '%s' failed: %w", s.scraperType, err)
	}

	return data, nil
}

// Start implements ScraperManager.
func (s *manager) Start(
	ctx context.Context,
	_ component.Host,
) error {
	// Break processing when context is closed.
	if synchronization.IsContextClosed(ctx) {
		s.logger.Debug(
			"context is closed, breaking processing of start for scraper",
			zap.String("scraper_type", s.scraperType.String()),
		)
		return ctx.Err()
	}

	// Check if internals were properly initialized.
	methodName := "start"
	if err := s.checkRuntimeReadiness(methodName); err != nil {
		return err
	}

	// Activate inits on scope emitters in parallel.
	ch := s.processInitOnScopeEmitters()

	// Receive and evaluate results from scope emitters.
	if err := s.receiveAndEvaluateInitResults(ch); err != nil {
		return err
	}

	// Initialize delayed processing through feature manager
	// proxy.
	if err := s.featureManager.InitDelayedProcessing(
		s.config.DelayedProcessingConfig,
		time.Now(),
	); err != nil {
		return err
	}

	s.logger.Debug(
		"scraper manager start finished successfully",
		zap.String("scraper_type", s.scraperType.String()),
	)
	return nil
}

type initResultChannel = chan error

func (s *manager) processInitOnScopeEmitters() initResultChannel {
	// Creation of communication channel.
	seCount := len(s.scraperRuntime.ScopeEmitters)
	ch := make(initResultChannel, seCount)

	emittersWg := new(sync.WaitGroup)
	emittersWg.Add(seCount)

	// Spin scope emitters.
	for name, emitter := range s.scraperRuntime.ScopeEmitters {
		s.logger.Debug("starting init on scope emitter", zap.String("emitter_name", name))

		go func(ch initResultChannel, emitter scope.Emitter) {
			defer emittersWg.Done()

			err := emitter.Init()
			ch <- err
		}(ch, emitter)
	}

	emittersWg.Wait()
	close(ch)

	return ch
}

func (s *manager) receiveAndEvaluateInitResults(ch initResultChannel) error {
	errs := make([]error, 0)

	// Receive init results.
	for err := range ch {
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Evaluate init results.
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return fmt.Errorf("init action on scope emitters for scraper '%s' failed: %w", s.scraperType, err)
	}

	return nil
}

// Type implements ScraperManager.
func (s *manager) Type() component.Type {
	return s.scraperType
}

// Shutdown implements ScraperManager.
func (s *manager) Shutdown(context.Context) error {
	return nil
}

func (s *manager) checkRuntimeReadiness(
	pointOfOrigin string,
) error {
	if s.scraperRuntime == nil {
		return fmt.Errorf("scraper manager at '%s' for scraper '%s' is not ready with runtime", pointOfOrigin, s.scraperType)
	}
	return nil
}
