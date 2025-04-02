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

package solarwindsexporter

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/solarwinds/solarwinds-otel-collector-releases/exporter/solarwindsexporter/internal/metadata"
)

// NewFactory creates a factory for Solarwinds Exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		NewDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, metadata.MetricsStability),
		exporter.WithLogs(createLogsExporter, metadata.LogsStability),
		exporter.WithTraces(createTracesExporter, metadata.TracesStability))
}

func createMetricsExporter(
	ctx context.Context,
	settings exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	exporterCfg, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected config type: %T", cfg)
	}

	metricsExporter, err := newExporter(exporterCfg, settings, metricsExporterType)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	return exporterhelper.NewMetrics(
		ctx,
		settings,
		cfg,
		metricsExporter.pushMetrics,
		exporterhelper.WithTimeout(exporterCfg.Timeout),
		exporterhelper.WithRetry(exporterCfg.BackoffSettings),
		exporterhelper.WithQueue(exporterCfg.QueueSettings),
		exporterhelper.WithStart(metricsExporter.start),
		exporterhelper.WithShutdown(metricsExporter.shutdown),
	)
}

func createLogsExporter(
	ctx context.Context,
	settings exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	exporterCfg, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected config type: %T", cfg)
	}

	logsExporter, err := newExporter(exporterCfg, settings, logsExporterType)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewLogs(
		ctx,
		settings,
		cfg,
		logsExporter.pushLogs,
		exporterhelper.WithTimeout(exporterCfg.Timeout),
		exporterhelper.WithRetry(exporterCfg.BackoffSettings),
		exporterhelper.WithQueue(exporterCfg.QueueSettings),
		exporterhelper.WithStart(logsExporter.start),
		exporterhelper.WithShutdown(logsExporter.shutdown),
	)
}

func createTracesExporter(ctx context.Context,
	settings exporter.Settings,
	cfg component.Config) (exporter.Traces, error) {
	exporterCfg, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected config type: %T", cfg)
	}

	tracesExporter, err := newExporter(exporterCfg, settings, tracesExporterType)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTraces(
		ctx,
		settings,
		cfg,
		tracesExporter.pushTraces,
		exporterhelper.WithTimeout(exporterCfg.Timeout),
		exporterhelper.WithRetry(exporterCfg.BackoffSettings),
		exporterhelper.WithQueue(exporterCfg.QueueSettings),
		exporterhelper.WithStart(tracesExporter.start),
		exporterhelper.WithShutdown(tracesExporter.shutdown),
	)
}
