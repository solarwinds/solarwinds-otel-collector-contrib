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

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.uber.org/zap"
)

type EventDetector[T any] struct {
	attributeMapper AttributeMapper
	events          config.EventsGroup[T]
	keyBuilder      KeyBuilder
	logger          *zap.Logger
}

func NewEventDetector[T any](
	attributeMapper AttributeMapper,
	events config.EventsGroup[T],
	logger *zap.Logger,
) *EventDetector[T] {
	return &EventDetector[T]{
		attributeMapper: attributeMapper,
		events:          events,
		keyBuilder:      NewKeyBuilder(),
		logger:          logger,
	}
}

func (e *EventDetector[T]) Detect(ctx context.Context, resourceAttrs Attributes, transformCtx T) ([]Event, error) {
	ee, re, err := e.processEvents(ctx, transformCtx)
	if err != nil {
		return nil, err
	}
	return e.collectEvents(resourceAttrs, ee, re)
}

// collectEvents processes the attributes and configured events to create a list of detected events.
// First, it gathers entity events and relationship events with their associated entities.
// Then, it validates the relationship entities against the configured entity events to ensure that only
// valid and unique entity events are included.
func (e *EventDetector[T]) collectEvents(
	attrs Attributes,
	configuredEvents []*config.EntityEvent,
	configuredRelationships []*config.RelationshipEvent,
) ([]Event, error) {
	entityEvents := e.getEntityEvents(attrs, configuredEvents)
	relationshipEvents := e.getRelationshipEvents(attrs, configuredRelationships)
	validRelationshipEntities := e.validateEntityEvents(configuredEvents, entityEvents, relationshipEvents)

	allEvents := make([]Event, 0, len(relationshipEvents)+len(entityEvents)+len(validRelationshipEntities))

	// append all entity events detected from entity events
	for _, entity := range entityEvents {
		allEvents = append(allEvents, entity)
	}

	// append all relationship events
	for _, relationshipEvent := range relationshipEvents {
		allEvents = append(allEvents, relationshipEvent)
	}

	// append all valid relationship entities that were not duplicates of entity events
	// and are inferred from relationship events
	allEvents = append(allEvents, validRelationshipEntities...)

	return allEvents, nil
}

// validateEntityEvents checks if the entities created from relationship events are not duplicates
// of those created from entity events. It filters out entities that already exist in the configured entity events.
// It returns a slice of events that are compared to configured entity events to ensure that only valid events are returned.
func (e *EventDetector[T]) validateEntityEvents(
	validEntityEvents []*config.EntityEvent,
	alreadyExistingEntities map[string]Entity,
	relationshipEntities []*Relationship,
) []Event {
	events := make([]Event, 0)
	for _, actualEvent := range relationshipEntities {
		sourceEntity, err := e.validateRelationshipEntity(actualEvent.Source, validEntityEvents, alreadyExistingEntities)
		if err != nil {
			e.logger.Debug("failed to validate source entity for relationship event", zap.Error(err))
			continue
		}
		if sourceEntity != nil {
			events = append(events, *sourceEntity)
		}

		destEntity, err := e.validateRelationshipEntity(actualEvent.Destination, validEntityEvents, alreadyExistingEntities)
		if err != nil {
			e.logger.Debug("failed to validate destination entity for relationship event", zap.Error(err))
			continue
		}
		if destEntity != nil {
			events = append(events, *destEntity)
		}
	}
	return events
}

// validateRelationshipEntity checks if the entity created from a relationship event is valid.
// Valid means that the event:
//  1. is configured in the connector configuration
//  2. is not already existing in the alreadyExistingEntities map, which contains entities created from entity events
//  3. conditions for the entity event were met, so we can infer the entity from the relationship event; these events are stored in validEntityEvents
func (e *EventDetector[T]) validateRelationshipEntity(
	entity Entity,
	validEntityEvents []*config.EntityEvent,
	alreadyExistingEntities map[string]Entity,
) (*Entity, error) {
	entityHash, err := e.keyBuilder.BuildEntityKey(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to build entity hash for relationship event")
	}

	if _, exists := alreadyExistingEntities[entityHash]; exists {
		// If the entity already exists (created from entity event, not from relationship event), we can skip it.
		return nil, nil
	}

	// If entity was created from relationship, we need to check whether it can be sent in the resulting log.
	// We do not want to send entity event if it was not configured in the entity events.
	// validEntityEvents contains all entity events that were configured in the entity events and matched the conditions
	// for current processed resource.
	for _, event := range validEntityEvents {
		// If the event type matches, we know that we can infer the entity
		// because conditions were met.
		if event.Entity == entity.Type {
			entity.Action = event.Action
			return &entity, nil
		}
	}

	return nil, nil
}

// getEntityEvents creates entity update events based on the configured entity events.
func (e *EventDetector[T]) getEntityEvents(attrs Attributes, entityEvents []*config.EntityEvent) map[string]Entity {
	detectedEntityEvents := make(map[string]Entity)
	for _, entityEvent := range entityEvents {
		event, err := e.attributeMapper.getEntity(entityEvent.Entity, attrs)
		if err != nil {
			e.logger.Debug("failed to create entity update event", zap.Error(err))
			continue
		}

		// Build hash for later comparison between entities created
		// from relationship and entity events.
		eventHash, err := e.keyBuilder.BuildEntityKey(*event)
		if err != nil {
			e.logger.Debug("failed to build entity hash", zap.Error(err))
			continue
		}
		event.Action = entityEvent.Action
		detectedEntityEvents[eventHash] = *event
	}
	return detectedEntityEvents
}

// getRelationshipEvents creates relationship events based on the configured relationships.
// It also builds a map of entities tied to that relationship - source and destination.
// The map is used to compare entities created from relationship events with entities created from entity events,
// so duplicated can be filtered out in case of unprefixed attributes.
func (e *EventDetector[T]) getRelationshipEvents(attrs Attributes, configuredRelationships []*config.RelationshipEvent) []*Relationship {
	relationshipEvents := make([]*Relationship, 0, len(configuredRelationships))
	for _, relationshipEvent := range configuredRelationships {
		sourceEntity, destEntity, err := e.attributeMapper.getRelationshipEntities(relationshipEvent.Source, relationshipEvent.Destination, attrs)
		if err != nil {
			e.logger.Debug("failed to create relationship entities", zap.Error(err))
			continue
		}

		relationship, err := createRelationship(relationshipEvent, sourceEntity, destEntity, attrs)
		if err != nil {
			e.logger.Debug("failed to create relationship event", zap.Error(err))
			continue
		}
		relationshipEvents = append(relationshipEvents, &relationship)
	}
	return relationshipEvents
}

// ProcessEvents evaluates the conditions for entityConfigs and relationships events.
// If the conditions are met, it appends the corresponding entity or relationship update event to the event builder.
// Multiple condition items are evaluated using OR logic.
func (e *EventDetector[T]) processEvents(
	ctx context.Context,
	tc T) ([]*config.EntityEvent, []*config.RelationshipEvent, error) {

	// will be reworked to channel
	entityEvents := make([]*config.EntityEvent, 0)
	relationshipEvents := make([]*config.RelationshipEvent, 0)

	for _, entityEvent := range e.events.Entities {
		ok, err := entityEvent.ConditionSeq.Eval(ctx, tc)
		if err != nil {
			return []*config.EntityEvent{}, []*config.RelationshipEvent{}, err
		}

		if ok {
			entityEvents = append(entityEvents, entityEvent.Definition)
		}
	}

	for _, relationshipEvent := range e.events.Relationships {
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

func createRelationship(relationship *config.RelationshipEvent, source, dest *Entity, attrs Attributes) (Relationship, error) {
	action, err := GetActionString(relationship.Action)
	if err != nil {
		return Relationship{}, fmt.Errorf("failed to get action type for relationship event")
	}
	r := Relationship{
		Type:        relationship.Type,
		Source:      *source,
		Destination: *dest,
		Action:      action,
	}

	r.Attributes = getOptionalAttributes(relationship.Attributes, attrs.Common)

	return r, nil
}
