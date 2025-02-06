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

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/types"
)

// DelayedProcessing prescribes how implementation of delayed processing will
// looks like.
type DelayedProcessing interface {
	// InitDelayedProcessing method initializes delay processing feature.
	// config represents delayed processing configuration, which will be bounded
	// to feature instance.
	// initialTime represents time of first processing. From this initial time next
	// following time will be calculated.
	InitDelayedProcessing(config *types.DelayedProcessingConfig, initialTime time.Time) error
	// IsReady method returns true, when remaining time has passed and processing
	// can be made. False is returned when time has not yet pass.
	// consideredTime is time against which check is made.
	IsReady(consideredTime time.Time) bool
	// UpdateLastProcessedTime method upates timestamp of last processin. New
	// timestamp is set by provided time.
	UpdateLastProcessedTime(time time.Time)
}

type delayedProcessing struct {
	lastProcessing  time.Time
	delayedInterval time.Duration
}

var _ DelayedProcessing = (*delayedProcessing)(nil)

func NewDelayedProcessing() DelayedProcessing {
	return new(delayedProcessing)
}

// Init implements DelayedProcessing.
func (d *delayedProcessing) InitDelayedProcessing(
	config *types.DelayedProcessingConfig,
	initialTime time.Time,
) error {
	d.delayedInterval = config.CollectionInterval
	d.lastProcessing = initialTime
	return nil
}

// IsReady implements DelayedProcessing.
func (d *delayedProcessing) IsReady(consideredTime time.Time) bool {
	permittedNextProcesing := d.lastProcessing.Add(d.delayedInterval)
	return !permittedNextProcesing.After(consideredTime)
}

// UpdateLastProcessedTime implements DelayedProcessing.
func (d *delayedProcessing) UpdateLastProcessedTime(time time.Time) {
	d.lastProcessing = time
}
