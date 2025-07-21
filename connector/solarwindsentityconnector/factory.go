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

package solarwindsentityconnector

import (
	"context"
	"fmt"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
)

func NewFactory() connector.Factory {
	return connector.NewFactory(
		metadata.Type,
		NewDefaultConfig,
		connector.WithMetricsToLogs(createMetricsToLogsConnector, metadata.MetricsToLogsStability),
		connector.WithLogsToLogs(createLogsToLogsConnector, metadata.LogsToLogsStability),
	)
}

func createLogsToLogsConnector(ctx context.Context, settings connector.Settings, config component.Config, logs consumer.Logs) (connector.Logs, error) {
	return createConnector(settings, config, logs)
}

func createMetricsToLogsConnector(ctx context.Context, settings connector.Settings, config component.Config, logs consumer.Logs) (connector.Metrics, error) {
	return createConnector(settings, config, logs)
}

func createConnector(settings connector.Settings, config component.Config, logs consumer.Logs) (*solarwindsentity, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("expected config of type *Config, got %T", config)
	}
	events := cfg.Schema.NewEvents(settings.TelemetrySettings)
	attributeMapper := internal.NewAttributeMapper(cfg.Schema.NewEntities())

	se := &solarwindsentity{
		logger: settings.Logger,
		eventDetector: internal.NewEventDetector(
			attributeMapper,
			events.LogEvents,
			events.MetricEvents,
			settings.Logger,
		),
		sourcePrefix:     cfg.SourcePrefix,
		destPrefix:       cfg.DestinationPrefix,
		expirationPolicy: cfg.Expiration,
		logsConsumer:     logs,
	}
	return se, nil
}
