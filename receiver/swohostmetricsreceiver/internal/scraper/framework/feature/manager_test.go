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

package feature

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/framework/feature/features"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/types"
)

const once = int32(1)

type CustomScraperConfig struct {
	types.ScraperConfig
	types.DelayedProcessingConfig
}

func Test_ConfigIngestioByFeatureManager(t *testing.T) {
	// Existing configuration for scraper.
	sc := &CustomScraperConfig{
		ScraperConfig: types.ScraperConfig{},
		DelayedProcessingConfig: types.DelayedProcessingConfig{
			CollectionInterval: 1000 * time.Second,
		},
	}

	artificialScraper := scraper.RequireComponentType(t, "ArtificialScraper")

	// Configuration enabling delayed processing.
	fmc := &ManagerConfig{
		ScraperType:             artificialScraper,
		DelayedProcessingConfig: &sc.DelayedProcessingConfig, // Not nil config.
	}

	sut := NewFeatureManager()
	err := sut.Init(fmc)

	assert.Nil(t, err)
}

func Test_Init_DelayedProcessingActivated(t *testing.T) {
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	features := make(map[Flag]*void)
	config := &ManagerConfig{
		ScraperType:             testingScraper,
		DelayedProcessingConfig: &types.DelayedProcessingConfig{},
	}

	sut := &manager{
		activatedFeatures: features,
	}
	err := sut.Init(config)

	assert.NoError(t, err, "init must not fail")
	_, found := features[DelayedProcessingFeature]
	assert.True(t, found, "delayed processing must be activated")
}

func Test_Init_DelayedProcessingNotActivated(t *testing.T) {
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	features := make(map[Flag]*void)
	config := &ManagerConfig{
		ScraperType: testingScraper,
	}

	sut := &manager{
		activatedFeatures: features,
	}
	err := sut.Init(config)

	assert.NoError(t, err, "init must not fail")
	_, found := features[DelayedProcessingFeature]
	assert.False(t, found, "delayed processing must not be activated")
}

func Test_InitDelayedProcessing_whenFeatureActivatedImplementationIsCalled(t *testing.T) {
	dpm := features.CreateDelayedProcessingMock(nil, true)
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	config := &ManagerConfig{
		ScraperType: testingScraper,
		DelayedProcessingConfig: &types.DelayedProcessingConfig{
			CollectionInterval: 10,
		},
	}

	sut := createFeatureManager(dpm)
	err := sut.Init(config)
	assert.NoError(t, err, "init must not fail")
	_ = sut.InitDelayedProcessing(nil, time.Now())

	assert.Equal(t, once, dpm.InitCC.GetCallsCount(), "implementation must be called through")
}

func Test_InitDelayedProcessing_whenFeatureNotActivatedImplementationIsNotCalled(t *testing.T) {
	dpm := features.CreateDelayedProcessingMock(nil, true)
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	config := &ManagerConfig{
		ScraperType: testingScraper,
	}

	sut := createFeatureManager(dpm)
	err := sut.Init(config)
	assert.NoError(t, err, "init must not fail")
	_ = sut.InitDelayedProcessing(nil, time.Now())

	assert.Zero(t, dpm.InitCC.GetCallsCount(), "implementation must not be called through")
}

func Test_IsReady_whenFeatureActivatedImplementationIsCalled(t *testing.T) {
	dpm := features.CreateDelayedProcessingMock(nil, true)
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	config := &ManagerConfig{
		ScraperType: testingScraper,
		DelayedProcessingConfig: &types.DelayedProcessingConfig{
			CollectionInterval: 10,
		},
	}

	sut := createFeatureManager(dpm)
	err := sut.Init(config)
	_ = sut.IsReady(time.Now())

	assert.NoError(t, err, "init must not fail")
	assert.Equal(t, once, dpm.IsReadyCC.GetCallsCount(), "implementation must be called through")
}

func Test_IsReady_whenFeatureNotActivatedImplementationIsNotCalled(t *testing.T) {
	dpm := features.CreateDelayedProcessingMock(nil, true)
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	config := &ManagerConfig{
		ScraperType: testingScraper,
	}

	sut := createFeatureManager(dpm)
	err := sut.Init(config)
	_ = sut.IsReady(time.Now())

	assert.NoError(t, err, "init must not fail")
	assert.Zero(t, dpm.IsReadyCC.GetCallsCount(), "implementation must not be called through")
}

func Test_UpdateLastProcessedTime_whenFeatureActivatedImplementationIsCalled(t *testing.T) {
	dpm := features.CreateDelayedProcessingMock(nil, true)
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	config := &ManagerConfig{
		ScraperType: testingScraper,
		DelayedProcessingConfig: &types.DelayedProcessingConfig{
			CollectionInterval: 10,
		},
	}

	sut := createFeatureManager(dpm)
	err := sut.Init(config)
	sut.UpdateLastProcessedTime(time.Now())

	assert.NoError(t, err, "init must not fail")
	assert.Equal(t, once, dpm.UpdateCC.GetCallsCount(), "implementation must be called through")
}

func Test_UpdateLastProcessedTime_whenFeatureNotActivatedImplementationIsNotCalled(t *testing.T) {
	dpm := features.CreateDelayedProcessingMock(nil, true)
	testingScraper := scraper.RequireComponentType(t, "testing_scraper")

	config := &ManagerConfig{
		ScraperType: testingScraper,
	}

	sut := createFeatureManager(dpm)
	err := sut.Init(config)
	sut.UpdateLastProcessedTime(time.Now())

	assert.NoError(t, err, "init must not fail")
	assert.Zero(t, dpm.UpdateCC.GetCallsCount(), "implementation must not be called through")
}
