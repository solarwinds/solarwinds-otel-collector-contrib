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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/framework/feature"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/types"
)

const once = int32(1)

func Test_Example_HowToFillResourceMetric(t *testing.T) {
	// product of scope emitter
	sms := pmetric.NewScopeMetricsSlice()
	sm := sms.AppendEmpty()
	sm.SetSchemaUrl("kokoha/schema")
	sm.Scope().SetName("kokoha.scope.metric")
	sm.Scope().SetVersion("1.0.0")

	// writing scope metric into resource metric
	rm := pmetric.NewResourceMetrics()
	sms.MoveAndAppendTo(rm.ScopeMetrics())

	assert.Equal(t, 1, rm.ScopeMetrics().Len(), "There must be exactly on scope metric")
	assert.Equal(t, "kokoha/schema", rm.ScopeMetrics().At(0).SchemaUrl(), "Schema must be the same")
}

func Test_Example_BasicManagerUsage(t *testing.T) {
	testingContext := context.TODO()
	testingScraper := scraper.RequireComponentType(t, scraperName)

	scraperDefinition := &Descriptor{
		Type: testingScraper,
		ScopeDescriptors: map[string]scope.Descriptor{
			scope1: {
				ScopeName: scope1,
				MetricDescriptors: map[string]metric.Descriptor{
					scope1metric1: {
						Create: CreateMetricEmitter1a,
					},
					scope1metric2: {
						Create: CreateMetricEmitter1b,
					},
					scope1metric3: {
						Create: CreateMetricEmitter1c,
					},
				},
				// Without custom scope alocator - most of the cases.
				// Default one is expect to be taken.
			},
			scope2: {
				ScopeName: scope2,
				// Allocator callback for creation of custom scope emitter
				// for this particullar scope - scope2. In case we need to
				// create something with totaly custom implementation.
				Create: scope.CreateCustomScopeEmitter,
				MetricDescriptors: map[string]metric.Descriptor{
					scope2metric1: {
						Create: CreateMetricEmitter2a,
					},
					scope2metric2: {
						Create: CreateMetricEmitter2b,
					},
				},
			},
		},
	}

	// Scraper config.
	scraperConfig := &types.ScraperConfig{
		Metrics: map[string]types.MetricSettingsConfig{
			scope1metric1: {Enabled: true},
			// Will be removed from final runtime => not enabled.
			scope1metric2: {Enabled: false},
			// Will be removed from final runtime => not enabled.
			scope1metric3: {Enabled: false},
			scope2metric1: {Enabled: true},
			scope2metric2: {Enabled: true},
		},
	}

	scraperManagerConfigNoDP := &ManagerConfig{
		ScraperConfig: scraperConfig,
	}

	// Factory content: create scraper. When scraper utilizes scraping manager
	// following lines pay also for scraper, not only for scrapping manager.
	sut := NewScraperManager()

	// Factory content: initialize it.
	err := sut.Init(scraperDefinition, scraperManagerConfigNoDP)
	assert.Nil(t, err, "init must not fail")

	// OTEL Collector processing: Start aka init of scrapping.
	err = sut.Start(testingContext, nil)
	assert.Nil(t, err, "start must not fail")

	// OTEL Collector processing: Scraping data periodically.
	ms, err := sut.Scrape(testingContext)
	assert.Nil(t, err, "scrape must not fail")
	assert.Equal(t, 3, ms.MetricCount(), "result must have 3 metrics")
	assert.Equal(t, 7, ms.DataPointCount(), "result must have 7 datapoints")
}

func Test_Init_OnFailingFeatureManagerInitFails(t *testing.T) {
	eMessage := errors.New("feature manager init failed")
	fm := feature.CreateFeatureManagerMock(eMessage, nil, false)
	sut := createScraperManager(fm, nil)

	descriptor := createArtificialScraperDescriptor(t)
	config := createEmptyManagerConfig()

	err := sut.Init(descriptor, config)
	assert.Error(t, err, "init with failing feature manager must fail")
	assert.ErrorContains(t, err, "")
	assert.Equal(t, once, fm.InitCC.GetCallsCount(), "init must be called exactly once")
}

func Test_Init_OnFailingSchedulerInitFails(t *testing.T) {
	eMessage := errors.New("scheduling failed")
	fm := feature.CreateFeatureManagerMock(nil, nil, false)
	sm := CreateSchedulerMock(eMessage, nil)
	sut := createScraperManager(fm, sm)

	descriptor := createArtificialScraperDescriptor(t)
	config := createEmptyManagerConfig()
	err := sut.Init(descriptor, config)

	assert.Error(t, err, "init must fail")
	assert.ErrorContains(t, err, eMessage.Error(), "message must be found")
	assert.Equal(t, once, fm.InitCC.GetCallsCount(), "init must be called exactly once")
}

func Test_Init_InitSucceedsWhenAllComponentsSucceeds(t *testing.T) {
	fm := feature.CreateFeatureManagerMock(nil, nil, false)
	sm := CreateSchedulerMock(nil, nil)
	sut := createScraperManager(fm, sm)

	descriptor := createArtificialScraperDescriptor(t)
	config := createEmptyManagerConfig()
	err := sut.Init(descriptor, config)

	assert.NoError(t, err, "init must succeed")
	assert.Equal(t, once, fm.InitCC.GetCallsCount(), "init on feature manager must be called")
	assert.Equal(t, once, sm.scheduleCC.GetCallsCount(), "schedule on scheduler must be called")
}

func Test_Scrape_OnClosedContext_ScrapeFailsNoMetricsEmitted(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	sut := createScraperManager(nil, nil)
	ms, err := sut.Scrape(ctx)

	assert.Zero(t, ms.MetricCount(), "result must have no metrics")
	assert.Error(t, err, "error must occur")
}

func Test_Scrape_OnNonInitManager_ScrapeFailsNoMetricsEmitted(t *testing.T) {
	ctx := context.Background()

	sut := &manager{
		scraperRuntime: nil, // runtime is not initialized due to skipped manager init
	}
	ms, err := sut.Scrape(ctx)

	assert.Zero(t, ms.MetricCount(), "no metric is returned on failure")
	assert.Error(t, err, "failure must be signeld with error")
	assert.ErrorContains(t, err, "is not ready with runtime")
}

func Test_Scrape_OnOneFailedEmit_ScrapeFailsAndNoMetricsEmitted(t *testing.T) {
	ctx := context.Background()

	errExpected := errors.New("scope2 emitter failed")
	sms := pmetric.NewScopeMetricsSlice()
	sms.AppendEmpty()

	testingScraper := scraper.RequireComponentType(t, "test_scraper")
	sut := &manager{
		scraperRuntime: &Runtime{
			ScopeEmitters: map[string]scope.Emitter{
				"scope1": scope.CreateEmitterMock(
					&scope.Result{
						Data:  sms,
						Error: nil,
					},
					nil, ""),
				"scope2": scope.CreateEmitterMock(
					&scope.Result{
						Data:  pmetric.NewScopeMetricsSlice(),
						Error: errExpected,
					},
					nil, ""),
			},
		},
		scraperType:    testingScraper,
		config:         &ManagerConfig{},
		featureManager: feature.CreateFeatureManagerMock(nil, nil, true),
	}
	ms, err := sut.Scrape(ctx)

	assert.Error(t, err, "when even one emitter fail scrape result must be error")
	assert.ErrorContains(t, err, "emit action on scope emitters for scraper 'test_scraper' failed")
	assert.Zero(t, ms.ResourceMetrics().Len(), "no resource metrics available on failure")
	assert.Zero(t, ms.MetricCount(), "no metrics at all")
}

func Test_Scrape_OnSuccessfulEmit_ScrapeSucceedsAndMetricsAreProvided(t *testing.T) {
	ctx := context.Background()
	// scope 1
	sms1 := pmetric.NewScopeMetricsSlice()
	sm1 := sms1.AppendEmpty()
	// scope 1 metric 1
	s1m1 := sm1.Metrics().AppendEmpty()
	s1m1.SetName("s1.m1.whatever")
	// scope 1 metric 2
	s1m2 := sm1.Metrics().AppendEmpty()
	s1m2.SetName("s1.m2.whatever")
	// scope 2
	sms2 := pmetric.NewScopeMetricsSlice()
	sm2 := sms2.AppendEmpty()
	// scope 2 metric 1
	s2m1 := sm2.Metrics().AppendEmpty()
	s2m1.SetName("s2.m1.whatever")
	// scope 2 metric 2
	s2m2 := sm2.Metrics().AppendEmpty()
	s2m2.SetName("s2.m2.whatever")

	testingScraper := scraper.RequireComponentType(t, "test_scraper")

	sut := &manager{
		scraperRuntime: &Runtime{
			ScopeEmitters: map[string]scope.Emitter{
				"scope1": scope.CreateEmitterMock(
					&scope.Result{
						Data:  sms1,
						Error: nil,
					},
					nil, ""),
				"scope2": scope.CreateEmitterMock(
					&scope.Result{
						Data:  sms2,
						Error: nil,
					},
					nil, ""),
			},
		},
		scraperType:    testingScraper,
		config:         &ManagerConfig{},
		featureManager: feature.CreateFeatureManagerMock(nil, nil, true),
	}
	ms, err := sut.Scrape(ctx)

	assert.NoError(t, err, "scrape must succeed")
	assert.Equal(t, 4, ms.MetricCount(), "metric must be available")
}

func createArtificialScraperDescriptor(t *testing.T) *Descriptor {
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	return &Descriptor{
		Type:             testingScraper,
		ScopeDescriptors: map[string]scope.Descriptor{},
	}
}

func createEmptyManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		ScraperConfig:           &types.ScraperConfig{},
		DelayedProcessingConfig: &types.DelayedProcessingConfig{},
	}
}

func Test_Start_FailsOnClosedContext(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	sut := NewScraperManager()
	err := sut.Start(ctx, nil)

	assert.Error(t, err, "on closed context start must fail")
	assert.ErrorContains(t, err, "context canceled")
}

func Test_Start_FailsOnUninitManager(t *testing.T) {
	sut := NewScraperManager()
	err := sut.Start(context.Background(), nil)

	assert.Error(t, err, "uninitialized manager must fail")
	assert.ErrorContains(t, err, "scraper manager at 'start' for scraper '' is not ready with runtime")
}

func Test_Start_FailsWhenAtLeastOneScopeEmitterFailsOnInit(t *testing.T) {
	expectedError := errors.New("failing init in scope emitter scope2")
	testingScraper := scraper.RequireComponentType(t, "test_scraper")

	sut := &manager{
		scraperRuntime: &Runtime{
			ScopeEmitters: map[string]scope.Emitter{
				// succeeding scope emitter
				"scope1": scope.CreateEmitterMock(nil, nil, "scope1"),
				// failing scope emitter in init
				"scope2": scope.CreateEmitterMock(
					nil,
					expectedError,
					"scope2"),
			},
		},
		scraperType: testingScraper,
	}
	err := sut.Start(context.Background(), nil)

	assert.Error(t, err, "start must fail when one scope emitter fails to init")
	assert.ErrorContains(t, err, "failing init in scope emitter scope2")
}

func Test_Start_FailsOnFailedDelayedProcessingInit(t *testing.T) {
	expectedError := errors.New("init of delayed processing in feature manager failed")
	testingScraper := scraper.RequireComponentType(t, "test_scraper")

	fmm := feature.CreateFeatureManagerMock(
		nil,
		expectedError, true)
	sut := &manager{
		scraperRuntime: &Runtime{
			ScopeEmitters: map[string]scope.Emitter{
				"scope1": scope.CreateEmitterMock(nil, nil, "scope1"),
				"scope2": scope.CreateEmitterMock(nil, nil, "scope2"),
			},
		},
		scraperType:    testingScraper,
		featureManager: fmm,
		config: &ManagerConfig{
			DelayedProcessingConfig: &types.DelayedProcessingConfig{
				CollectionInterval: 10,
			},
		},
	}
	err := sut.Start(context.Background(), nil)

	assert.Error(t, err, "start must fail")
	assert.ErrorContains(t, err, expectedError.Error())
}

func Test_Start_SucceedsOnSuccessfulScopeEmittersInit(t *testing.T) {
	fmm := feature.CreateFeatureManagerMock(nil, nil, true)
	testingScraper := scraper.RequireComponentType(t, "test_scraper")

	sut := &manager{
		scraperRuntime: &Runtime{
			ScopeEmitters: map[string]scope.Emitter{
				"scope1": scope.CreateEmitterMock(nil, nil, "scope1"),
				"scope2": scope.CreateEmitterMock(nil, nil, "scope2"),
			},
		},
		scraperType:    testingScraper,
		featureManager: fmm,
		config: &ManagerConfig{
			DelayedProcessingConfig: &types.DelayedProcessingConfig{
				CollectionInterval: 10,
			},
		},
	}
	err := sut.Start(context.Background(), nil)

	assert.NoError(t, err, "nothing fails start must succeed")
}
