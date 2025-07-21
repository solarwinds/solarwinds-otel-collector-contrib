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

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
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

func TestDetectLog_EntityAndRelationshipEvents(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()
	logger := zap.NewNop()

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
	entityEvent := config.ParsedEntityEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: entity.Entity, Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}
	relationshipEvent := config.ParsedRelationshipEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: relationshipType, Source: entity.Entity, Destination: entity.Entity, Event: config.Event{Action: EventUpdateAction}},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Entities:      []config.ParsedEntityEvent[ottllog.TransformContext]{entityEvent},
		Relationships: []config.ParsedRelationshipEvent[ottllog.TransformContext]{relationshipEvent},
		Parser:        &parser,
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
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)

	events, err := eventDetector.DetectLog(ctx, attributes, tc)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 3)
}

func TestDetectLog_NoEvents(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()
	logger := zap.NewNop()

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
	entityEvent := config.ParsedEntityEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: entity.Entity},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Entities:      []config.ParsedEntityEvent[ottllog.TransformContext]{entityEvent},
		Relationships: nil,
		Parser:        &parser,
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
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)

	events, err := eventDetector.DetectLog(ctx, attributes, tc)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 0)
}

func TestDetectMetric_EntityAndRelationshipEvents(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()
	logger := zap.NewNop()

	// Prepare OTTL parser and condition sequence that always evaluates to true
	parser, err := ottlmetric.NewParser(nil, settings)
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
	entityEvent := config.ParsedEntityEvent[ottlmetric.TransformContext]{
		Definition: &config.EntityEvent{
			Event:  config.Event{Action: EventUpdateAction},
			Entity: entity.Entity,
		},
		ConditionSeq: seq,
	}
	relationshipEvent := config.ParsedRelationshipEvent[ottlmetric.TransformContext]{
		Definition: &config.RelationshipEvent{
			Event:       config.Event{Action: EventUpdateAction},
			Type:        relationshipType,
			Source:      entity.Entity,
			Destination: entity.Entity,
		},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottlmetric.TransformContext]{
		Entities:      []config.ParsedEntityEvent[ottlmetric.TransformContext]{entityEvent},
		Relationships: []config.ParsedRelationshipEvent[ottlmetric.TransformContext]{relationshipEvent},
		Parser:        &parser,
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
	metrics := pmetric.NewMetrics()
	rMetrics := metrics.ResourceMetrics().AppendEmpty()
	resource := rMetrics.Resource()
	scopeMetrics := rMetrics.ScopeMetrics().AppendEmpty()
	scope := scopeMetrics.Scope()
	metricSlice := scopeMetrics.Metrics()
	metric := metricSlice.AppendEmpty()
	tc := ottlmetric.NewTransformContext(metric, metricSlice, scope, resource, scopeMetrics, rMetrics)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})

	// Create EventDetector
	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		eventsGroup,
		logger,
	)

	events, err := eventDetector.DetectMetric(ctx, attributes, tc)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Len(t, events, 3)
}

func TestDetectMetric_NoEvents(t *testing.T) {
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()
	logger := zap.NewNop()

	// Prepare OTTL parser and condition sequence that always evaluates to false
	parser, err := ottlmetric.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parser.ParseConditions([]string{"false"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entity := config.Entity{
		Entity:     "Entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.ParsedEntityEvent[ottlmetric.TransformContext]{
		Definition:   &config.EntityEvent{Entity: entity.Entity},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottlmetric.TransformContext]{
		Entities:      []config.ParsedEntityEvent[ottlmetric.TransformContext]{entityEvent},
		Relationships: nil,
		Parser:        &parser,
	}

	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id":   pcommon.NewValueStr("idvalue"),
			"attr": pcommon.NewValueStr("attrvalue"),
		},
	}

	metrics := pmetric.NewMetrics()
	rMetrics := metrics.ResourceMetrics().AppendEmpty()
	resource := rMetrics.Resource()
	scopeMetrics := rMetrics.ScopeMetrics().AppendEmpty()
	scope := scopeMetrics.Scope()
	metricSlice := scopeMetrics.Metrics()
	metric := metricSlice.AppendEmpty()
	tc := ottlmetric.NewTransformContext(metric, metricSlice, scope, resource, scopeMetrics, rMetrics)

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		eventsGroup,
		logger,
	)

	events, err := eventDetector.DetectMetric(ctx, attributes, tc)
	require.NoError(t, err)
	require.NotNil(t, events)
	require.Empty(t, events)
}

func TestProcessEvents_ConditionTrue_EventsCreated(t *testing.T) {
	// Initialize test data
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()
	logs := plog.NewLogs()

	parserLog, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parserLog.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entityEvent := config.ParsedEntityEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: "test-entity"},
		ConditionSeq: seq,
	}
	relationshipEvent := config.ParsedRelationshipEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: "test-rel"},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Entities:      []config.ParsedEntityEvent[ottllog.TransformContext]{entityEvent},
		Relationships: []config.ParsedRelationshipEvent[ottllog.TransformContext]{relationshipEvent},
		Parser:        &parserLog,
	}

	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	// Tested function
	entities, relationships, err := processEvents(ctx, eventsGroup, tc)
	require.NoError(t, err)

	require.Len(t, entities, 1)
	require.Len(t, relationships, 1)
}

func TestProcessEvents_ConditionFalse_EventsNotCreated(t *testing.T) {
	// Initialize test data
	ctx := context.Background()
	settings := componenttest.NewNopTelemetrySettings()
	logs := plog.NewLogs()

	parserLog, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parserLog.ParseConditions([]string{"false"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entityEvent := config.ParsedEntityEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Entity: "test-entity"},
		ConditionSeq: seq,
	}
	relationshipEvent := config.ParsedRelationshipEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: "test-rel"},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Entities:      []config.ParsedEntityEvent[ottllog.TransformContext]{entityEvent},
		Relationships: []config.ParsedRelationshipEvent[ottllog.TransformContext]{relationshipEvent},
		Parser:        &parserLog,
	}

	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	// Tested function
	entities, relationships, err := processEvents(ctx, eventsGroup, tc)
	require.NoError(t, err)
	require.Empty(t, entities)
	require.Empty(t, relationships)
}

func TestGetRelationships_WithDifferentTypes(t *testing.T) {
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

	relationships := []*config.RelationshipEvent{{
		Type:        "MemberOf",
		Action:      "update",
		Source:      srcEntity.Entity,
		Destination: destEntity.Entity,
	},
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Entity:  srcEntity,
		destEntity.Entity: destEntity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	relationshipEvents := eventDetector.getRelationshipEvents(attributes, relationships)
	assert.Equal(t, 1, len(relationshipEvents))
	r1 := relationshipEvents[0]
	assert.Equal(t, "MemberOf", r1.Type)

	assert.Equal(t, srcEntity.Entity, r1.Source.Type)
	assert.Equal(t, destEntity.Entity, r1.Destination.Type)
}

func TestGetRelationships_WithSameType(t *testing.T) {
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

	relationships := []*config.RelationshipEvent{{
		Action:      "update",
		Type:        "Has",
		Source:      "KubernetesCluster",
		Destination: "KubernetesCluster",
	}}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster": entity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	relationshipEvents := eventDetector.getRelationshipEvents(attributes, relationships)
	assert.Equal(t, 1, len(relationshipEvents))
	r1 := relationshipEvents[0]
	assert.Equal(t, "Has", r1.Type)

	assert.Equal(t, entity.Entity, r1.Source.Type)
	assert.Equal(t, entity.Entity, r1.Destination.Type)
}

func TestCollectEvents_WithEntity_AttributesPresent(t *testing.T) {
	// arrange
	testEntity := config.Entity{
		Entity:     "testEntityType",
		IDs:        []string{"id1", "id2"},
		Attributes: []string{"attr1", "attr2"},
	}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"id2":   pcommon.NewValueStr("idvalue2"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
			"attr2": pcommon.NewValueStr("attrvalue2"),
		},
	}

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Entity: testEntity})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	events, err := eventBuilder.collectEvents(attributes, []*config.EntityEvent{
		{
			Entity: testEntity.Entity,
			Action: EventUpdateAction},
	},
		nil)
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

func TestCollectEvents_WithEntity_IDAttributesMissing(t *testing.T) {
	// arrange
	testEntity := config.Entity{Entity: "testEntityType", IDs: []string{"id1", "id2"}, Attributes: []string{}}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1": pcommon.NewValueStr("idvalue1"),
		},
	}
	logger := zap.NewNop()

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Entity: testEntity})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)

	events, err := eventBuilder.collectEvents(attributes, []*config.EntityEvent{{Entity: testEntity.Entity}}, nil)
	require.NoError(t, err)

	// assert
	require.Len(t, events, 0)
}

func TestCollectEvents_WithEntity_SomeAttributesMissing(t *testing.T) {
	// arrange
	testEntity := config.Entity{Entity: "testEntityType", IDs: []string{"id1"}, Attributes: []string{"attr1", "attr2"}}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
		},
	}

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Entity: testEntity})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	events, err := eventBuilder.collectEvents(attributes, []*config.EntityEvent{{Type: testEntity.Entity, Action: EventUpdateAction}}, nil)
	require.NoError(t, err)
	require.Len(t, events, 1)
	entityEvent, ok := events[0].(Entity)
	require.True(t, ok)

	assert.Equal(t, testEntity.Entity, entityEvent.Type)
	assert.Equal(t, "update", entityEvent.Action)
	assert.Equal(t, 1, entityEvent.IDs.Len())
	assertAttributeIsPresent(t, entityEvent.IDs, "id1", "idvalue1")
	assert.Equal(t, 1, entityEvent.Attributes.Len())
	assertAttributeIsPresent(t, entityEvent.Attributes, "attr1", "attrvalue1")
}

func TestCollectEvents_WithRelationship_AttributesPresent(t *testing.T) {
	// arrange
	srcEntity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id1"}, Attributes: []string{"attr1"}}
	destEntity := config.Entity{Entity: "KubernetesNamespace", IDs: []string{"id2"}, Attributes: []string{"attr2"}}
	testRelationship := config.RelationshipEvent{Source: srcEntity.Entity, Destination: destEntity.Entity, Action: EventUpdateAction}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"id2":   pcommon.NewValueStr("idvalue2"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
			"attr2": pcommon.NewValueStr("attrvalue2"),
		},
	}
	logger := zap.NewNop()

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Entity:  srcEntity,
		destEntity.Entity: destEntity,
	})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)

	events, err := eventBuilder.collectEvents(
		attributes,
		nil,
		[]*config.RelationshipEvent{&testRelationship},
	)
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

func TestCollectEvents_WithRelationship_SameType_AttributesPresent(t *testing.T) {
	// arrange
	entity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id"}, Attributes: []string{"attr"}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster", Action: EventUpdateAction}
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
	logger := zap.NewNop()

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)

	events, err := eventBuilder.collectEvents(
		attributes,
		nil,
		[]*config.RelationshipEvent{&testRelationship},
	)
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

func TestCollectEvents_WithRelationship_IDAttributesMissing(t *testing.T) {
	// arrange
	srcEntity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id1"}, Attributes: []string{}}
	destEntity := config.Entity{Entity: "KubernetesNamespace", IDs: []string{"id2"}, Attributes: []string{}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesNamespace"}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1": pcommon.NewValueStr("idvalue1"),
		},
	}
	logger := zap.NewNop()
	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Entity:  srcEntity,
		destEntity.Entity: destEntity,
	})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)

	events, err := eventBuilder.collectEvents(
		attributes,
		nil,
		[]*config.RelationshipEvent{{Type: testRelationship.Type}},
	)
	require.NoError(t, err)

	// assert
	require.Len(t, events, 0)
}

func TestCollectEvents_WithRelationship_SameType_IDAttributesMissing(t *testing.T) {
	// arrange
	entity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id"}, Attributes: []string{}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster"}
	attributes := Attributes{
		Source: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue1"),
		},
	}
	logger := zap.NewNop()
	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster": entity,
	})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)
	events, err := eventBuilder.collectEvents(
		attributes,
		nil,
		[]*config.RelationshipEvent{{Type: testRelationship.Type}},
	)
	require.NoError(t, err)

	// assert
	require.Len(t, events, 0)
}

func TestCollectEvents_WithRelationship_RelationshipAttributesPresent(t *testing.T) {
	// arrange
	srcEntity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id1"}}
	destEntity := config.Entity{Entity: "KubernetesNamespace", IDs: []string{"id2"}}
	testRelationship := config.RelationshipEvent{
		Source:      "KubernetesCluster",
		Destination: "KubernetesNamespace",
		Attributes:  []string{"relationshipAttr"},
		Action:      EventUpdateAction,
	}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":              pcommon.NewValueStr("idvalue1"),
			"id2":              pcommon.NewValueStr("idvalue2"),
			"relationshipAttr": pcommon.NewValueStr("relationshipValue"),
		},
	}
	logger := zap.NewNop()

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster":   srcEntity,
		"KubernetesNamespace": destEntity,
	})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)
	events, err := eventBuilder.collectEvents(
		attributes,
		nil,
		[]*config.RelationshipEvent{&testRelationship},
	)
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

func TestCollectEvents_WithRelationship_SameType_RelationshipAttributesPresent(t *testing.T) {
	// arrange
	entity := config.Entity{Entity: "KubernetesCluster", IDs: []string{"id"}}
	testRelationship := config.RelationshipEvent{
		Source:      "KubernetesCluster",
		Destination: "KubernetesCluster",
		Attributes:  []string{"relationshipAttr"},
		Action:      EventUpdateAction,
		Type:        "Has",
	}

	attributes := Attributes{
		Source: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue1"),
		},
		Destination: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue2"),
		},
		Common: map[string]pcommon.Value{
			"relationshipAttr": pcommon.NewValueStr("relationshipValue"),
		},
	}
	logger := zap.NewNop()

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Entity: entity})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)
	events, err := eventBuilder.collectEvents(
		attributes,
		nil,
		[]*config.RelationshipEvent{&testRelationship},
	)
	require.NoError(t, err)
	require.Len(t, events, 1)
	r1, ok := events[0].(*Relationship)
	require.True(t, ok)
	assert.Equal(t, "KubernetesCluster", r1.Source.Type)
	assertAttributeIsPresent(t, r1.Source.IDs, "id", "idvalue1")
	assert.Equal(t, "KubernetesCluster", r1.Destination.Type)
	assertAttributeIsPresent(t, r1.Destination.IDs, "id", "idvalue2")
	assert.Equal(t, "update", r1.Action)
	assert.Equal(t, "Has", r1.Type)
}

func assertAttributeIsPresent(t *testing.T, attrs pcommon.Map, key string, expected string) {
	if val, ok := attrs.Get(key); ok {
		assert.Equal(t, true, ok)
		assert.Equal(t, expected, val.Str())
	}
}
