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

package internal

import (
	"context"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.uber.org/zap"
)

type EventDetector struct {
	attributeMapper AttributeMapper
	logEvents       config.EventsGroup[ottllog.TransformContext]
	metricEvents    config.EventsGroup[ottlmetric.TransformContext]
	logger          *zap.Logger
}

func NewEventDetector(
	attributeMapper AttributeMapper,
	logEvents config.EventsGroup[ottllog.TransformContext],
	metricEvents config.EventsGroup[ottlmetric.TransformContext],
	logger *zap.Logger,
) *EventDetector {
	return &EventDetector{
		attributeMapper: attributeMapper,
		logEvents:       logEvents,
		metricEvents:    metricEvents,
		logger:          logger,
	}
}

func (e *EventDetector) DetectLog(ctx context.Context, resourceAttrs Attributes, transformCtx ottllog.TransformContext) ([]Event, error) {
	ee, re, err := processEvents(ctx, e.logEvents, transformCtx)
	if err != nil {
		return nil, err
	}
	return e.collectEvents(resourceAttrs, ee, re)
}

func (e *EventDetector) DetectMetric(ctx context.Context, resourceAttrs Attributes, transformCtx ottlmetric.TransformContext) ([]Event, error) {
	ee, re, err := processEvents(ctx, e.metricEvents, transformCtx)
	if err != nil {
		return nil, err
	}

	return e.collectEvents(resourceAttrs, ee, re)

}

func (e *EventDetector) collectEvents(attrs Attributes, ee []*config.EntityEvent, re []*config.RelationshipEvent) ([]Event, error) {
	events := make([]Event, 0, len(ee)+len(re))
	for _, entityEvent := range ee {
		newEvents, err := e.attributeMapper.getEntities(entityEvent.Type, attrs)
		if err != nil {
			e.logger.Debug("failed to create entity update event", zap.Error(err))
			continue
		}

		action, err := GetActionString(entityEvent.Action)
		for _, newEvent := range newEvents {
			newEvent.Action = action
			events = append(events, newEvent)
		}
	}

	for _, relationshipEvent := range re {
		newRel, err := e.attributeMapper.getRelationship(relationshipEvent, attrs)
		if err != nil {
			e.logger.Debug("failed to create relationship update event", zap.Error(err))
			continue
		}

		action, err := GetActionString(relationshipEvent.Action)
		if err != nil {
			e.logger.Debug("failed to get action type for relationship event", zap.Error(err))
			continue
		}
		newRel.Action = action
		events = append(events, newRel)
	}

	return events, nil
}

// ProcessEvents evaluates the conditions for entityConfigs and relationships events.
// If the conditions are met, it appends the corresponding entity or relationship update event to the event builder.
// Multiple condition items are evaluated using OR logic.
func processEvents[C any](
	ctx context.Context,
	events config.EventsGroup[C],
	tc C) ([]*config.EntityEvent, []*config.RelationshipEvent, error) {

	// will be reworked to channel
	entityEvents := make([]*config.EntityEvent, 0)
	relationshipEvents := make([]*config.RelationshipEvent, 0)

	for _, entityEvent := range events.Entities {
		ok, err := entityEvent.ConditionSeq.Eval(ctx, tc)
		if err != nil {
			return []*config.EntityEvent{}, []*config.RelationshipEvent{}, err
		}

		if ok {
			entityEvents = append(entityEvents, entityEvent.Definition)
		}
	}

	for _, relationshipEvent := range events.Relationships {
		ok, err := relationshipEvent.ConditionSeq.Eval(ctx, tc)
		if err != nil {
			return []*config.EntityEvent{}, []*config.RelationshipEvent{}, err
		}

		if ok {
			relationshipEvents = append(relationshipEvents, relationshipEvent.Definition)
		}
	}
	return entityEvents, relationshipEvents, nil
}
