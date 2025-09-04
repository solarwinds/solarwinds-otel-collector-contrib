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

package solarwindsprocessor

import (
	"context"
	"fmt"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/attributesdecorator"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/container"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/solarwindsprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
)

type solarwindsprocessor struct {
	logger            *zap.Logger
	cfg               *Config
	hostAttributes    *internal.HostAttributesDecorator
	extensionProvider ExtensionProvider
}

func (p *solarwindsprocessor) start(ctx context.Context, host component.Host) error {
	if p.extensionProvider == nil {
		return nil
	}

	extensionName := p.cfg.GetExtensionName()
	_, err := p.extensionProvider.Init(p.logger, extensionName, host)
	if err != nil {
		return fmt.Errorf("failed to get Solarwinds extension %q: %w", extensionName, err)
	}

	err = p.extensionProvider.SetAttributes(&p.cfg.ResourceAttributes)
	if err != nil {
		return fmt.Errorf("failed to decor resource attributes by extension properties: %w", err)
	}

	return nil
}

func notifySignalSizeLimitExceeded(signal any, limit int, logger *zap.Logger) error {
	if limit <= 0 {
		return nil // No limit set, skip the check.
	}

	var err error
	var bs []byte
	var signalName string

	switch v := signal.(type) {
	case plog.Logs:
		signalName = "Logs"
		er := plogotlp.NewExportRequestFromLogs(v)
		bs, err = er.MarshalProto()
	case pmetric.Metrics:
		signalName = "Metrics"
		er := pmetricotlp.NewExportRequestFromMetrics(v)
		bs, err = er.MarshalProto()
	case ptrace.Traces:
		signalName = "Traces"
		er := ptraceotlp.NewExportRequestFromTraces(v)
		bs, err = er.MarshalProto()
	default:
		return fmt.Errorf("unsupported signal type: %T", v)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal signal: %w", err)
	}

	bsLen := len(bs)
	limitBytes := limit * 1024 * 1024 // Convert MiB to bytes
	if bsLen > limitBytes {
		msg := fmt.Sprintf(
			"%s size %d bytes exceeds the limit of %d MiB",
			signalName, bsLen, limit,
		)
		logger.Warn(msg, zap.Int("size_bytes", bsLen), zap.Int("limit_mib", limit))
	}

	return nil
}

func (p *solarwindsprocessor) processLogs(
	ctx context.Context,
	logs plog.Logs,
) (plog.Logs, error) {
	attributesdecorator.DecorateResourceAttributesByFunction(logs.ResourceLogs(), p.hostAttributes.ApplyAttributes)
	attributesdecorator.DecorateResourceAttributes(logs.ResourceLogs(), p.cfg.ResourceAttributes)

	err := notifySignalSizeLimitExceeded(logs, p.cfg.MaxSizeMib, p.logger)
	if err != nil {
		msg := "failed to notify logs size limit exceeded"
		p.logger.Error(msg, zap.Error(err))
		return plog.Logs{}, fmt.Errorf("%s: %w", msg, err)
	}
	return logs, nil
}

func (p *solarwindsprocessor) processMetrics(
	ctx context.Context,
	metrics pmetric.Metrics,
) (pmetric.Metrics, error) {
	attributesdecorator.DecorateResourceAttributesByFunction(metrics.ResourceMetrics(), p.hostAttributes.ApplyAttributes)
	attributesdecorator.DecorateResourceAttributes(metrics.ResourceMetrics(), p.cfg.ResourceAttributes)

	err := notifySignalSizeLimitExceeded(metrics, p.cfg.MaxSizeMib, p.logger)
	if err != nil {
		msg := "failed to notify metrics size limit exceeded"
		p.logger.Error(msg, zap.Error(err))
		return pmetric.Metrics{}, fmt.Errorf("%s: %w", msg, err)
	}
	return metrics, nil
}

func (p *solarwindsprocessor) processTraces(
	ctx context.Context,
	traces ptrace.Traces,
) (ptrace.Traces, error) {
	attributesdecorator.DecorateResourceAttributesByFunction(traces.ResourceSpans(), p.hostAttributes.ApplyAttributes)
	attributesdecorator.DecorateResourceAttributes(traces.ResourceSpans(), p.cfg.ResourceAttributes)

	err := notifySignalSizeLimitExceeded(traces, p.cfg.MaxSizeMib, p.logger)
	if err != nil {
		msg := "failed to notify traces size limit exceeded"
		p.logger.Error(msg, zap.Error(err))
		return ptrace.Traces{}, fmt.Errorf("%s: %w", msg, err)
	}
	return traces, nil
}

func createProcessor(logger *zap.Logger, cfg component.Config) (*solarwindsprocessor, error) {
	c, err := checkConfig(cfg)
	if err != nil {
		return nil, err
	}
	return newProcessor(logger, c), nil
}

func checkConfig(cfg component.Config) (*Config, error) {
	c, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid configuration type: %T, expected *solarwindsprocessor.Config", cfg)
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	return c, nil
}

func newProcessor(logger *zap.Logger, cfg *Config) *solarwindsprocessor {
	var hostAttributes *internal.HostAttributesDecorator
	if cfg.HostAttributesDecoration.Enabled {
		hostAttributes = internal.NewHostAttributes(cfg.HostAttributesDecoration, container.NewProvider(logger), logger)
	}
	var extensionProvider ExtensionProvider
	if cfg.CollectorAttributesDecoration.Enabled {
		extensionProvider = NewExtensionProvider()
	}

	return &solarwindsprocessor{
		logger:            logger,
		cfg:               cfg,
		hostAttributes:    hostAttributes,
		extensionProvider: extensionProvider,
	}
}
