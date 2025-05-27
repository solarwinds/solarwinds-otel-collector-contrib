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
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type solarwindsentity struct {
	logger *zap.Logger

	logsConsumer      consumer.Logs
	entities          map[string]config.Entity
	relationships     []config.RelationshipEvent
	sourcePrefix      string
	destinationPrefix string
	entityEvents      []config.EntityEvent
	telemetrySettings component.TelemetrySettings

	component.StartFunc
	component.ShutdownFunc
}

var _ connector.Metrics = (*solarwindsentity)(nil)
var _ connector.Logs = (*solarwindsentity)(nil)

func (s *solarwindsentity) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func newLogConditionSeqLog(
	settings component.TelemetrySettings,
	conditions []string,
) (*ottl.ConditionSequence[ottllog.TransformContext], error) {

	if len(conditions) == 0 || (len(conditions) == 1 && strings.TrimSpace(conditions[0]) == "") {
		conditions = []string{"true"}
	}

	parser, err := ottllog.NewParser(nil, settings)
	if err != nil {
		return nil, err
	}

	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return nil, err
	}

	seq := ottllog.NewConditionSequence(stmts, settings)
	return &seq, nil
}

func newLogConditionSeqMetric(
	settings component.TelemetrySettings,
	conditions []string,
) (*ottl.ConditionSequence[ottlmetric.TransformContext], error) {

	parser, err := ottlmetric.NewParser(nil, settings)
	if err != nil {
		return nil, err
	}

	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return nil, err
	}

	seq := ottlmetric.NewConditionSequence(stmts, settings)
	return &seq, nil
}

func (s *solarwindsentity) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	eventLogs := plog.NewLogs()
	eventBuilder := internal.NewEventBuilder(s.entities, s.relationships, s.sourcePrefix, s.destinationPrefix, &eventLogs, s.logger)

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		resourceMetric := metrics.ResourceMetrics().At(i)
		resourceAttrs := resourceMetric.Resource().Attributes()

		for j := 0; j < resourceMetric.ScopeMetrics().Len(); j++ {
			scopeMetric := resourceMetric.ScopeMetrics().At(j)

			for k := 0; k < scopeMetric.Metrics().Len(); k++ {
				metric := scopeMetric.Metrics().At(k)

				tc := ottlmetric.NewTransformContext(metric, scopeMetric.Metrics(), scopeMetric.Scope(), resourceMetric.Resource(), scopeMetric, resourceMetric)
				for _, entityEvent := range s.entityEvents {
					if entityEvent.Context != "metric" {
						continue
					}

					condition := entityEvent.Conditions
					seq, err := newLogConditionSeqMetric(s.telemetrySettings, condition)
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
						entity := s.entities[entityEvent.Type]
						eventBuilder.AppendEntityUpdateEvent(entity, resourceAttrs)
					}
				}

				for _, relationship := range s.relationships {
					if relationship.Context != "metric" {
						continue
					}

					conditions := relationship.Conditions
					seq, err := newLogConditionSeqMetric(s.telemetrySettings, conditions)
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
						eventBuilder.AppendRelationshipUpdateEvent(relationship, resourceAttrs)
					}
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
	eventBuilder := internal.NewEventBuilder(s.entities, s.relationships, s.sourcePrefix, s.destinationPrefix, &eventLogs, s.logger)

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		resourceLog := logs.ResourceLogs().At(i)
		resourceAttrs := resourceLog.Resource().Attributes()

		for j := 0; j < resourceLog.ScopeLogs().Len(); j++ {
			scopeLog := resourceLog.ScopeLogs().At(j)

			for k := 0; k < scopeLog.LogRecords().Len(); k++ {
				logRecord := scopeLog.LogRecords().At(k)
				tc := ottllog.NewTransformContext(logRecord, scopeLog.Scope(), resourceLog.Resource(), scopeLog, resourceLog)

				for _, entityEvent := range s.entityEvents {

					if entityEvent.Context != "log" {
						continue
					}

					condition := entityEvent.Conditions
					fmt.Println("Entity Event Conditions:", condition)
					seq, err := newLogConditionSeqLog(s.telemetrySettings, condition)
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
						entity := s.entities[entityEvent.Type]
						eventBuilder.AppendEntityUpdateEvent(entity, resourceAttrs)
					}

				}

				for _, relationship := range s.relationships {

					if relationship.Context != "log" {
						continue
					}

					conditions := relationship.Conditions
					seq, err := newLogConditionSeqLog(s.telemetrySettings, conditions)
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
						eventBuilder.AppendRelationshipUpdateEvent(relationship, resourceAttrs)
					}

				}

			}
		}
	}

	if eventLogs.LogRecordCount() == 0 {
		return nil
	}

	return s.logsConsumer.ConsumeLogs(ctx, eventLogs)
}
