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
		Type:       "Entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.ParsedEntityEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Type: entity.Type, Action: EventUpdateAction},
		ConditionSeq: seq,
	}
	relationshipEvent := config.ParsedRelationshipEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: relationshipType, Source: entity.Type, Destination: entity.Type, Action: EventUpdateAction},
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

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Type: entity})

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
	require.Len(t, events, 2)
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
		Type:       "Entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.ParsedEntityEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Type: entity.Type},
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

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Type: entity})

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
		Type:       "Entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.ParsedEntityEvent[ottlmetric.TransformContext]{
		Definition:   &config.EntityEvent{Type: entity.Type, Action: EventUpdateAction},
		ConditionSeq: seq,
	}
	relationshipEvent := config.ParsedRelationshipEvent[ottlmetric.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: relationshipType, Source: entity.Type, Destination: entity.Type, Action: EventUpdateAction},
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

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Type: entity})

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
	require.Len(t, events, 2)
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
		Type:       "Entity",
		IDs:        []string{"id"},
		Attributes: []string{"attr"},
	}
	entityEvent := config.ParsedEntityEvent[ottlmetric.TransformContext]{
		Definition:   &config.EntityEvent{Type: entity.Type},
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

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Type: entity})

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
		Definition:   &config.EntityEvent{Type: "test-entity"},
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
		Definition:   &config.EntityEvent{Type: "test-entity"},
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

func TestCreateEntity(t *testing.T) {
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
		},
	}

	entity := config.Entity{
		Type:       "KubernetesCluster",
		IDs:        []string{"id1"},
		Attributes: []string{"attr1"},
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Type: entity})

	// Create the event builder with a new logs instance
	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	entityEvent, err := eventDetector.attributeMapper.getEntities("KubernetesCluster", attributes)
	assert.Nil(t, err)
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	for _, event := range entityEvent {
		// Update the log records with the entity event
		event.Update(logRecords)
	}
	logRecord := logRecords.At(0)
	assert.Equal(t, 4, logRecord.Attributes().Len())

	actualEntityEventType, _ := logRecord.Attributes().Get(entityEventType)
	assert.Equal(t, entityUpdateEventType, actualEntityEventType.Str())

	actualEntityType, _ := logRecord.Attributes().Get(entityType)
	assert.Equal(t, "KubernetesCluster", actualEntityType.Str())

	actualEntityIDs, _ := logRecord.Attributes().Get(entityIds)
	assert.Equal(t, 1, actualEntityIDs.Map().Len())
	actualEntityId, _ := actualEntityIDs.Map().Get("id1")
	assert.Equal(t, "idvalue1", actualEntityId.Str())

	actualEntityAttributes, _ := logRecord.Attributes().Get(entityAttributes)
	assert.Equal(t, 1, actualEntityAttributes.Map().Len())
	actualEntityAttr, _ := actualEntityAttributes.Map().Get("attr1")
	assert.Equal(t, "attrvalue1", actualEntityAttr.Str())
}

func TestCreateEntityWithNoAttributes(t *testing.T) {
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1": pcommon.NewValueStr("idvalue1"),
		},
	}

	entity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{},
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Type: entity})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	_, err := eventDetector.attributeMapper.getEntities("KubernetesCluster", attributes)
	assert.NotNil(t, err)
}

func TestCreateRelationshipEvent(t *testing.T) {
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"id2":   pcommon.NewValueStr("idvalue2"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
		},
	}

	srcEntity := config.Entity{
		Type:       "KubernetesCluster",
		IDs:        []string{"id1"},
		Attributes: []string{"attr1"},
	}

	destEntity := config.Entity{
		Type:       "KubernetesNode",
		IDs:        []string{"id2"},
		Attributes: []string{"attr2"},
	}

	relationship := config.RelationshipEvent{
		Type:        "MemberOf",
		Source:      srcEntity.Type,
		Destination: destEntity.Type,
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Type:  srcEntity,
		destEntity.Type: destEntity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	relationshipEvent, err := eventDetector.attributeMapper.getRelationship(&relationship, attributes)
	assert.Nil(t, err)
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	relationshipEvent.Update(logRecords)
	logRecord := logRecords.At(0)
	assert.Equal(t, 6, logRecord.Attributes().Len())

	actualEntityEventType, _ := logRecord.Attributes().Get(entityEventType)
	assert.Equal(t, relationshipUpdateEventType, actualEntityEventType.Str())

	actualRelationshipType, _ := logRecord.Attributes().Get(relationshipType)
	assert.Equal(t, "MemberOf", actualRelationshipType.Str())

	actualSrcEntityType, _ := logRecord.Attributes().Get(srcEntityType)
	assert.Equal(t, "KubernetesCluster", actualSrcEntityType.Str())

	actualDestEntityType, _ := logRecord.Attributes().Get(destEntityType)
	assert.Equal(t, "KubernetesNode", actualDestEntityType.Str())
}

func TestCreateRelationshipEventWithNoAttributes(t *testing.T) {
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1": pcommon.NewValueStr("idvalue1"),
			"id2": pcommon.NewValueStr("idvalue2"),
		},
	}

	srcEntity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{},
	}

	destEntity := config.Entity{
		Type: "KubernetesNode",
		IDs:  []string{},
	}

	relationship := config.RelationshipEvent{
		Type:        "MemberOf",
		Source:      srcEntity.Type,
		Destination: destEntity.Type,
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Type:  srcEntity,
		destEntity.Type: destEntity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	_, err := eventDetector.attributeMapper.getRelationship(&relationship, attributes)
	assert.NotNil(t, err)
}

func TestCreateRelationshipEventWithoutResourceAttributes(t *testing.T) {
	attributes := Attributes{}

	srcEntity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{"id1"},
	}

	destEntity := config.Entity{
		Type: "KubernetesNode",
		IDs:  []string{"id2"},
	}

	relationship := config.RelationshipEvent{
		Type:        "MemberOf",
		Source:      "KubernetesCluster",
		Destination: "KubernetesNode",
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster": srcEntity,
		"KubernetesNode":    destEntity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	_, err := eventDetector.attributeMapper.getRelationship(&relationship, attributes)
	assert.NotNil(t, err)
}

func TestCreateSameTypeRelationshipEvent(t *testing.T) {
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
		Type:       "KubernetesCluster",
		IDs:        []string{"id"},
		Attributes: []string{"attr1"},
	}

	relationship := config.RelationshipEvent{
		Type:        "VirtualizationTopologyConnection",
		Source:      "KubernetesCluster",
		Destination: "KubernetesCluster",
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster": entity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	relationshipEvent, err := eventDetector.attributeMapper.getRelationship(&relationship, attributes)
	assert.Nil(t, err)
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	relationshipEvent.Update(logRecords)
	logRecord := logRecords.At(0)
	assert.Equal(t, 6, logRecord.Attributes().Len())

	actualEntityEventType, _ := logRecord.Attributes().Get(entityEventType)
	assert.Equal(t, relationshipUpdateEventType, actualEntityEventType.Str())

	actualRelationshipType, _ := logRecord.Attributes().Get(relationshipType)
	assert.Equal(t, "VirtualizationTopologyConnection", actualRelationshipType.Str())

	actualSrcEntityType, _ := logRecord.Attributes().Get(srcEntityType)
	assert.Equal(t, "KubernetesCluster", actualSrcEntityType.Str())

	actualDestEntityType, _ := logRecord.Attributes().Get(destEntityType)
	assert.Equal(t, "KubernetesCluster", actualDestEntityType.Str())
}

func TestCreateSameTypeRelationshipEventWithNoAttributesSameType(t *testing.T) {
	attributes := Attributes{
		Source: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue1"),
		},
		Destination: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("idvalue2"),
		},
	}

	entity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{},
	}

	relationship := config.RelationshipEvent{
		Type:        "VirtualizationTopologyConnection",
		Source:      "KubernetesCluster",
		Destination: "KubernetesCluster",
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster": entity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	_, err := eventDetector.attributeMapper.getRelationship(&relationship, attributes)
	assert.NotNil(t, err)
}

func TestCreateSameTypeRelationshipEventWithoutResourceAttributes(t *testing.T) {
	attributes := Attributes{}

	entity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{"id"},
	}

	relationship := config.RelationshipEvent{
		Type:        "VirtualizationTopologyConnection",
		Source:      "KubernetesCluster",
		Destination: "KubernetesCluster",
	}

	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		"KubernetesCluster": entity,
	})

	eventDetector := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	_, err := eventDetector.attributeMapper.getRelationship(&relationship, attributes)
	assert.NotNil(t, err)
}

func TestCollectEventsWithEntitiesWhenAttributesArePresent(t *testing.T) {
	// arrange
	testEntity := config.Entity{Type: "testEntityType", IDs: []string{"id1", "id2"}, Attributes: []string{"attr1", "attr2"}}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"id2":   pcommon.NewValueStr("idvalue2"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
			"attr2": pcommon.NewValueStr("attrvalue2"),
		},
	}

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Type: testEntity})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	events, err := eventBuilder.collectEvents(attributes, []*config.EntityEvent{{Type: testEntity.Type}}, nil)
	require.NoError(t, err)
	require.Len(t, events, 1)
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	events[0].Update(logRecords)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logRecords.At(0)
	assertEntityType(t, actualLogRecord.Attributes(), testEntity.Type)
	assertEventType(t, actualLogRecord.Attributes(), entityUpdateEventType)

	ids := getMap(actualLogRecord.Attributes(), entityIds)
	assert.Equal(t, 2, ids.Len())
	assertAttributeIsPresent(t, ids, "id1", "idvalue1")
	assertAttributeIsPresent(t, ids, "id2", "idvalue2")

	attrs := getMap(actualLogRecord.Attributes(), entityAttributes)
	assert.Equal(t, 2, attrs.Len())
	assertAttributeIsPresent(t, attrs, "attr1", "attrvalue1")
	assertAttributeIsPresent(t, attrs, "attr2", "attrvalue2")
	assertOtelEventAsLogIsPresent(t, logs)
}

func TestDoesNotCollectEventsWithEntitiesWhenIDAttributeIsMissing(t *testing.T) {
	// arrange
	testEntity := config.Entity{Type: "testEntityType", IDs: []string{"id1", "id2"}, Attributes: []string{}}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1": pcommon.NewValueStr("idvalue1"),
		},
	}
	logger := zap.NewNop()

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Type: testEntity})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger,
	)

	events, err := eventBuilder.collectEvents(attributes, []*config.EntityEvent{{Type: testEntity.Type}}, nil)
	require.NoError(t, err)

	// assert
	require.Len(t, events, 0)
}

func TestCollectEventsWithEntitiesWhenAttributeIsMissing(t *testing.T) {
	// arrange
	testEntity := config.Entity{Type: "testEntityType", IDs: []string{"id1"}, Attributes: []string{"attr1", "attr2"}}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1":   pcommon.NewValueStr("idvalue1"),
			"attr1": pcommon.NewValueStr("attrvalue1"),
		},
	}

	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{testEntity.Type: testEntity})

	eventBuilder := NewEventDetector(
		attributeMapper,
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil,
	)

	events, err := eventBuilder.collectEvents(attributes, []*config.EntityEvent{{Type: testEntity.Type}}, nil)
	require.NoError(t, err)
	require.Len(t, events, 1)
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	events[0].Update(logRecords)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logRecords.At(0)
	assertEntityType(t, actualLogRecord.Attributes(), testEntity.Type)
	assertEventType(t, actualLogRecord.Attributes(), entityUpdateEventType)
	assertOtelEventAsLogIsPresent(t, logs)

	ids := getMap(actualLogRecord.Attributes(), entityIds)
	assert.Equal(t, 1, ids.Len())
	assertAttributeIsPresent(t, ids, "id1", "idvalue1")

	attrs := getMap(actualLogRecord.Attributes(), entityAttributes)
	assert.Equal(t, 1, attrs.Len())
	assertAttributeIsPresent(t, attrs, "attr1", "attrvalue1")
}

func TestCollectEventsWithRelationshipsWhenAttributesArePresent(t *testing.T) {
	// arrange
	srcEntity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id1"}, Attributes: []string{"attr1"}}
	destEntity := config.Entity{Type: "KubernetesNamespace", IDs: []string{"id2"}, Attributes: []string{"attr2"}}
	testRelationship := config.RelationshipEvent{Source: srcEntity.Type, Destination: destEntity.Type, Action: EventUpdateAction}
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
		srcEntity.Type:  srcEntity,
		destEntity.Type: destEntity,
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
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	events[0].Update(logRecords)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logRecords.At(0)
	assertEventType(t, actualLogRecord.Attributes(), relationshipUpdateEventType)
	assertOtelEventAsLogIsPresent(t, logs)

	srcIds := getMap(actualLogRecord.Attributes(), relationshipSrcEntityIds)
	assert.Equal(t, 1, srcIds.Len())
	assertAttributeIsPresent(t, srcIds, "id1", "idvalue1")
	assertAttributeIsPresent(t, srcIds, "attr1", "attrvalue1")
	assertRelationshipEntityType(t, actualLogRecord.Attributes(), srcEntity.Type, srcEntityType)

	destIds := getMap(actualLogRecord.Attributes(), relationshipDestEntityIds)
	assert.Equal(t, 1, destIds.Len())
	assertAttributeIsPresent(t, destIds, "id2", "idvalue2")
	assertAttributeIsPresent(t, destIds, "attr2", "attrvalue2")
	assertRelationshipEntityType(t, actualLogRecord.Attributes(), destEntity.Type, destEntityType)
}

func TestAppendSameTypeRelationshipUpdateEventWhenAttributesArePresent(t *testing.T) {
	// arrange
	entity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id"}, Attributes: []string{"attr"}}
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
	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Type: entity})

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
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	events[0].Update(logRecords)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logRecords.At(0)
	assertEventType(t, actualLogRecord.Attributes(), relationshipUpdateEventType)
	assertOtelEventAsLogIsPresent(t, logs)

	srcIds := getMap(actualLogRecord.Attributes(), relationshipSrcEntityIds)
	assert.Equal(t, 1, srcIds.Len())
	assertAttributeIsPresent(t, srcIds, "id", "idvalue1")
	assertAttributeIsPresent(t, srcIds, "attr", "attrvalue")
	assertRelationshipEntityType(t, actualLogRecord.Attributes(), entity.Type, srcEntityType)

	destIds := getMap(actualLogRecord.Attributes(), relationshipDestEntityIds)
	assert.Equal(t, 1, destIds.Len())
	assertAttributeIsPresent(t, destIds, "id", "idvalue2")
	assertAttributeIsPresent(t, destIds, "attr", "attrvalue")
	assertRelationshipEntityType(t, actualLogRecord.Attributes(), entity.Type, destEntityType)
}

func TestDoesNotcollectEventsWhenIDAttributeIsMissing(t *testing.T) {
	// arrange
	srcEntity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id1"}, Attributes: []string{}}
	destEntity := config.Entity{Type: "KubernetesNamespace", IDs: []string{"id2"}, Attributes: []string{}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesNamespace"}
	attributes := Attributes{
		Common: map[string]pcommon.Value{
			"id1": pcommon.NewValueStr("idvalue1"),
		},
	}
	logger := zap.NewNop()
	// act
	attributeMapper := NewAttributeMapper(map[string]config.Entity{
		srcEntity.Type:  srcEntity,
		destEntity.Type: destEntity,
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

func TestDoesNotAppendSameTypeRelationshipUpdateEventWhenIDAttributeIsMissing(t *testing.T) {
	// arrange
	entity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id"}, Attributes: []string{}}
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

func TestCollectEventsWithRelationshipAttribute(t *testing.T) {
	// arrange
	srcEntity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id1"}}
	destEntity := config.Entity{Type: "KubernetesNamespace", IDs: []string{"id2"}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesNamespace", Attributes: []string{"relationshipAttr"}, Action: EventUpdateAction}
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
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	events[0].Update(logRecords)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logRecords.At(0)
	assertOtelEventAsLogIsPresent(t, logs)

	attrs := getMap(actualLogRecord.Attributes(), relationshipAttributes)
	assert.Equal(t, 1, attrs.Len())
	assertAttributeIsPresent(t, attrs, "relationshipAttr", "relationshipValue")
}

func TestAppendSameTypeRelationshipUpdateEventWithRelationshipAttribute(t *testing.T) {
	// arrange
	entity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id"}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster", Attributes: []string{"relationshipAttr"}, Action: EventUpdateAction}

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
	attributeMapper := NewAttributeMapper(map[string]config.Entity{entity.Type: entity})

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
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	events[0].Update(logRecords)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logRecords.At(0)
	assertOtelEventAsLogIsPresent(t, logs)

	attrs := getMap(actualLogRecord.Attributes(), relationshipAttributes)
	assert.Equal(t, 1, attrs.Len())
	assertAttributeIsPresent(t, attrs, "relationshipAttr", "relationshipValue")
}

func assertRelationshipEntityType(t *testing.T, attrs pcommon.Map, expected string, accessor string) {
	if val, ok := attrs.Get(accessor); ok {
		assert.Equal(t, true, ok)
		assert.Equal(t, expected, val.Str())
	}
}

func assertEventType(t *testing.T, attrs pcommon.Map, expected string) {
	if val, ok := attrs.Get(entityEventType); ok {
		assert.Equal(t, true, ok)
		assert.Equal(t, expected, val.Str())
	}
}

func assertEntityType(t *testing.T, attrs pcommon.Map, expected string) {
	if val, ok := attrs.Get(entityType); ok {
		assert.Equal(t, true, ok)
		assert.Equal(t, expected, val.Str())
	}
}

func assertAttributeIsPresent(t *testing.T, attrs pcommon.Map, key string, expected string) {
	if val, ok := attrs.Get(key); ok {
		assert.Equal(t, true, ok)
		assert.Equal(t, expected, val.Str())
	}
}

func assertOtelEventAsLogIsPresent(t *testing.T, logs plog.Logs) {
	isEntityEvent, ok := logs.ResourceLogs().At(0).ScopeLogs().At(0).Scope().Attributes().Get(entityEventAsLog)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, isEntityEvent.Bool())
}

func getMap(attrs pcommon.Map, key string) pcommon.Map {
	if val, ok := attrs.Get(key); ok {
		return val.Map()
	}
	return pcommon.Map{}
}
