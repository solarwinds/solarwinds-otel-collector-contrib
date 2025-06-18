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
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestConditionTrue(t *testing.T) {
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
	_, _, err = processEvents(ctx, eventsGroup, tc)
	require.NoError(t, err)

}

func TestConditionFalse(t *testing.T) {
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
	_, _, err = processEvents(ctx, eventsGroup, tc)
	require.NoError(t, err)
}

func TestCreateEntity(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("attr1", "attrvalue1")

	entity := config.Entity{
		Type:       "KubernetesCluster",
		IDs:        []string{"id1"},
		Attributes: []string{"attr1"},
	}

	// Create the event builder with a new logs instance
	eventDetector := NewEventDetector(
		map[string]config.Entity{entity.Type: entity},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)

	entityEvent, err := eventDetector.createEntity(resourceAttrs, &config.EntityEvent{Type: entity.Type})
	assert.Nil(t, err)
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	entityEvent.Update(logRecords)
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
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")

	entity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{},
	}

	eventDetector := NewEventDetector(
		map[string]config.Entity{entity.Type: entity},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)
	_, err := eventDetector.createEntity(resourceAttrs, &config.EntityEvent{Type: entity.Type})
	assert.NotNil(t, err)
}

func TestCreateRelationshipEvent(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")
	resourceAttrs.PutStr("attr1", "attrvalue1")

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

	eventDetector := NewEventDetector(
		map[string]config.Entity{
			srcEntity.Type:  srcEntity,
			destEntity.Type: destEntity,
		},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)

	relationshipEvent, err := eventDetector.createRelationship(resourceAttrs, &relationship)
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
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")

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

	eventDetector := NewEventDetector(
		map[string]config.Entity{
			srcEntity.Type:  srcEntity,
			destEntity.Type: destEntity,
		},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)

	_, err := eventDetector.createRelationship(resourceAttrs, &relationship)
	assert.NotNil(t, err)
}

func TestCreateRelationshipEventWithoutResourceAttributes(t *testing.T) {
	resourceAttrs := pcommon.NewMap()

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

	eventDetector := NewEventDetector(
		map[string]config.Entity{
			"KubernetesCluster": srcEntity,
			"KubernetesNode":    destEntity,
		},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)

	_, err := eventDetector.createRelationship(resourceAttrs, &relationship)
	assert.NotNil(t, err)
}

func TestCreateSameTypeRelationshipEvent(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")
	resourceAttrs.PutStr("attr1", "attrvalue1")

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

	eventDetector := NewEventDetector(
		map[string]config.Entity{
			"KubernetesCluster": entity,
		},
		"src.",
		"dst.",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)

	relationshipEvent, err := eventDetector.createRelationship(resourceAttrs, &relationship)
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
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	entity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{},
	}

	relationship := config.RelationshipEvent{
		Type:        "VirtualizationTopologyConnection",
		Source:      "KubernetesCluster",
		Destination: "KubernetesCluster",
	}

	eventDetector := NewEventDetector(
		map[string]config.Entity{
			"KubernetesCluster": entity,
		},
		"src.",
		"dst.",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)

	_, err := eventDetector.createRelationship(resourceAttrs, &relationship)
	assert.NotNil(t, err)
}

func TestCreateSameTypeRelationshipEventWithoutResourceAttributes(t *testing.T) {
	resourceAttrs := pcommon.NewMap()

	entity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{"id"},
	}

	relationship := config.RelationshipEvent{
		Type:        "VirtualizationTopologyConnection",
		Source:      "KubernetesCluster",
		Destination: "KubernetesCluster",
	}

	eventDetector := NewEventDetector(
		map[string]config.Entity{
			"KubernetesCluster": entity,
		},
		"src.",
		"dst.",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)

	_, err := eventDetector.createRelationship(resourceAttrs, &relationship)
	assert.NotNil(t, err)
}

func TestCollectEventsWithEntitiesWhenAttributesArePresent(t *testing.T) {
	// arrange
	testEntity := config.Entity{Type: "testEntityType", IDs: []string{"id1", "id2"}, Attributes: []string{"attr1", "attr2"}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")
	resourceAttrs.PutStr("attr1", "attrvalue1")
	resourceAttrs.PutStr("attr2", "attrvalue2")

	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{testEntity.Type: testEntity},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)
	events, err := eventBuilder.collectEvents(resourceAttrs, []*config.EntityEvent{{Type: testEntity.Type}}, nil)
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
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	logger := zap.NewNop()

	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{testEntity.Type: testEntity},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger)

	events, err := eventBuilder.collectEvents(resourceAttrs, []*config.EntityEvent{{Type: testEntity.Type}}, nil)
	require.NoError(t, err)

	// assert
	require.Len(t, events, 0)
}

func TestCollectEventsWithEntitiesWhenAttributeIsMissing(t *testing.T) {
	// arrange
	testEntity := config.Entity{Type: "testEntityType", IDs: []string{"id1"}, Attributes: []string{"attr1", "attr2"}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("attr1", "attrvalue1")

	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{testEntity.Type: testEntity},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		nil)
	events, err := eventBuilder.collectEvents(resourceAttrs, []*config.EntityEvent{{Type: testEntity.Type}}, nil)
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
	testRelationship := config.RelationshipEvent{Source: srcEntity.Type, Destination: destEntity.Type}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")
	resourceAttrs.PutStr("attr1", "attrvalue1")
	resourceAttrs.PutStr("attr2", "attrvalue2")
	logger := zap.NewNop()

	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{
			srcEntity.Type:  srcEntity,
			destEntity.Type: destEntity,
		},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger)
	events, err := eventBuilder.collectEvents(
		resourceAttrs,
		nil,
		[]*config.RelationshipEvent{&testRelationship})
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
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster"}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")
	resourceAttrs.PutStr("attr", "attrvalue")
	logger := zap.NewNop()

	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{entity.Type: entity},
		"src.",
		"dst.",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger)

	events, err := eventBuilder.collectEvents(
		resourceAttrs,
		nil,
		[]*config.RelationshipEvent{&testRelationship})
	require.NoError(t, err)
	require.Len(t, events, 1)
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)
	events[0].Update(logRecords)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
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
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	logger := zap.NewNop()
	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{
			srcEntity.Type:  srcEntity,
			destEntity.Type: destEntity,
		},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger)

	events, err := eventBuilder.collectEvents(
		resourceAttrs,
		nil,
		[]*config.RelationshipEvent{{Type: testRelationship.Type}})
	require.NoError(t, err)

	// assert
	require.Len(t, events, 0)
}

func TestDoesNotAppendSameTypeRelationshipUpdateEventWhenIDAttributeIsMissing(t *testing.T) {
	// arrange
	entity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id"}, Attributes: []string{}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster"}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	logger := zap.NewNop()
	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{
			"KubernetesCluster": entity,
		},
		"src.",
		"dst.",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger)
	events, err := eventBuilder.collectEvents(
		resourceAttrs,
		nil,
		[]*config.RelationshipEvent{{Type: testRelationship.Type}})
	require.NoError(t, err)

	// assert
	require.Len(t, events, 0)
}

func TestCollectEventsWithRelationshipAttribute(t *testing.T) {
	// arrange
	srcEntity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id1"}}
	destEntity := config.Entity{Type: "KubernetesNamespace", IDs: []string{"id2"}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesNamespace", Attributes: []string{"relationshipAttr"}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")
	resourceAttrs.PutStr("relationshipAttr", "relationshipValue")
	logger := zap.NewNop()

	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{
			"KubernetesCluster":   srcEntity,
			"KubernetesNamespace": destEntity,
		},
		"",
		"",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger)
	events, err := eventBuilder.collectEvents(
		resourceAttrs,
		nil,
		[]*config.RelationshipEvent{&testRelationship})
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
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster", Attributes: []string{"relationshipAttr"}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")
	resourceAttrs.PutStr("relationshipAttr", "relationshipValue")
	logger := zap.NewNop()

	// act
	eventBuilder := NewEventDetector(
		map[string]config.Entity{entity.Type: entity},
		"src.",
		"dst.",
		config.EventsGroup[ottllog.TransformContext]{},
		config.EventsGroup[ottlmetric.TransformContext]{},
		logger)
	events, err := eventBuilder.collectEvents(
		resourceAttrs,
		nil,
		[]*config.RelationshipEvent{&testRelationship})
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
