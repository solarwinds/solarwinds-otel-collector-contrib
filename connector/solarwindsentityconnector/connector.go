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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type solarwindsentity struct {
	logger *zap.Logger

	logsConsumer        consumer.Logs
	entitiesDefinitions map[string]config.Entity
	sourcePrefix        string
	destinationPrefix   string
	telemetrySettings   component.TelemetrySettings
	events              map[string]*config.Events

	component.StartFunc
	component.ShutdownFunc
}

var _ connector.Metrics = (*solarwindsentity)(nil)
var _ connector.Logs = (*solarwindsentity)(nil)

const (
	constMetric = ottlmetric.ContextName
	constLog    = ottllog.ContextName
)

func (s *solarwindsentity) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (s *solarwindsentity) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	eventLogs := plog.NewLogs()
	metricsEvents := s.events[constMetric]
	metricRelationshipEvents := metricsEvents.Relationships
	metricEntityEvents := metricsEvents.Entities
	eventBuilder := internal.NewEventBuilder(s.entitiesDefinitions, metricRelationshipEvents, s.sourcePrefix, s.destinationPrefix, &eventLogs, s.logger)

	parser, err := ottlmetric.NewParser(nil, s.telemetrySettings)
	if err != nil {
		s.logger.Error("Failed to create metrics parser", zap.Error(err))
		return err
	}

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		resourceMetric := metrics.ResourceMetrics().At(i)
		resourceAttrs := resourceMetric.Resource().Attributes()

		for j := 0; j < resourceMetric.ScopeMetrics().Len(); j++ {
			scopeMetric := resourceMetric.ScopeMetrics().At(j)

			for k := 0; k < scopeMetric.Metrics().Len(); k++ {
				metric := scopeMetric.Metrics().At(k)
				tc := ottlmetric.NewTransformContext(metric, scopeMetric.Metrics(), scopeMetric.Scope(), resourceMetric.Resource(), scopeMetric, resourceMetric)

				err = processEvents(ctx, eventBuilder, metricEntityEvents, metricRelationshipEvents, resourceAttrs, s, &parser, tc)
				if err != nil {
					s.logger.Error("Failed to process metric condition", zap.Error(err))
					return err
				}
			}
		}
	}

	if eventLogs.LogRecordCount() == 0 {
		return nil
	}

	return s.logsConsumer.ConsumeLogs(ctx, eventLogs)
}

func (s *solarwindsentity) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	eventLogs := plog.NewLogs()
	logEvents := s.events[constLog]
	logRelationshipEvents := logEvents.Relationships
	logEntityEvents := logEvents.Entities
	eventBuilder := internal.NewEventBuilder(s.entitiesDefinitions, logRelationshipEvents, s.sourcePrefix, s.destinationPrefix, &eventLogs, s.logger)
	parser, err := ottllog.NewParser(nil, s.telemetrySettings)
	if err != nil {
		s.logger.Error("Failed to create logs parser", zap.Error(err))
		return err
	}

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		resourceLog := logs.ResourceLogs().At(i)
		resourceAttrs := resourceLog.Resource().Attributes()

		for j := 0; j < resourceLog.ScopeLogs().Len(); j++ {
			scopeLog := resourceLog.ScopeLogs().At(j)

			for k := 0; k < scopeLog.LogRecords().Len(); k++ {
				logRecord := scopeLog.LogRecords().At(k)
				tc := ottllog.NewTransformContext(logRecord, scopeLog.Scope(), resourceLog.Resource(), scopeLog, resourceLog)

				err = processEvents(ctx, eventBuilder, logEntityEvents, logRelationshipEvents, resourceAttrs, s, &parser, tc)
				if err != nil {
					s.logger.Error("Failed to process logs condition", zap.Error(err))
					return err
				}
			}
		}
	}

	if eventLogs.LogRecordCount() == 0 {
		return nil
	}

	return s.logsConsumer.ConsumeLogs(ctx, eventLogs)
}

// processCondition evaluates the conditions for entities and relationships.
// If the conditions are met, it appends the corresponding entity or relationship update event to the event builder.
func processEvents[C any](
	ctx context.Context,
	eventBuilder *internal.EventBuilder,
	contextEntityEvents []config.EntityEvent,
	contextRelationshipEvents []config.RelationshipEvent,
	resourceAttrs pcommon.Map,
	s *solarwindsentity,
	parser *ottl.Parser[C],
	tc C) error {

	for _, entityEvent := range contextEntityEvents {

		condition := entityEvent.Conditions
		seq, err := newConditionSeq(
			*parser,
			s.telemetrySettings,
			condition,
		)

		if err != nil {
			s.logger.Error("Failed to create condition sequence", zap.Error(err))
			return err
		}

		ok, err := seq.Eval(ctx, tc)
		if err != nil {
			s.logger.Error("Failed to evaluate condition", zap.Error(err))
			return err
		}

		if ok {
			entity := s.entitiesDefinitions[entityEvent.Type]
			eventBuilder.AppendEntityUpdateEvent(entity, resourceAttrs)
		}
	}

	for _, relationshipEvent := range contextRelationshipEvents {

		conditions := relationshipEvent.Conditions
		seq, err := newConditionSeq(
			*parser,
			s.telemetrySettings,
			conditions,
		)

		if err != nil {
			s.logger.Error("Failed to create condition sequence", zap.Error(err))
			return err
		}

		ok, err := seq.Eval(ctx, tc)
		if err != nil {
			s.logger.Error("Failed to evaluate condition", zap.Error(err))
			return err
		}

		if ok {
			eventBuilder.AppendRelationshipUpdateEvent(relationshipEvent, resourceAttrs)
		}
	}
	return nil
}

// newConditionSeq creates a new condition sequence from the provided conditions.
// If no conditions are provided, it defaults to a sequence that always evaluates to true.
func newConditionSeq[C any](
	parser ottl.Parser[C],
	settings component.TelemetrySettings,
	conditions []string,
) (*ottl.ConditionSequence[C], error) {
	if len(conditions) == 0 {
		conditions = []string{"true"}
	}

	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return nil, err
	}

	seq := ottl.NewConditionSequence(stmts, settings)
	return &seq, nil
}
