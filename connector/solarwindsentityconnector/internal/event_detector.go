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
	"fmt"
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

func (e *EventDetector) collectEvents(attrs Attributes, entityEvents []*config.EntityEvent, relationshipEvents []*config.RelationshipEvent) ([]Event, error) {
	detectedRelationshipEvents, detectedRelationshipEntities := e.getRelationships(attrs, relationshipEvents)
	detectedValidRelationshipEntities := e.validateEntityEvents(entityEvents, detectedRelationshipEntities)
	detectedEntityEvents := e.getEntities(attrs, entityEvents, detectedValidRelationshipEntities)

	allEvents := append(detectedEntityEvents)
	for _, entity := range detectedValidRelationshipEntities {
		allEvents = append(allEvents, entity)
	}
	allEvents = append(allEvents, detectedRelationshipEvents...)
	return allEvents, nil
}

func (e *EventDetector) validateEntityEvents(validEvents []*config.EntityEvent, actualEvents map[string]Entity) map[string]Entity {
	events := make(map[string]Entity)
	for entityHash, actualEvent := range actualEvents {
		for _, event := range validEvents {
			if event.Type == actualEvent.Type {
				// If the event type matches, we know that we can infer the entity
				// because conditions were met.
				actualEvent.Action = event.Action
				events[entityHash] = actualEvent
			}
		}
	}
	return events
}

func (e *EventDetector) getEntities(attrs Attributes, entityEvents []*config.EntityEvent, relationshipEntities map[string]Entity) []Event {
	detectedEntityEvents := make([]Event, 0, len(entityEvents))
	for _, entityEvent := range entityEvents {
		event, err := e.attributeMapper.getEntity(entityEvent.Type, attrs)
		if err != nil {
			e.logger.Debug("failed to create entity update event", zap.Error(err))
			continue
		}

		eventHash, err := buildEntityHash(*event)
		_, alreadyExists := relationshipEntities[eventHash]
		if !alreadyExists {
			event.Action = entityEvent.Action
			detectedEntityEvents = append(detectedEntityEvents, *event)
		}
	}
	return detectedEntityEvents
}

func (e *EventDetector) getRelationships(attrs Attributes, relationshipEvents []*config.RelationshipEvent) ([]Event, map[string]Entity) {
	detectedRelationshipEvents := make([]Event, 0, len(relationshipEvents))
	detectedRelationshipEntities := make(map[string]Entity)
	for _, relationshipEvent := range relationshipEvents {
		sourceEntity, destEntity, err := e.attributeMapper.getRelationshipEntities(relationshipEvent.Source, relationshipEvent.Destination, attrs)
		if err != nil {
			e.logger.Debug("failed to create relationship entities", zap.Error(err))
			continue
		}

		srcHash, err := buildEntityHash(*sourceEntity)
		detectedRelationshipEntities[srcHash] = *sourceEntity

		dstHash, err := buildEntityHash(*destEntity)
		detectedRelationshipEntities[dstHash] = *destEntity

		relationship, err := createRelationship(relationshipEvent, sourceEntity, destEntity, attrs)
		if err != nil {
			e.logger.Debug("failed to create relationship event", zap.Error(err))
			continue
		}
		detectedRelationshipEvents = append(detectedRelationshipEvents, relationship)
	}
	return detectedRelationshipEvents, detectedRelationshipEntities
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

func createRelationship(relationship *config.RelationshipEvent, source, dest *Entity, attrs Attributes) (*Relationship, error) {
	action, err := GetActionString(relationship.Action)
	if err != nil {
		return nil, fmt.Errorf("failed to get action type for relationship event")
	}
	r := Relationship{
		Type: relationship.Type,
		Source: RelationshipEntity{
			Type: source.Type,
			IDs:  source.IDs,
		},
		Destination: RelationshipEntity{
			Type: dest.Type,
			IDs:  dest.IDs,
		},
		Action: action,
	}

	r.Attributes = getOptionalAttributes(relationship.Attributes, attrs.Common)

	return &r, nil
}

// BuildEntityKey constructs a unique key for the entity referenced in the relationship.
// The key is composition of entity type and its ID attributes.
func buildEntityHash(entity Entity) (string, error) {
	// Marshal directly to bytes instead of using a buffer and encoder
	data := struct {
		Type string
		IDs  map[string]any
	}{
		Type: entity.Type,
		IDs:  entity.IDs.AsRaw(),
	}
	return HashObject(data)
}
