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

package solarwindsextension

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector/extension/solarwindsextension/internal"
)

// CollectorNameAttribute is a resource attribute
// representing the configured name of the collector.
const CollectorNameAttribute = internal.CollectorNameAttribute

type SolarwindsExtension struct {
	logger    *zap.Logger
	config    *internal.Config
	heartbeat *internal.Heartbeat
}

func newExtension(ctx context.Context, set extension.Settings, cfg *internal.Config) (*SolarwindsExtension, error) {
	set.Logger.Info("Creating Solarwinds Extension")
	set.Logger.Info("Config", zap.Any("config", cfg))

	e := &SolarwindsExtension{
		logger: set.Logger,
		config: cfg,
	}
	var err error
	e.heartbeat, err = internal.NewHeartbeat(ctx, set, cfg)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *SolarwindsExtension) GetCommonConfig() CommonConfig { return newCommonConfig(e.config) }

func (e *SolarwindsExtension) Start(ctx context.Context, host component.Host) error {
	e.logger.Info("Starting Solarwinds Extension")
	return e.heartbeat.Start(ctx, host)
}

func (e *SolarwindsExtension) Shutdown(ctx context.Context) error {
	e.logger.Info("Shutting down Solarwinds Extension")
	// Everything must be shut down, regardless of the failure.
	return e.heartbeat.Shutdown(ctx)

}
