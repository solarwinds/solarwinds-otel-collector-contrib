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

package swootelentityrefprocessor // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swootelentityrefprocessor"

import (
	"context"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swootelentityrefprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/processor/xprocessor"
	"go.uber.org/zap"
)

// NewFactory creates the processor factory for swootelentityref.
func NewFactory() processor.Factory {
	return xprocessor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		xprocessor.WithLogs(createLogsProcessor, metadata.LogsStability),
		xprocessor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
		xprocessor.WithTraces(createTracesProcessor, metadata.TracesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{Action: ActionInsert}
}

func buildProcessor(cfg component.Config, logger *zap.Logger) (*swootelentityrefprocessor, error) {
	processorCfg := cfg.(*Config)
	if err := processorCfg.Validate(); err != nil {
		return nil, err
	}
	return &swootelentityrefprocessor{cfg: processorCfg, logger: logger}, nil
}

func createLogsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	p, err := buildProcessor(cfg, params.Logger)
	if err != nil {
		return nil, err
	}
	return processorhelper.NewLogs(
		ctx,
		params,
		cfg,
		nextConsumer,
		p.processLogs,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
		processorhelper.WithStart(p.Start),
	)
}

func createMetricsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	p, err := buildProcessor(cfg, params.Logger)
	if err != nil {
		return nil, err
	}
	return processorhelper.NewMetrics(
		ctx,
		params,
		cfg,
		nextConsumer,
		p.processMetrics,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
	)
}

func createTracesProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	p, err := buildProcessor(cfg, params.Logger)
	if err != nil {
		return nil, err
	}
	return processorhelper.NewTraces(
		ctx,
		params,
		cfg,
		nextConsumer,
		p.processTraces,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
	)
}
