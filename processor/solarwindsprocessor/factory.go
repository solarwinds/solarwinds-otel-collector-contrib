package solarwindsprocessor

import (
	"context"
	"fmt"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/solarwindsprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, metadata.TracesStability),
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
		processor.WithLogs(createLogsProcessor, metadata.LogsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		MaxSizeMib:         0,
		ResourceAttributes: make(map[string]string),
	}
}

func createLogsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	next consumer.Logs,
) (processor.Logs, error) {
	p, err := createProcessor(set.Logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Solarwinds processor: %w", err)
	}
	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		next,
		p.processLogs,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
		processorhelper.WithStart(p.start),
	)
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	next consumer.Metrics,
) (processor.Metrics, error) {
	p, err := createProcessor(set.Logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Solarwinds processor: %w", err)
	}
	return processorhelper.NewMetrics(
		ctx,
		set,
		cfg,
		next,
		p.processMetrics,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
		processorhelper.WithStart(p.start),
	)
}

func createTracesProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	next consumer.Traces,
) (processor.Traces, error) {
	p, err := createProcessor(set.Logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Solarwinds processor: %w", err)
	}
	return processorhelper.NewTraces(
		ctx,
		set,
		cfg,
		next,
		p.processTraces,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
		processorhelper.WithStart(p.start),
	)
}
