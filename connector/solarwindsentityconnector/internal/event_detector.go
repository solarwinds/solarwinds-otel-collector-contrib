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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type EventDetector struct {
	entities         map[string]config.Entity
	entityIdentifier EntityIdentifier
	logEvents        config.EventsGroup[ottllog.TransformContext]
	metricEvents     config.EventsGroup[ottlmetric.TransformContext]
	sourcePrefix     string
	destPrefix       string
	logger           *zap.Logger
}

func NewEventDetector(
	entities map[string]config.Entity,
	sourcePrefix, destPrefix string,
	logEvents config.EventsGroup[ottllog.TransformContext],
	metricEvents config.EventsGroup[ottlmetric.TransformContext],
	logger *zap.Logger,
) *EventDetector {
	ei := EntityIdentifier{entities: entities}
	return &EventDetector{
		entities:         entities,
		entityIdentifier: ei,
		logEvents:        logEvents,
		metricEvents:     metricEvents,
		sourcePrefix:     sourcePrefix,
		destPrefix:       destPrefix,
		logger:           logger,
	}
}

func (e *EventDetector) DetectLog(ctx context.Context, resourceAttrs pcommon.Map, transformCtx ottllog.TransformContext) ([]Event, error) {
	ee, re, err := processEvents(ctx, e.logEvents, transformCtx)
	if err != nil {
		return nil, err
	}
	attrs := IdentifyAttributes(resourceAttrs, "src", "dst")

	return e.collectEvents(resourceAttrs, attrs, ee, re)
}

func (e *EventDetector) DetectMetric(ctx context.Context, resourceAttrs pcommon.Map, transformCtx ottlmetric.TransformContext) ([]Event, error) {
	ee, re, err := processEvents(ctx, e.metricEvents, transformCtx)
	if err != nil {
		return nil, err
	}
	attrs := IdentifyAttributes(resourceAttrs, "src", "dst")

	return e.collectEvents(resourceAttrs, attrs, ee, re)

}

func (e *EventDetector) collectEvents(resAttrs pcommon.Map, attrs Attributes, ee []*config.EntityEvent, re []*config.RelationshipEvent) ([]Event, error) {
	events := make([]Event, 0, len(ee)+len(re))
	for _, entityEvent := range ee {
		newEvents, err := e.entityIdentifier.getEntities(entityEvent.Type, attrs)
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
		newRel, err := e.entityIdentifier.getRelationship(relationshipEvent, attrs)
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

func (e *EventDetector) createRelationshipEvent(resourceAttrs pcommon.Map, relationship *config.RelationshipEvent) (*Relationship, error) {
	source, ok := e.entities[relationship.Source]
	if !ok {
		return nil, fmt.Errorf("bad source entity")
	}

	dest, ok := e.entities[relationship.Destination]
	if !ok {
		return nil, fmt.Errorf("bad destination entity")
	}

	lr := plog.NewLogRecord()
	attrs := lr.Attributes()

	if source.Type == dest.Type {
		if err := e.setAttributesForSameTypeRelationships(attrs, source, dest, resourceAttrs); err != nil {
			return nil, err
		}

	} else {
		if err := e.setAttributesForDifferentTypeRelationships(attrs, source, dest, resourceAttrs); err != nil {
			return nil, err
		}
	}

	relationshipAttrs := getAttributes(relationship.Attributes, resourceAttrs)

	sourceIds, _ := attrs.Get(relationshipSrcEntityIds)
	destIds, _ := attrs.Get(relationshipDestEntityIds)
	action, err := GetActionString(relationship.Action)
	if err != nil {
		e.logger.Debug("failed to get action type for relationship event", zap.Error(err))
		return nil, err
	}
	return &Relationship{
		Action: action,
		Type:   relationship.Type,
		Source: RelationshipEntity{
			Type: relationship.Source,
			IDs:  sourceIds.Map(),
		},
		Destination: RelationshipEntity{
			Type: relationship.Destination,
			IDs:  destIds.Map(),
		},
		Attributes: relationshipAttrs,
	}, nil
}

// ProcessEvents evaluates the conditions for entities and relationships events.
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

func (e *EventDetector) setAttributesForSameTypeRelationships(attrs pcommon.Map, source config.Entity, dest config.Entity, resourceAttrs pcommon.Map) error {
	if e.sourcePrefix == "" || e.destPrefix == "" {
		return fmt.Errorf("prefixes are mandatory for same type relationships")
	}

	hasPrefixSrc, err := setIdAttributesForRelationships(attrs, source.IDs, resourceAttrs, relationshipSrcEntityIds, e.sourcePrefix)
	if err != nil || !hasPrefixSrc {
		return fmt.Errorf("missing prefixed ID attribute for source entity")
	}

	hasPrefixDst, err := setIdAttributesForRelationships(attrs, dest.IDs, resourceAttrs, relationshipDestEntityIds, e.destPrefix)
	if err != nil || !hasPrefixDst {
		return fmt.Errorf("missing prefixed ID attribute for destination entity")
	}
	return nil
}

func (e *EventDetector) setAttributesForDifferentTypeRelationships(attrs pcommon.Map, source config.Entity, dest config.Entity, resourceAttrs pcommon.Map) error {
	// For different type relationships, prefixes are optional.
	_, err := setIdAttributesForRelationships(attrs, source.IDs, resourceAttrs, relationshipSrcEntityIds, e.sourcePrefix)
	if err != nil {
		return fmt.Errorf("missing ID attribute for source entity")
	}

	_, err = setIdAttributesForRelationships(attrs, dest.IDs, resourceAttrs, relationshipDestEntityIds, e.destPrefix)
	if err != nil {
		return fmt.Errorf("missing ID attribute for destination entity")
	}
	return nil
}
