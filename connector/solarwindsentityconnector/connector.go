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
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/storage"
	"sort"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type solarwindsentity struct {
	logger *zap.Logger

	logsConsumer      consumer.Logs
	entities          map[string]config.Entity
	relationships     []config.Relationship
	sourcePrefix      string
	destinationPrefix string

	expirationPolicy config.ExpirationPolicy

	component.ShutdownFunc
}

var _ connector.Metrics = (*solarwindsentity)(nil)
var _ connector.Logs = (*solarwindsentity)(nil)

func (s *solarwindsentity) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (s *solarwindsentity) Start(ctx context.Context, host component.Host) error {
	if s.expirationPolicy.Enabled {
		if err := s.expirationPolicy.IsValid(); err != nil {
			s.logger.Error("expiration policy is invalid", zap.Error(err))
			return err
		}
		storage.StartCache(s.expirationPolicy)
	} else {
		s.logger.Debug("expiration policy is disabled, no expiration logs will be generated")
	}

	return nil
}

// Will be removed when condition logic is added.
// getReverseSortKeys returns the entity types in reverse sorted order.
func getReverseSortKeys(m map[string]config.Entity) []string {
	entityTypes := make([]string, 0, len(m))
	for k := range m {
		entityTypes = append(entityTypes, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(entityTypes)))
	return entityTypes
}

func (s *solarwindsentity) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	eventLogs := plog.NewLogs()
	eventBuilder := internal.NewEventBuilder(s.entities, s.relationships, s.sourcePrefix, s.destinationPrefix, &eventLogs, s.logger)

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		resourceMetric := metrics.ResourceMetrics().At(i)
		resourceAttrs := resourceMetric.Resource().Attributes()

		// This will be replaced with actual logic when conditions are introduced
		entityTypes := getReverseSortKeys(s.entities) // Will be removed when condition logic is added. Prevents random entity type order
		for _, k := range entityTypes {
			entity := s.entities[k]
			eventBuilder.AppendEntityUpdateEvent(entity, resourceAttrs)
		}

		// This will be replaced with actual logic when conditions are introduced
		for _, relationship := range s.relationships {
			eventBuilder.AppendRelationshipUpdateEvent(relationship, resourceAttrs)
		}
	}

	if eventLogs.LogRecordCount() == 0 {
		return nil
	}

	return s.logsConsumer.ConsumeLogs(ctx, eventLogs)
}

func (s *solarwindsentity) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	eventLogs := plog.NewLogs()
	eventBuilder := internal.NewEventBuilder(s.entities, s.relationships, s.sourcePrefix, s.destinationPrefix, &eventLogs, s.logger)

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		resourceLog := logs.ResourceLogs().At(i)
		resourceAttrs := resourceLog.Resource().Attributes()

		// This will be replaced with actual logic when conditions are introduced
		entityTypes := getReverseSortKeys(s.entities) // Will be removed when condition logic is added. Prevents random entity type order
		for _, k := range entityTypes {
			entity := s.entities[k]
			eventBuilder.AppendEntityUpdateEvent(entity, resourceAttrs)
		}

		// This will be replaced with actual logic when conditions are introduced
		for _, relationship := range s.relationships {
			eventBuilder.AppendRelationshipUpdateEvent(relationship, resourceAttrs)
		}
	}

	if eventLogs.LogRecordCount() == 0 {
		return nil
	}

	return s.logsConsumer.ConsumeLogs(ctx, eventLogs)
}
