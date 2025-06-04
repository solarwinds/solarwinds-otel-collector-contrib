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
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

func TestAppendEntityUpdateEventWhenAttributesArePresent(t *testing.T) {
	// arrange
	logs := plog.NewLogs()
	testEntity := config.Entity{Type: "testEntityType", IDs: []string{"id1", "id2"}, Attributes: []string{"attr1", "attr2"}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")
	resourceAttrs.PutStr("attr1", "attrvalue1")
	resourceAttrs.PutStr("attr2", "attrvalue2")

	// act
	eventBuilder := NewEventBuilder(nil, "", "", &logs, nil)
	eventBuilder.AppendEntityUpdateEvent(testEntity, resourceAttrs)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
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

func TestDoesNotAppendEntityUpdateEventWhenIDAttributeIsMissing(t *testing.T) {
	// arrange
	logs := plog.NewLogs()
	testEntity := config.Entity{Type: "testEntityType", IDs: []string{"id1", "id2"}, Attributes: []string{}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	logger := zap.NewNop()

	// act
	eventBuilder := NewEventBuilder(nil, "", "", &logs, logger)
	eventBuilder.AppendEntityUpdateEvent(testEntity, resourceAttrs)

	// assert
	assert.Equal(t, 0, logs.LogRecordCount())
}

func TestAppendEntityUpdateEventWhenAttributeIsMissing(t *testing.T) {
	// arrange
	logs := plog.NewLogs()
	testEntity := config.Entity{Type: "testEntityType", IDs: []string{"id1"}, Attributes: []string{"attr1", "attr2"}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("attr1", "attrvalue1")

	// act
	eventBuilder := NewEventBuilder(nil, "", "", &logs, nil)
	eventBuilder.AppendEntityUpdateEvent(testEntity, resourceAttrs)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
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

func TestAppendRelationshipUpdateEventWhenAttributesArePresent(t *testing.T) {
	// arrange
	logs := plog.NewLogs()
	srcEntity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id1"}, Attributes: []string{"attr1"}}
	destEntity := config.Entity{Type: "KubernetesNamespace", IDs: []string{"id2"}, Attributes: []string{"attr2"}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesNamespace"}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")
	resourceAttrs.PutStr("attr1", "attrvalue1")
	resourceAttrs.PutStr("attr2", "attrvalue2")

	// act
	eventBuilder := NewEventBuilder(map[string]config.Entity{
		"KubernetesCluster":   srcEntity,
		"KubernetesNamespace": destEntity,
	}, "", "", &logs, nil)
	eventBuilder.AppendRelationshipUpdateEvent(testRelationship, resourceAttrs)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
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
	logs := plog.NewLogs()
	entity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id"}, Attributes: []string{"attr"}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster"}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")
	resourceAttrs.PutStr("attr", "attrvalue")

	// act
	eventBuilder := NewEventBuilder(map[string]config.Entity{
		"KubernetesCluster": entity,
	}, "src.", "dst.", &logs, nil)
	eventBuilder.AppendRelationshipUpdateEvent(testRelationship, resourceAttrs)

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

func TestDoesNotAppendRelationshipUpdateEventWhenIDAttributeIsMissing(t *testing.T) {
	// arrange
	logs := plog.NewLogs()
	srcEntity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id1"}, Attributes: []string{}}
	destEntity := config.Entity{Type: "KubernetesNamespace", IDs: []string{"id2"}, Attributes: []string{}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesNamespace"}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	logger := zap.NewNop()
	// act
	eventBuilder := NewEventBuilder(map[string]config.Entity{
		"KubernetesCluster":   srcEntity,
		"KubernetesNamespace": destEntity,
	}, "", "", &logs, logger)
	eventBuilder.AppendRelationshipUpdateEvent(testRelationship, resourceAttrs)

	// assert
	assert.Equal(t, 0, logs.LogRecordCount())
}

func TestDoesNotAppendSameTypeRelationshipUpdateEventWhenIDAttributeIsMissing(t *testing.T) {
	// arrange
	logs := plog.NewLogs()
	entity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id"}, Attributes: []string{}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster"}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	logger := zap.NewNop()
	// act
	eventBuilder := NewEventBuilder(map[string]config.Entity{
		"KubernetesCluster": entity,
	}, "src.", "dst.", &logs, logger)
	eventBuilder.AppendRelationshipUpdateEvent(testRelationship, resourceAttrs)

	// assert
	assert.Equal(t, 0, logs.LogRecordCount())
}

func TestAppendRelationshipUpdateEventWithRelationshipAttribute(t *testing.T) {
	// arrange
	logs := plog.NewLogs()
	srcEntity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id1"}}
	destEntity := config.Entity{Type: "KubernetesNamespace", IDs: []string{"id2"}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesNamespace", Attributes: []string{"relationshipAttr"}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")
	resourceAttrs.PutStr("relationshipAttr", "relationshipValue")

	// act
	eventBuilder := NewEventBuilder(map[string]config.Entity{
		"KubernetesCluster":   srcEntity,
		"KubernetesNamespace": destEntity,
	}, "", "", &logs, nil)
	eventBuilder.AppendRelationshipUpdateEvent(testRelationship, resourceAttrs)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	assertOtelEventAsLogIsPresent(t, logs)

	attrs := getMap(actualLogRecord.Attributes(), relationshipAttributes)
	assert.Equal(t, 1, attrs.Len())
	assertAttributeIsPresent(t, attrs, "relationshipAttr", "relationshipValue")
}

func TestAppendSameTypeRelationshipUpdateEventWithRelationshipAttribute(t *testing.T) {
	// arrange
	logs := plog.NewLogs()
	entity := config.Entity{Type: "KubernetesCluster", IDs: []string{"id"}}
	testRelationship := config.RelationshipEvent{Source: "KubernetesCluster", Destination: "KubernetesCluster", Attributes: []string{"relationshipAttr"}}
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")
	resourceAttrs.PutStr("relationshipAttr", "relationshipValue")

	// act
	eventBuilder := NewEventBuilder(map[string]config.Entity{
		"KubernetesCluster": entity,
	}, "src.", "dst.", &logs, nil)
	eventBuilder.AppendRelationshipUpdateEvent(testRelationship, resourceAttrs)

	// assert
	assert.Equal(t, 1, logs.LogRecordCount())
	actualLogRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
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
