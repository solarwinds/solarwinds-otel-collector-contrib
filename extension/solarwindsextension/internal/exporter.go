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

package internal

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type Exporter struct {
	logger   *zap.Logger
	exporter exporter.Metrics
}

func newExporter(ctx context.Context, set extension.Settings, cfg *Config) (*Exporter, error) {
	set.Logger.Debug("Creating Exporter")
	oCfg, err := cfg.OTLPConfig()
	if err != nil {
		return nil, err
	}
	expSet := toExporterSettings(set)

	exp := &Exporter{logger: set.Logger}
	exp.exporter, err = otlpexporter.NewFactory().CreateMetrics(ctx, expSet, oCfg)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

func (e *Exporter) start(ctx context.Context, host component.Host) error {
	e.logger.Debug("Starting exporter")
	return e.exporter.Start(ctx, host)
}

func (e *Exporter) shutdown(ctx context.Context) error {
	e.logger.Debug("Shutting down exporter")
	return e.exporter.Shutdown(ctx)
}

func toExporterSettings(set extension.Settings) exporter.Settings {
	return exporter.Settings{
		ID:                set.ID,
		TelemetrySettings: set.TelemetrySettings,
		BuildInfo:         set.BuildInfo,
	}
}

func (e *Exporter) push(ctx context.Context, md pmetric.Metrics) error {
	return e.exporter.ConsumeMetrics(ctx, md)
}
