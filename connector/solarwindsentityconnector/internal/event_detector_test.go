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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestDetect_EntityAndRelationshipEvents(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	// Prepare entity and relationship events
	relationshipType := "MemberOf"
	entity := config.Entity{
		Entity:     "Entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.EntityParsedEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: entity.Entity, Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}
	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: relationshipType, Source: entity.Entity, Destination: entity.Entity, Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{entityEvent, relationshipEvent},
	}

	// Prepare resource attributes
	attributes := Attributes{
		Source: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue1"),
			"attr": pcommon.NewValueStr("attrvalue1"),
		},
		Destination: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue2"),
			"attr": pcommon.NewValueStr("attrvalue2"),
		},
		Common: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue1"),
			"attr": pcommon.NewValueStr("attrvalue1"),
		},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})

	// Create EventDetector
	eventDetector := NewEventDetector(
		attributeMapper,
		eventsGroup,
		zap.NewNop(),
	)

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 3)
}

func TestDetect_NoEvents(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to false
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"false"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entity := config.Entity{
		Entity:     "Entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.EntityParsedEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: entity.Entity},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{entityEvent},
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue"),
			"attr": pcommon.NewValueStr("attrvalue"),
		},
	}

	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})

	eventDetector := NewEventDetector(
		attributeMapper,
		eventsGroup,
		zap.NewNop(),
	)

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 0)
}

func TestDetect_ConditionTrue_EventsCreated(t *testing.T) {
	// Initialize test data
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()
	logs := plog.NewLogs()

	parserLog, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parserLog.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entity := config.Entity{
		Entity:     "test-entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.EntityParsedEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: "test-entity", Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}
	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: "test-rel", Source: "test-entity", Destination: "test-entity", Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{entityEvent, relationshipEvent},
	}

	attributes := Attributes{
		Source: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue1"),
			"attr": pcommon.NewValueStr("attrvalue1"),
		},
		Destination: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue2"),
			"attr": pcommon.NewValueStr("attrvalue2"),
		},
		Common: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue1"),
			"attr": pcommon.NewValueStr("attrvalue1"),
		},
	}

	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)
	am := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})
	ev := NewEventDetector(am, eventsGroup, zap.NewNop())

	// Tested function
	events, err := ev.Detect(ctx, attributes, tc)
	require.NoError(t, err)

	// Should detect 1 entity event, 1 relationship event, and 1 additional entity event inferred from the relationship (same-type relationship)
	require.Len(t, events, 3)
}

func TestDetect_ConditionFalse_EventsNotCreated(t *testing.T) {
	// Initialize test data
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()
	logs := plog.NewLogs()

	parserLog, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parserLog.ParseConditions([]string{"false"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entity := config.Entity{
		Entity:     "test-entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.EntityParsedEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: "test-entity", Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}
	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: "test-rel", Source: "test-entity", Destination: "test-entity", Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{entityEvent, relationshipEvent},
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue1"),
			"attr": pcommon.NewValueStr("attrvalue1"),
		},
	}

	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)
	am := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})
	ev := NewEventDetector(am, eventsGroup, zap.NewNop())

	// Tested function
	events, err := ev.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestDetect_EntityAndRelationshipEventsWithDifferentTypes(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	// Prepare attributes for different entity types
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"id2":   pcommon.NewValueStr("idvalue2"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
		},
	}

	srcEntity := config.Entity{
		Entity:     "KubernetesCluster",
		IDs:        []string{"id1"},
		Attributes: []string{"attr1"},
	}

	destEntity := config.Entity{
		Entity:     "KubernetesNode",
		IDs:        []string{"id2"},
		Attributes: []string{"attr2"},
	}

	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition: &config.RelationshipEvent{
			Type:        "MemberOf",
			Source:      srcEntity.Entity,
			Destination: destEntity.Entity,
			Event:       config.Event{Action: EventUpdateAction},
		},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{relationshipEvent},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Entity:  srcEntity,
		destEntity.Entity: destEntity,
	})

	// Create EventDetector
	eventDetector := NewEventDetector(
		attributeMapper,
		eventsGroup,
		zap.NewNop(),
	)

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 1)

	// Should be a relationship event
	relationship, ok := events[0].(*Relationship)
	require.True(t, ok)
	assert.Equal(t, "MemberOf", relationship.Type)
	assert.Equal(t, srcEntity.Entity, relationship.Source.Type)
	assert.Equal(t, destEntity.Entity, relationship.Destination.Type)
}

func TestDetect_EntityAndRelationshipEventsWithSameType(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	attributes := Attributes{
		Source: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue1"),
		},
		Destination: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue2"),
		},
		Common: map[string]pcommon.Value{
			"attr1": pcommon.NewValueStr("attrvalue1"),
		},
	}

	entity := config.Entity{
		Entity:     "KubernetesCluster",
		IDs:        []string{"id"},
		Attributes: []string{"attr1"},
	}

	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition: &config.RelationshipEvent{
			Event:       config.Event{Action: EventUpdateAction},
			Type:        "Has",
			Source:      "KubernetesCluster",
			Destination: "KubernetesCluster",
		},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{relationshipEvent},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster": entity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		eventsGroup,
		zap.NewNop(),
	)

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 1)

	// Should be a relationship event
	relationship, ok := events[0].(*Relationship)
	require.True(t, ok)
	assert.Equal(t, "Has", relationship.Type)
	assert.Equal(t, entity.Entity, relationship.Source.Type)
	assert.Equal(t, entity.Entity, relationship.Destination.Type)
}

func TestDetect_EntityEvents_AttributesPresent(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	testEntity := config.Entity{
		Entity:     "testEntityType",
		IDs:        []string{"id1", "id2"},
		Attributes: []string{"attr1", "attr2"},
	}

	entityEvent := config.EntityParsedEvent[ottllog.TransformContext]{
		Definition: &config.EntityEvent{
			Entity: testEntity.Entity,
			Event:  config.Event{Action: EventUpdateAction},
		},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{entityEvent},
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"id2":   pcommon.NewValueStr("idvalue2"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
			"attr2": pcommon.NewValueStr("attrvalue2"),
		},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Entity: testEntity})
	eventDetector := NewEventDetector(attributeMapper, eventsGroup, zap.NewNop())

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.Len(t, events, 1)

	e1, ok := events[0].(Entity)
	require.True(t, ok)
	assert.Equal(t, "testEntityType", e1.Type)
	assert.Equal(t, "update", e1.Action)

	assert.Equal(t, 2, e1.IDs.Len())
	assertAttributeIsPresent(t, e1.IDs, "id1", "idvalue1")
	assertAttributeIsPresent(t, e1.IDs, "id2", "idvalue2")

	assert.Equal(t, 2, e1.Attributes.Len())
	assertAttributeIsPresent(t, e1.Attributes, "attr1", "attrvalue1")
	assertAttributeIsPresent(t, e1.Attributes, "attr2", "attrvalue2")
}

func TestDetect_EntityEvents_IDAttributesMissing(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	testEntity := config.Entity{Entity: "testEntityType", IDs: []string{"id1", "id2"}, Attributes: []string{}}

	entityEvent := config.EntityParsedEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: testEntity.Entity, Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{entityEvent},
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1": pcommon.NewValueStr("idvalue1"),
		},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Entity: testEntity})
	eventDetector := NewEventDetector(attributeMapper, eventsGroup, zap.NewNop())

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)

	// Should be empty because required ID attributes are missing
	require.Len(t, events, 0)
}

func TestDetect_EntityEvents_SomeAttributesMissing(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	testEntity := config.Entity{Entity: "testEntityType", IDs: []string{"id1"}, Attributes: []string{"attr1", "attr2"}}

	entityEvent := config.EntityParsedEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: testEntity.Entity, Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{entityEvent},
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
		},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Entity: testEntity})
	eventDetector := NewEventDetector(attributeMapper, eventsGroup, zap.NewNop())

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.Len(t, events, 1)

	entityEvent1, ok := events[0].(Entity)
	require.True(t, ok)

	assert.Equal(t, testEntity.Entity, entityEvent1.Type)
	assert.Equal(t, "update", entityEvent1.Action)
	assert.Equal(t, 1, entityEvent1.IDs.Len())
	assertAttributeIsPresent(t, entityEvent1.IDs, "id1", "idvalue1")
	assert.Equal(t, 1, entityEvent1.Attributes.Len())
	assertAttributeIsPresent(t, entityEvent1.Attributes, "attr1", "attrvalue1")
}

func TestDetect_RelationshipEvents_AttributesPresent(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	srcEntity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id1"}, Attributes: []string{"attr1"}}
	destEntity := config.Entity{Entity: "KubernetesNamespace", IDs: []string{"id2"}, Attributes: []string{"attr2"}}

	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition: &config.RelationshipEvent{
			Source:      srcEntity.Entity,
			Destination: destEntity.Entity,
			Event:       config.Event{Action: EventUpdateAction},
		},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{relationshipEvent},
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"id2":   pcommon.NewValueStr("idvalue2"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
			"attr2": pcommon.NewValueStr("attrvalue2"),
		},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Entity:  srcEntity,
		destEntity.Entity: destEntity,
	})
	eventDetector := NewEventDetector(attributeMapper, eventsGroup, zap.NewNop())

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.Len(t, events, 1)

	relationship, ok := events[0].(*Relationship)
	require.True(t, ok)

	assert.Equal(t, "update", relationship.Action)
	assert.Equal(t, srcEntity.Entity, relationship.Source.Type)
	assert.Equal(t, destEntity.Entity, relationship.Destination.Type)

	assert.Equal(t, 1, relationship.Source.IDs.Len())
	assertAttributeIsPresent(t, relationship.Source.IDs, "id1", "idvalue1")

	assert.Equal(t, 1, relationship.Destination.IDs.Len())
	assertAttributeIsPresent(t, relationship.Destination.IDs, "id2", "idvalue2")
}

func TestDetect_RelationshipEvents_SameType_AttributesPresent(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id"}, Attributes: []string{"attr"}}

	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition: &config.RelationshipEvent{
			Source:      "KubernetesCluster",
			Destination: "KubernetesCluster",
			Event:       config.Event{Action: EventUpdateAction},
		},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{relationshipEvent},
	}

	attributes := Attributes{
		Source: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue1"),
		},
		Destination: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue2"),
		},
		Common: map[string]pcommon.Value{
			"attr": pcommon.NewValueStr("attrvalue"),
		},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})
	eventDetector := NewEventDetector(attributeMapper, eventsGroup, zap.NewNop())

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.Len(t, events, 1)

	relationship, ok := events[0].(*Relationship)
	require.True(t, ok)

	// assert
	assert.Equal(t, "update", relationship.Action)
	assert.Equal(t, entity.Entity, relationship.Source.Type)
	assert.Equal(t, entity.Entity, relationship.Destination.Type)

	assert.Equal(t, 1, relationship.Source.IDs.Len())
	assertAttributeIsPresent(t, relationship.Source.IDs, "id", "idvalue1")

	assert.Equal(t, 1, relationship.Destination.IDs.Len())
	assertAttributeIsPresent(t, relationship.Destination.IDs, "id", "idvalue2")
}

func TestDetect_RelationshipEvents_IDAttributesMissing(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	srcEntity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id1"}, Attributes: []string{}}
	destEntity := config.Entity{Entity: "KubernetesNamespace", IDs: []string{"id2"}, Attributes: []string{}}

	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition: &config.RelationshipEvent{
			Source:      "KubernetesCluster",
			Destination: "KubernetesNamespace",
		},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{relationshipEvent},
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1": pcommon.NewValueStr("idvalue1"),
		},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Entity:  srcEntity,
		destEntity.Entity: destEntity,
	})
	eventDetector := NewEventDetector(attributeMapper, eventsGroup, zap.NewNop())

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)

	// Should be empty because required ID attributes are missing for destination entity
	require.Len(t, events, 0)
}

func TestDetect_RelationshipEvents_WithRelationshipAttributes(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	srcEntity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id1"}}
	destEntity := config.Entity{Entity: "KubernetesNamespace", IDs: []string{"id2"}}

	relationshipEvent := config.RelationshipParsedEvent[ottllog.TransformContext]{
		Definition: &config.RelationshipEvent{
			Source:      "KubernetesCluster",
			Destination: "KubernetesNamespace",
			Attributes:  []string{"relationshipAttr"},
			Event:       config.Event{Action: EventUpdateAction},
		},
		ConditionSeq: seq,
	}

	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Events: []config.ParsedEventInterface[ottllog.TransformContext]{relationshipEvent},
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":              pcommon.NewValueStr("idvalue1"),
			"id2":              pcommon.NewValueStr("idvalue2"),
			"relationshipAttr": pcommon.NewValueStr("relationshipValue"),
		},
	}

	// Prepare transform context
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster":   srcEntity,
		"KubernetesNamespace": destEntity,
	})
	eventDetector := NewEventDetector(attributeMapper, eventsGroup, zap.NewNop())

	events, err := eventDetector.Detect(ctx, attributes, tc)
	require.NoError(t, err)
	require.Len(t, events, 1)

	r1, ok := events[0].(*Relationship)
	require.True(t, ok)
	assert.Equal(t, "KubernetesCluster", r1.Source.Type)
	assert.Equal(t, "KubernetesNamespace", r1.Destination.Type)
	assert.Equal(t, "update", r1.Action)

	// assert relationship attributes
	assert.Equal(t, 1, r1.Attributes.Len())
	assertAttributeIsPresent(t, r1.Attributes, "relationshipAttr", "relationshipValue")
}

func assertAttributeIsPresent(t *testing.T, attrs pcommon.Map, key string, expected string) {
	if val, ok := attrs.Get(key); ok {
		assert.Equal(t, true, ok)
		assert.Equal(t, expected, val.Str())
	}
}
