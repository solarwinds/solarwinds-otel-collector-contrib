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

package example

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/receiver"
)

func Test_CreateScraperExplicitly_ScraperIsProvided(t *testing.T) {
	config := &ScraperConfig{
		DelayedProcessingConfig: types.DelayedProcessingConfig{},
		ScraperConfig: types.ScraperConfig{
			Metrics: map[string]types.MetricSettingsConfig{
				scope1metric1: {Enabled: true},
			},
		},
	}

	sut, err := CreateScraperExplicitly(
		context.TODO(),
		receiver.Settings{},
		config,
		zap.NewNop(),
	)

	assert.NoError(t, err, "creation must not fail")
	assert.NotNil(t, sut, "scraper must be provided")
}

func Test_CreateScraperExplicitly_FailsOnEmptyConfig(t *testing.T) {
	config := &ScraperConfig{
		ScraperConfig: types.ScraperConfig{},
	}

	sut, err := CreateScraperExplicitly(
		context.TODO(),
		receiver.Settings{},
		config,
		zap.NewNop(),
	)

	assert.Error(t, err, "creation must fail")
	assert.Nil(t, sut, "scraper must not be provided")
}

func Test_CreateScraperImplicitly_ScraperIsProvided(t *testing.T) {
	config := &ScraperConfig{
		DelayedProcessingConfig: types.DelayedProcessingConfig{},
		ScraperConfig: types.ScraperConfig{
			Metrics: map[string]types.MetricSettingsConfig{
				scope1metric1: {Enabled: true},
			},
		},
	}

	sut, err := CreateScraperImplicitly(
		context.TODO(),
		receiver.Settings{},
		config,
		zap.NewNop(),
	)

	assert.NoError(t, err, "creation must not fail")
	assert.NotNil(t, sut, "scraper must be provided")
}

func Test_CreateScraperImplicitly_FailsOnEmptyConfig(t *testing.T) {
	config := &ScraperConfig{
		ScraperConfig: types.ScraperConfig{},
	}

	sut, err := CreateScraperImplicitly(
		context.TODO(),
		receiver.Settings{},
		config,
		zap.NewNop(),
	)

	assert.Error(t, err, "creation must fail")
	assert.Nil(t, sut, "scraper must not be provided")
}
