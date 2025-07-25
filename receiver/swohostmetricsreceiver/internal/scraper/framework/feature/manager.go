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

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/feature/features"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
)

// ManagerConfig is definition of configuration for feature manager.
type ManagerConfig struct {
	ScraperType component.Type
	*types.DelayedProcessingConfig
	// another feature related configs follows
}

type Manager interface {
	// Init initializes whole feature manager based
	// on given config. When features are identified
	// in config, they are initialized for use.
	Init(*ManagerConfig) error
	// Feature of delayed processing.
	features.DelayedProcessing
}

type Flag int

// Feature flags for feature manager.
const (
	DelayedProcessingFeature Flag = iota
)

type void = struct{}

type manager struct {
	// activateFeatures is a map of feature flags, which
	// are recognized by config.
	activatedFeatures map[Flag]*void
	// delayedProcessing is core implementation of feature
	// representing delayed processing.
	delayedProcessing features.DelayedProcessing

	logger *zap.Logger
}

var _ Manager = (*manager)(nil)

// Creates product oriented feature manager.
func NewFeatureManager(logger *zap.Logger) Manager {
	return createFeatureManager(
		features.NewDelayedProcessing(),
		logger,
	)
}

func createFeatureManager(
	delayedProcessing features.DelayedProcessing,
	logger *zap.Logger,
) Manager {
	return &manager{
		activatedFeatures: make(map[Flag]*void),
		delayedProcessing: delayedProcessing,
		logger:            logger,
	}
}

func (fm *manager) isActivated(ff Flag) bool {
	if _, found := fm.activatedFeatures[ff]; found {
		return true
	}
	return false
}

// Init implements FeatureManager.
func (fm *manager) Init(c *ManagerConfig) error {
	// Attempt to activate delayed processing.
	fm.tryToActivateDelayedProcessing(c)

	return nil
}

func (fm *manager) tryToActivateDelayedProcessing(c *ManagerConfig) {
	// Delay processing is part of config.
	if c.DelayedProcessingConfig != nil {
		fm.logger.Sugar().Debugf(
			"activating delayed processing for scraper '%s'",
			c.ScraperType)
		fm.activateFeature(DelayedProcessingFeature)
	}
}

func (fm *manager) activateFeature(ff Flag) {
	fm.activatedFeatures[ff] = new(void)
}

// This set of method should be moved into seprated file.

// InitDelayedProcessing implements FeatureManager.
func (fm *manager) InitDelayedProcessing(
	c *types.DelayedProcessingConfig,
	it time.Time,
) error {
	if !fm.isActivated(DelayedProcessingFeature) {
		return nil
	}
	return fm.delayedProcessing.InitDelayedProcessing(c, it)
}

// IsReady implements FeatureManager.
func (fm *manager) IsReady(consideredTime time.Time) bool {
	if !fm.isActivated(DelayedProcessingFeature) {
		return true
	}
	return fm.delayedProcessing.IsReady(consideredTime)
}

// UpdateLastProcessedTime implements FeatureManager.
func (fm *manager) UpdateLastProcessedTime(time time.Time) {
	if !fm.isActivated(DelayedProcessingFeature) {
		return
	}
	fm.delayedProcessing.UpdateLastProcessedTime(time)
}
