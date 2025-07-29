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
	events          []config.ParsedEventInterface[T]
	keyBuilder      KeyBuilder
	logger          *zap.Logger
}

func NewEventDetector[T any](
	attributeMapper AttributeMapper,
	events []config.ParsedEventInterface[T],
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
	configuredEvents, err := e.processEvents(ctx, transformCtx)
	if err != nil {
		return nil, err
	}
	return e.collectEvents(resourceAttrs, configuredEvents)
}

// collectEvents processes the attributes and configured events to create a list of detected events.
// It processes entity and relationship events in a single pass, handling both types uniformly.
func (e *EventDetector[T]) collectEvents(
	attrs Attributes,
	configuredEvents []config.ParsedEventInterface[T],
) ([]Event, error) {
	allEvents := make([]Event, 0)
	seenEntities := make(map[string]struct{})

	// Build a lookup map for configured entity types for efficient inference
	entityConfig := e.buildEntityTypeMap(configuredEvents)

	for _, event := range configuredEvents {
		if event.IsEntityEvent() {
			if entity := e.processEntityEvent(event.GetEntityEvent(), attrs, seenEntities); entity != nil {
				allEvents = append(allEvents, *entity)
			}
		} else {
			if relationship := e.processRelationshipEvent(event.GetRelationshipEvent(), attrs, entityConfig, seenEntities, &allEvents); relationship != nil {
				allEvents = append(allEvents, relationship)
			}
		}
	}

	return allEvents, nil
}

// buildEntityTypeMap creates a lookup map for configured entity types and their actions
func (e *EventDetector[T]) buildEntityTypeMap(configuredEvents []config.ParsedEventInterface[T]) map[string]string {
	entityConfig := make(map[string]string)
	for _, event := range configuredEvents {
		if event.IsEntityEvent() {
			entityEvent := event.GetEntityEvent()
			entityConfig[entityEvent.Entity] = entityEvent.Action
		}
	}
	return entityConfig
}

// processEntityEvent processes a single entity event and returns the entity if valid
func (e *EventDetector[T]) processEntityEvent(
	entityEvent *config.EntityEvent,
	attrs Attributes,
	seenEntities map[string]struct{},
) *Entity {
	entity, err := e.attributeMapper.getEntity(entityEvent.Entity, attrs)
	if err != nil {
		e.logger.Debug("failed to create entity event", zap.Error(err))
		return nil
	}

	entity.Action = entityEvent.Action
	entityHash, err := e.keyBuilder.BuildEntityKey(*entity)
	if err != nil {
		e.logger.Debug("failed to build entity hash", zap.Error(err))
		return nil
	}

	if _, exists := seenEntities[entityHash]; exists {
		return nil // Already seen, skip duplicate
	}

	seenEntities[entityHash] = struct{}{}
	return entity
}

// processRelationshipEvent processes a single relationship event and returns the relationship if valid
func (e *EventDetector[T]) processRelationshipEvent(
	relationshipEvent *config.RelationshipEvent,
	attrs Attributes,
	entityConfig map[string]string,
	seenEntities map[string]struct{},
	allEvents *[]Event,
) Event {
	sourceEntity, destEntity, err := e.attributeMapper.getRelationshipEntities(
		relationshipEvent.Source, relationshipEvent.Destination, attrs)
	if err != nil {
		e.logger.Debug("failed to create relationship entities", zap.Error(err))
		return nil
	}

	relationship, err := createRelationship(relationshipEvent, sourceEntity, destEntity, attrs)
	if err != nil {
		e.logger.Debug("failed to create relationship event", zap.Error(err))
		return nil
	}

	// Add inferred entities if they're configured and not duplicates
	if inferredEntity := e.getInferredEntity(sourceEntity, entityConfig, seenEntities); inferredEntity != nil {
		*allEvents = append(*allEvents, *inferredEntity)
	}
	if inferredEntity := e.getInferredEntity(destEntity, entityConfig, seenEntities); inferredEntity != nil {
		*allEvents = append(*allEvents, *inferredEntity)
	}

	return &relationship
}

// getInferredEntity returns an entity inferred from a relationship if it's configured and not duplicate
func (e *EventDetector[T]) getInferredEntity(
	entity *Entity,
	entityConfig map[string]string,
	seenEntities map[string]struct{},
) *Entity {
	action, isConfigured := entityConfig[entity.Type]
	if !isConfigured {
		return nil
	}

	entityHash, err := e.keyBuilder.BuildEntityKey(*entity)
	if err != nil {
		e.logger.Debug("failed to build entity hash for inferred entity", zap.Error(err))
		return nil
	}

	if _, exists := seenEntities[entityHash]; exists {
		return nil // Already seen, skip duplicate
	}

	seenEntities[entityHash] = struct{}{}
	entity.Action = action
	return entity
}

// processEvents evaluates the conditions for entity and relationship events.
// Only events with conditions that evaluate to true are returned.
func (e *EventDetector[T]) processEvents(
	ctx context.Context,
	tc T) ([]config.ParsedEventInterface[T], error) {

	events := make([]config.ParsedEventInterface[T], 0, len(e.events))

	for _, event := range e.events {
		conditionSeq := event.GetConditionSeq()
		ok, err := conditionSeq.Eval(ctx, tc)
		if err != nil {
			return nil, err
		}

		if ok {
			events = append(events, event)
		}
	}
	return events, nil
}

func createRelationship(relationship *config.RelationshipEvent, source, dest *Entity, attrs Attributes) (Relationship, error) {
	action, err := GetActionString(relationship.Action)
	if err != nil {
		return Relationship{}, fmt.Errorf("failed to get action type for relationship event")
	}

	return Relationship{
		Type:        relationship.Type,
		Source:      *source,
		Destination: *dest,
		Action:      action,
		Attributes:  getOptionalAttributes(relationship.Attributes, attrs.Common),
	}, nil
}
