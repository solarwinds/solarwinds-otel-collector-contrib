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
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal/models"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type EventDetector struct {
	entities     map[string]config.Entity
	logEvents    config.EventsGroup[ottllog.TransformContext]
	metricEvents config.EventsGroup[ottlmetric.TransformContext]
	sourcePrefix string
	destPrefix   string
}

func NewEventDetector(
	entities map[string]config.Entity,
	sourcePrefix, destPrefix string,
	logEvents config.EventsGroup[ottllog.TransformContext],
	metricEvents config.EventsGroup[ottlmetric.TransformContext],
) *EventDetector {
	return &EventDetector{
		entities:     entities,
		logEvents:    logEvents,
		metricEvents: metricEvents,
		sourcePrefix: sourcePrefix,
		destPrefix:   destPrefix,
	}
}

func (e *EventDetector) DetectLog(ctx context.Context, resourceAttrs pcommon.Map, transformCtx ottllog.TransformContext) ([]models.Subject, error) {
	ee, re, err := processEvents(ctx, e.logEvents, transformCtx)
	if err != nil {
		return nil, err
	}

	events := make([]models.Subject, 0, len(ee)+len(re))
	for _, entityEvent := range ee {
		newEvent, _ := e.createEntity(resourceAttrs, entityEvent)
		events = append(events, newEvent)
	}

	return events, err
}
func (e *EventDetector) DetectMetric(ctx context.Context, resourceAttrs pcommon.Map, transformCtx ottlmetric.TransformContext) ([]models.Subject, error) {
	ee, re, err := processEvents(ctx, e.metricEvents, transformCtx)
	if err != nil {
		return nil, err
	}

	events := make([]models.Subject, 0, len(ee)+len(re))
	for _, entityEvent := range ee {
		newEvent, _ := e.createEntity(resourceAttrs, entityEvent)
		events = append(events, newEvent)
	}

	return events, err
}

func (e *EventDetector) createEntity(resourceAttrs pcommon.Map, event *config.EntityEvent) (models.Entity, error) {
	ids := pcommon.NewMap()
	entity := e.entities[event.Type]
	err := setIdAttributes1(ids, entity.IDs, resourceAttrs)
	if err != nil {
		return models.Entity{}, fmt.Errorf("failed to set ID attributes for entity %s: %w", entity.Type, err)
	}

	attrs := pcommon.NewMap()
	setAttributes1(attrs, entity.Attributes, resourceAttrs)

	return models.Entity{
		Action:     models.Update,
		Type:       entity.Type,
		IDs:        ids,
		Attributes: attrs,
	}, nil
}

func (e *EventDetector) createRelationship(resourceAttrs pcommon.Map, relationship *config.RelationshipEvent) (*models.Relationship, error) {
	sourceIds := pcommon.NewMap()
	source, _ := e.entities[relationship.Source]
	dest, _ := e.entities[relationship.Destination]
	hasSourcePrefix, err := setIdAttributesWithPrefix(sourceIds, source.IDs, resourceAttrs, relationship.Source, e.sourcePrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to set source entity ID attributes: %w", err)
	}
	if !hasSourcePrefix {
		return nil, fmt.Errorf("missing source entity ID attributes with prefix %s", e.sourcePrefix)
	}

	destIds := pcommon.NewMap()
	hasDestPrefix, err := setIdAttributesWithPrefix(destIds, dest.IDs, resourceAttrs, relationship.Destination, e.destPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to set destination entity ID attributes: %w", err)
	}
	if !hasDestPrefix {
		return nil, fmt.Errorf("missing destination entity ID attributes with prefix %s", e.destPrefix)
	}

	relAttrs := pcommon.NewMap()
	setAttributes1(relAttrs, relationship.Attributes, resourceAttrs)

	return &models.Relationship{
		Action: models.Update,
		Type:   relationship.Type,
		Source: models.RelationshipEntity{
			Type: relationship.Source,
			IDs:  sourceIds,
		},
		Destination: models.RelationshipEntity{
			Type: relationship.Destination,
			IDs:  destIds,
		},
		Attributes: relAttrs,
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

// setIdAttributes sets the entity id attributes in the log record as needed by SWO.
// Attributes are used to infer the entity in the system.
//
// Returns error if any of the attributes are missing in the resourceAttrs.
// If any ID attribute is missing, the entity would not be inferred.
func setIdAttributes1(attrs pcommon.Map, entityIds []string, resourceAttrs pcommon.Map) error {
	if len(entityIds) == 0 {
		return fmt.Errorf("entity ID attributes are empty")
	}

	for _, id := range entityIds {

		value, exists := findAttribute(id, resourceAttrs)
		if !exists {
			return fmt.Errorf("missing entity ID attribute: %s", id)
		}

		putAttribute1(&attrs, id, value)
	}
	return nil
}

// setEntityAttributes sets the entity attributes in the log record as needed by SWO.
// Attributes are used to update the entity.
func setAttributes1(attrs pcommon.Map, entityAttrs []string, resourceAttrs pcommon.Map) {
	if len(entityAttrs) == 0 {
		return
	}

	for _, attr := range entityAttrs {
		value, exists := findAttribute1(attr, resourceAttrs)
		if !exists {
			continue
		}
		putAttribute1(&attrs, attr, value)
	}
}

// setIdAttributesWithPrefix sets the entity id attributes in the log record as needed by SWO for same type relationships.
// Verifies that prefix is present in the resource attributes at least once within entity IDs.
func setIdAttributesWithPrefix1(attrs pcommon.Map, entityIds []string, resourceAttrs pcommon.Map, name, prefix string) (bool, error) {
	if len(entityIds) == 0 {
		return false, fmt.Errorf("entity ID attributes are empty")
	}

	hasPrefix := false
	logIds := attrs.PutEmptyMap(name)
	for _, id := range entityIds {

		value, exists := findAttribute1(prefix+id, resourceAttrs)
		if exists {
			putAttribute1(&logIds, id, value)
			hasPrefix = true
			continue
		}

		value, exists = findAttribute1(id, resourceAttrs)
		if exists {
			putAttribute1(&logIds, id, value)
			continue
		}

		return false, fmt.Errorf("missing entity ID attribute: %s", id)
	}
	return hasPrefix, nil
}

// findAttribute checks if the attribute identified as key exists in the source pcommon.Map.
func findAttribute1(key string, src pcommon.Map) (pcommon.Value, bool) {
	attrVal, ok := src.Get(key)
	return attrVal, ok
}

// putAttribute copies the value of attribute identified as key, to destination pcommon.Map.
func putAttribute1(dest *pcommon.Map, key string, attrValue pcommon.Value) {
	switch typeAttr := attrValue.Type(); typeAttr {
	case pcommon.ValueTypeInt:
		dest.PutInt(key, attrValue.Int())
	case pcommon.ValueTypeDouble:
		dest.PutDouble(key, attrValue.Double())
	case pcommon.ValueTypeBool:
		dest.PutBool(key, attrValue.Bool())
	case pcommon.ValueTypeBytes:
		value := attrValue.Bytes().AsRaw()
		dest.PutEmptyBytes(key).FromRaw(value)
	default:
		dest.PutStr(key, attrValue.Str())
	}
}
