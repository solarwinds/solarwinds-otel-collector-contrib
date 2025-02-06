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

package types

import (
	"fmt"
	"time"
)

// DelayedProcessing prescribes behavior for component with delayed processing. OTEL platform is
// designed to have one collection interval per receiver. If we design scraper
// or metric emitter, which should have delayed processing against parent receiver
// this interface should be implemented.
type DelayedProcessing interface {
	// SetLastProcessing Sets up last processing timestamp.
	// lastProcessing is timestamp of last processing.
	SetLastProcessing(lastProcessing time.Time)

	// IsReadyForProcessing determines if it is time for next processing.
	// nextProcessingTime is time, which will be used for determination.
	IsReadyForProcessing(nextProcessing time.Time) bool

	// InitDelayedProcessing sets DelayedCollectionInterval.
	InitDelayedProcessing(interval time.Duration) error

	// ShouldDelayRun returns true in case the scrape should be delayed,
	// otherwise returns false if the run should be executed
	ShouldDelayRun() bool
}

// DelayedProcessingInternals contains fields for delayed processing concept, easing storing and handling
// variables for delayed processing.
type DelayedProcessingInternals struct {
	// Timestamp of last processing
	LastProcessing time.Time

	// Collection interval of delayed processing
	DelayedCollectionInterval time.Duration
}

var _ DelayedProcessing = (*DelayedProcessingInternals)(nil)

// ShouldDelayRun implements types.DelayedProcessing.
func (internals *DelayedProcessingInternals) ShouldDelayRun() bool {
	now := time.Now()
	if internals.IsReadyForProcessing(now) {
		// time for next scrape => update last scrape with now
		internals.SetLastProcessing(now)
		return false
	}

	return true
}

// InitDelayedProcessing implements types.DelayedProcessing.
func (internals *DelayedProcessingInternals) InitDelayedProcessing(interval time.Duration) error {
	if interval.Seconds() < 0 { // todo: cannot run with <1 in functional test, because it is set in factory
		return fmt.Errorf("invalid collection interval value (%s)", interval)
	}

	internals.DelayedCollectionInterval = interval
	internals.LastProcessing = time.Now()
	return nil
}

// IsReadyForProcessing implements types.DelayedProcessing.
func (internals *DelayedProcessingInternals) IsReadyForProcessing(nextProcessing time.Time) bool {
	permittedNextTime := internals.LastProcessing.Add(internals.DelayedCollectionInterval)
	return !permittedNextTime.After(nextProcessing)
}

// SetLastProcessing implements types.DelayedProcessing.
func (internals *DelayedProcessingInternals) SetLastProcessing(lastProcessing time.Time) {
	internals.LastProcessing = lastProcessing
}

// Configuration struct for delayed processing feature.
type DelayedProcessingConfig struct {
	// processing interval
	CollectionInterval time.Duration `mapstructure:"delayed_collection_interval"`
}
