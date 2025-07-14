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

// collectEvents processes the attributes and configured events to create a list of detected events.
// First, it gathers entity events and relationship events with their associated entities.
// Then, it validates the relationship entities against the configured entity events to ensure that only
// valid and unique entity events are included.
func (e *EventDetector) collectEvents(
	attrs Attributes,
	configuredEvents []*config.EntityEvent,
	configuredRelationships []*config.RelationshipEvent,
) ([]Event, error) {
	entityEvents := e.getEntities(attrs, configuredEvents)
	relationshipEvents, relationshipEntityEvents := e.getRelationships(attrs, configuredRelationships)
	validRelationshipEntities := e.validateEntityEvents(configuredEvents, entityEvents, relationshipEntityEvents)

	allEvents := append(relationshipEvents)
	for _, entity := range entityEvents {
		allEvents = append(allEvents, entity)
	}
	allEvents = append(allEvents, validRelationshipEntities...)
	return allEvents, nil
}

// validateEntityEvents checks if the entities created from relationship events are not duplicates
// of those created from entity events. It filters out entities that already exist in the configured entity events.
// It returns a slice of events that are compared to configured entity events to ensure that only valid events are returned.
func (e *EventDetector) validateEntityEvents(
	configuredEntityEvents []*config.EntityEvent,
	alreadyExistingEntities,
	relationshipEntities map[string]Entity,
) []Event {
	events := make([]Event, 0)
	for entityHash, actualEvent := range relationshipEntities {
		if _, exists := alreadyExistingEntities[entityHash]; exists {
			// If the entity already exists (created from entity event, not from relationship event), we can skip it.
			continue
		}

		// If entity was created from relationship, we need to check whether it can be sent in the resulting log.
		// We do not want to send entity event if it was not configured in the entity events.
		for _, event := range configuredEntityEvents {
			// If the event type matches, we know that we can infer the entity
			// because conditions were met.
			if event.Type == actualEvent.Type {
				actualEvent.Action = event.Action
				events = append(events, actualEvent)
			}
		}
	}
	return events
}

// getEntities creates entity update events based on the configured entity events.
func (e *EventDetector) getEntities(attrs Attributes, entityEvents []*config.EntityEvent) map[string]Entity {
	detectedEntityEvents := make(map[string]Entity)
	for _, entityEvent := range entityEvents {
		event, err := e.attributeMapper.getEntity(entityEvent.Type, attrs)
		if err != nil {
			e.logger.Debug("failed to create entity update event", zap.Error(err))
			continue
		}

		// Build hash for later comparison between entities created
		// from relationship and entity events.
		eventHash, err := buildEntityHash(*event)
		if err != nil {
			e.logger.Debug("failed to build entity hash", zap.Error(err))
			continue
		}
		event.Action = entityEvent.Action
		detectedEntityEvents[eventHash] = *event
	}
	return detectedEntityEvents
}

// getRelationships creates relationship events based on the configured relationships.
// It also builds a map of entities tied to that relationship - source and destination.
// The map is used to compare entities created from relationship events with entities created from entity events,
// so duplicated can be filtered out in case of unprefixed attributes.
func (e *EventDetector) getRelationships(attrs Attributes, configuredRelationships []*config.RelationshipEvent) ([]Event, map[string]Entity) {
	relationshipEvents := make([]Event, 0, len(configuredRelationships))
	relationshipEntityEvents := make(map[string]Entity)
	for _, relationshipEvent := range configuredRelationships {
		sourceEntity, destEntity, err := e.attributeMapper.getRelationshipEntities(relationshipEvent.Source, relationshipEvent.Destination, attrs)
		if err != nil {
			e.logger.Debug("failed to create relationship entities", zap.Error(err))
			continue
		}

		// Prepare entity hashes for later comparison with entities created from entity events.
		srcHash, err := buildEntityHash(*sourceEntity)
		relationshipEntityEvents[srcHash] = *sourceEntity

		dstHash, err := buildEntityHash(*destEntity)
		relationshipEntityEvents[dstHash] = *destEntity

		if err != nil {
			e.logger.Debug("failed to build entity hash when creating relationship event", zap.Error(err))
			continue
		}

		relationship, err := createRelationship(relationshipEvent, sourceEntity, destEntity, attrs)
		if err != nil {
			e.logger.Debug("failed to create relationship event", zap.Error(err))
			continue
		}
		relationshipEvents = append(relationshipEvents, relationship)
	}
	return relationshipEvents, relationshipEntityEvents
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
