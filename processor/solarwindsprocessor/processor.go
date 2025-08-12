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

	"github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/attributesdecorator"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/container"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/extensionfinder"
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
	logger         *zap.Logger
	cfg            *Config
	hostAttributes internal.HostAttributes
}

func (p *solarwindsprocessor) start(ctx context.Context, host component.Host) error {
	swiExtension, err := extensionfinder.FindExtension[*solarwindsextension.SolarwindsExtension](
		p.logger,
		p.cfg.ExtensionName,
		host,
	)
	if err != nil {
		msg := fmt.Sprintf("failed to find Solarwinds Extension %q", p.cfg.ExtensionName)
		p.logger.Error(msg, zap.Error(err))
		return fmt.Errorf("%s: %w", msg, err)
	}

	// Adjust configuration by fetching values from the extension.
	p.adjustConfigurationByExtension(swiExtension)

	// Fetch additional properties for host detection if enabled.
	if p.cfg.HostAttributesEnabled == true {
		containerProvider := container.NewProvider()
		containerId, err := containerProvider.ReadContainerInstanceID()
		if err != nil {
			return fmt.Errorf("failed to read container instance ID: %w", err)
		}
		isInContainerd := containerProvider.IsRunInContainerd()

		p.hostAttributes = internal.HostAttributes{
			ContainerID:       containerId,
			IsRunInContainerd: isInContainerd,
			ClientId:          p.cfg.ClientId,
		}
	}

	return nil
}

func (p *solarwindsprocessor) adjustConfigurationByExtension(
	swiExtension *solarwindsextension.SolarwindsExtension,
) {
	extCfg := swiExtension.GetCommonConfig()

	// Without entity flag.
	if !extCfg.WithoutEntity() {
		p.cfg.ResourceAttributes[solarwindsextension.EntityCreation] = solarwindsextension.EntityCreationValue
	}

	// Collector name.
	p.cfg.ResourceAttributes[solarwindsextension.CollectorNameAttribute] = extCfg.CollectorName()
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
	attributesdecorator.DecorateResourceAttributes(logs.ResourceLogs(), p.cfg.ResourceAttributes)
	attributesdecorator.DecorateResourceAttributesByPluginIdentifiers(logs.ResourceLogs(), p.cfg.HostIdentification)

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
	attributesdecorator.DecorateResourceAttributes(metrics.ResourceMetrics(), p.cfg.ResourceAttributes)
	attributesdecorator.DecorateResourceAttributesByPluginIdentifiers(metrics.ResourceMetrics(), p.cfg.HostIdentification)

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
	attributesdecorator.DecorateResourceAttributes(traces.ResourceSpans(), p.cfg.ResourceAttributes)
	attributesdecorator.DecorateResourceAttributesByPluginIdentifiers(traces.ResourceSpans(), p.cfg.HostIdentification)

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
	return &solarwindsprocessor{
		logger: logger,
		cfg:    cfg,
	}
}
