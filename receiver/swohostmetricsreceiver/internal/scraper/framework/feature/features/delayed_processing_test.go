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

package features

import (
	"testing"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"github.com/stretchr/testify/assert"
)

func Test_InitDelayedProcessing_initialStateIsValid(t *testing.T) {
	initialTime := time.Now()
	collectionInterval := 30 * time.Second
	config := &types.DelayedProcessingConfig{
		CollectionInterval: collectionInterval,
	}

	sut := &delayedProcessing{}
	err := sut.InitDelayedProcessing(config, initialTime)

	assert.NoError(t, err, "init must not fail")
	assert.Equal(t, collectionInterval, sut.delayedInterval, "interval must fit after init")
	assert.Equal(t, initialTime, sut.lastProcessing, "last processing must fit after init")
}

func Test_IsReady_isPossitiveAfterInterval(t *testing.T) {
	initialTime := time.Now()
	collectionInterval := 30 * time.Second
	config := &types.DelayedProcessingConfig{
		CollectionInterval: collectionInterval,
	}
	afterInterval := initialTime.Add((30 * 2) * time.Second)

	sut := NewDelayedProcessing()
	err := sut.InitDelayedProcessing(config, initialTime)

	assert.NoError(t, err, "init must not fail")
	assert.True(t, sut.IsReady(afterInterval), "must be ready after the interval")
}

func Test_IsReady_isNegativeBeforeInterval(t *testing.T) {
	initialTime := time.Now()
	collectionInterval := 30 * time.Second
	config := &types.DelayedProcessingConfig{
		CollectionInterval: collectionInterval,
	}
	afterInterval := initialTime.Add((30 / 2) * time.Second)

	sut := NewDelayedProcessing()
	err := sut.InitDelayedProcessing(config, initialTime)

	assert.NoError(t, err, "init must not fail")
	assert.False(t, sut.IsReady(afterInterval), "must be ready after the interval")
}

func Test_UpdateLastProcessedTime_lastProcessingIsValid(t *testing.T) {
	lastProcessing := time.Now()

	sut := &delayedProcessing{}
	sut.UpdateLastProcessedTime(lastProcessing)

	assert.Equal(t, lastProcessing, sut.lastProcessing, "after update time must fit")
}
