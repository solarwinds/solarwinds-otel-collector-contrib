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
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-releases/pkg/testutil"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/types"
)

type DelayedProcessingMock struct {
	InitCC        testutil.CallsCounter
	IsReadyCC     testutil.CallsCounter
	UpdateCC      testutil.CallsCounter
	initResult    error
	isReadyResult bool
}

var _ DelayedProcessing = (*DelayedProcessingMock)(nil)

func CreateDelayedProcessingMock(
	initResult error,
	isReadyResult bool,
) *DelayedProcessingMock {
	return &DelayedProcessingMock{
		InitCC:        testutil.CallsCounter{},
		IsReadyCC:     testutil.CallsCounter{},
		UpdateCC:      testutil.CallsCounter{},
		initResult:    initResult,
		isReadyResult: isReadyResult,
	}
}

// InitDelayedProcessing implements DelayedProcessing.
func (d *DelayedProcessingMock) InitDelayedProcessing(
	*types.DelayedProcessingConfig,
	time.Time,
) error {
	d.InitCC.IncrementCount()
	return d.initResult
}

// IsReady implements DelayedProcessing.
func (d *DelayedProcessingMock) IsReady(time.Time) bool {
	d.IsReadyCC.IncrementCount()
	return d.isReadyResult
}

// UpdateLastProcessedTime implements DelayedProcessing.
func (d *DelayedProcessingMock) UpdateLastProcessedTime(time.Time) {
	d.UpdateCC.IncrementCount()
}
