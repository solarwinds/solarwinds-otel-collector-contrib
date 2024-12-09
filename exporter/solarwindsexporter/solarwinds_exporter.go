// Copyright 2024 SolarWinds Worldwide, LLC. All rights reserved.
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
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type exporterType int

const (
	metricsExporterType exporterType = iota
	logsExporterType
	tracesExporterType
)

type solarwindsExporter struct {
	exporterType
	config   *Config
	settings component.TelemetrySettings
	metrics  exporter.Metrics
	logs     exporter.Logs
	traces   exporter.Traces
}

func newExporter(
	ctx context.Context,
	cfg *Config,
	settings exporter.Settings,
	typ exporterType,
) *solarwindsExporter {

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	swiExporter := &solarwindsExporter{
		config:   cfg,
		settings: settings.TelemetrySettings,
	}
	if err := swiExporter.initExporterType(ctx, settings, typ); err != nil {
		panic(err)
	}

	return swiExporter
}

func (e *solarwindsExporter) initExporterType(
	ctx context.Context,
	settings exporter.Settings,
	typ exporterType,
) error {
	e.exporterType = typ
	otlpExporter := otlpexporter.NewFactory()
	otlpCfg, err := e.config.OTLPConfig()
	if err != nil {
		return err
	}

	switch typ {
	case metricsExporterType:
		e.metrics, err = otlpExporter.CreateMetrics(ctx, settings, otlpCfg)
		return err
	case logsExporterType:
		e.logs, err = otlpExporter.CreateLogs(ctx, settings, otlpCfg)
		return err
	case tracesExporterType:
		e.traces, err = otlpExporter.CreateTraces(ctx, settings, otlpCfg)
		return err
	default:
		return fmt.Errorf("unknown exporter type: %v", typ)
	}

}

func (e *solarwindsExporter) start(ctx context.Context, host component.Host) error {
	switch e.exporterType {
	case metricsExporterType:
		return e.metrics.Start(ctx, host)
	case logsExporterType:
		return e.logs.Start(ctx, host)
	case tracesExporterType:
		return e.traces.Start(ctx, host)
	default:
		return fmt.Errorf("unknown exporter type: %v", e.exporterType)
	}
}

func (e *solarwindsExporter) shutdown(ctx context.Context) error {
	switch e.exporterType {
	case metricsExporterType:
		return e.metrics.Shutdown(ctx)
	case logsExporterType:
		return e.logs.Shutdown(ctx)
	case tracesExporterType:
		return e.traces.Shutdown(ctx)
	default:
		return fmt.Errorf("unknown exporter type: %v", e.exporterType)
	}
}

func (e *solarwindsExporter) pushMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	if metrics.MetricCount() == 0 {
		return nil
	}

	return e.metrics.ConsumeMetrics(ctx, metrics)
}

func (e *solarwindsExporter) pushLogs(ctx context.Context, logs plog.Logs) error {
	if logs.LogRecordCount() == 0 {
		return nil
	}

	return e.logs.ConsumeLogs(ctx, logs)
}

func (e *solarwindsExporter) pushTraces(ctx context.Context, traces ptrace.Traces) error {
	if traces.SpanCount() == 0 {
		return nil
	}

	return e.traces.ConsumeTraces(ctx, traces)
}
