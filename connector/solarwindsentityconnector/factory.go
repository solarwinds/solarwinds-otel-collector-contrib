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

func createMetricsToLogsConnector(ctx context.Context, settings connector.Settings, config component.Config, logs consumer.Logs) (connector.Metrics, error) {
	cfg := config.(*Config)
	// TODO: think of better way to create the struct, when condition handling is introduced
	return &solarwindsentity{
		logger:            settings.Logger,
		sourcePrefix:      cfg.SourcePrefix,
		destinationPrefix: cfg.DestinationPrefix,
		entities:          cfg.Schema.NewEntities(),
		relationships:     cfg.Schema.NewRelationships(),
		logsConsumer:      logs,
	}, nil
}

func createLogsToLogsConnector(ctx context.Context, settings connector.Settings, config component.Config, logs consumer.Logs) (connector.Logs, error) {
	cfg := config.(*Config)
	// TODO: think of better way to create the struct, when condition handling is introduced
	return &solarwindsentity{
		logger:            settings.Logger,
		sourcePrefix:      cfg.SourcePrefix,
		destinationPrefix: cfg.DestinationPrefix,
		entities:          cfg.Schema.NewEntities(),
		relationships:     cfg.Schema.NewRelationships(),
		logsConsumer:      logs,
	}, nil
}
