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
	"time"

	"github.com/solarwinds/solarwinds-otel-collector/pkg/testutil"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/types"
)

type ManagerMock struct {
	InitCC                    testutil.CallsCounter
	InitDelayedCC             testutil.CallsCounter
	IsReadyCC                 testutil.CallsCounter
	UpdateLastProcessedTimeCC testutil.CallsCounter
	initResult                error
	initDelayed               error
	isReadyResult             bool
}

var _ Manager = (*ManagerMock)(nil)

func CreateFeatureManagerMock(
	initResult error,
	initDelayed error,
	isReadyResult bool,
) *ManagerMock {
	return &ManagerMock{
		InitCC:                    testutil.CallsCounter{},
		InitDelayedCC:             testutil.CallsCounter{},
		IsReadyCC:                 testutil.CallsCounter{},
		UpdateLastProcessedTimeCC: testutil.CallsCounter{},
		initResult:                initResult,
		initDelayed:               initDelayed,
		isReadyResult:             isReadyResult,
	}
}

// Init implements Manager.
func (m *ManagerMock) Init(*ManagerConfig) error {
	m.InitCC.IncrementCount()
	return m.initResult
}

// InitDelayedProcessing implements Manager.
func (m *ManagerMock) InitDelayedProcessing(
	*types.DelayedProcessingConfig,
	time.Time,
) error {
	m.InitDelayedCC.IncrementCount()
	return m.initDelayed
}

// IsReady implements Manager.
func (m *ManagerMock) IsReady(time.Time) bool {
	m.IsReadyCC.IncrementCount()
	return m.isReadyResult
}

// UpdateLastProcessedTime implements Manager.
func (m *ManagerMock) UpdateLastProcessedTime(time.Time) {
	m.UpdateLastProcessedTimeCC.IncrementCount()
}
