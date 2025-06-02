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
)

func TestSetIdAttributesDefaultEmpty(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")

	destination := plog.NewLogRecord()
	err := setIdAttributes(destination.Attributes(), []string{}, resourceAttrs, entityIds)
	assert.NotNil(t, err)
}

func TestSetIdAttributesSameTypeEmpty(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	destination := plog.NewLogRecord()
	hasPrefix, err := setIdAttributesWithPrefix(destination.Attributes(), []string{}, resourceAttrs, entityIds, "src.")
	assert.NotNil(t, err)
	assert.False(t, hasPrefix)
}

func TestSetIdAttributesDefaultNoMatch(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")

	destination := plog.NewLogRecord()
	err := setIdAttributes(destination.Attributes(), []string{"id3"}, resourceAttrs, entityIds)
	assert.NotNil(t, err)
	ids, exists := destination.Attributes().Get(entityIds)
	assert.True(t, exists)
	assert.Equal(t, 0, ids.Map().Len())
}

func TestSetIdAttributesSameTypetIdNoMatch(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	destination := plog.NewLogRecord()
	hasPrefix, err := setIdAttributesWithPrefix(destination.Attributes(), []string{"id2"}, resourceAttrs, entityIds, "src.")
	assert.False(t, hasPrefix)
	assert.NotNil(t, err)
	ids, exists := destination.Attributes().Get(entityIds)
	assert.True(t, exists)
	assert.Equal(t, 0, ids.Map().Len())
}

func TestSetIdAttributesSameTypePrefixNoMatch(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	destination := plog.NewLogRecord()
	hasPrefix, err := setIdAttributesWithPrefix(destination.Attributes(), []string{"id"}, resourceAttrs, entityIds, "prefix.")
	assert.False(t, hasPrefix)
	assert.NotNil(t, err)
	ids, exists := destination.Attributes().Get(entityIds)
	assert.True(t, exists)
	assert.Equal(t, 0, ids.Map().Len())
}

func TestSetIdAttributesDefaultMultiple(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")

	destination := plog.NewLogRecord()
	err := setIdAttributes(destination.Attributes(), []string{"id1"}, resourceAttrs, entityIds)
	assert.Nil(t, err)
	ids, exists := destination.Attributes().Get(entityIds)
	assert.True(t, exists)
	assert.Equal(t, 1, ids.Map().Len())
	id, exists := ids.Map().Get("id1")
	assert.True(t, exists)
	assert.Equal(t, "idvalue1", id.Str())
}

func TestSetIdAttributesSameTypeMultiple(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	destination := plog.NewLogRecord()
	hasPrefix, err := setIdAttributesWithPrefix(destination.Attributes(), []string{"id"}, resourceAttrs, entityIds, "src.")
	assert.True(t, hasPrefix)
	assert.Nil(t, err)
	ids, exists := destination.Attributes().Get(entityIds)
	assert.True(t, exists)
	assert.Equal(t, 1, ids.Map().Len())
	id, exists := ids.Map().Get("id")
	assert.True(t, exists)
	assert.Equal(t, "idvalue1", id.Str())
}

func TestSetAttributesSingle(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("attr1", "attrvalue1")
	resourceAttrs.PutStr("attr2", "attrvalue2")

	destination := plog.NewLogRecord()
	setAttributes(destination.Attributes(), []string{"attr1"}, resourceAttrs, entityAttributes)
	attrs, exists := destination.Attributes().Get(entityAttributes)
	assert.True(t, exists)
	assert.Equal(t, 1, attrs.Map().Len())
}

func TestSetAttributesEmpty(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("attr1", "attrvalue1")
	resourceAttrs.PutStr("attr2", "attrvalue2")

	destination := plog.NewLogRecord()
	setAttributes(destination.Attributes(), []string{}, resourceAttrs, entityAttributes)
	_, exists := destination.Attributes().Get(entityAttributes)
	assert.False(t, exists)
}

func TestPutStringAttribute(t *testing.T) {
	value := pcommon.NewValueStr("stringValue")

	destination := pcommon.NewMap()
	putAttribute(&destination, "string", value)
	inDest, exists := destination.Get("string")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, "stringValue", inDest.Str(), "Attribute value should match")
}

func TestPutBoolAttribute(t *testing.T) {
	value := pcommon.NewValueBool(true)

	destination := pcommon.NewMap()
	putAttribute(&destination, "bool", value)
	inDest, exists := destination.Get("bool")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, true, inDest.Bool(), "Attribute value should match")
}

func TestPutIntAttribute(t *testing.T) {
	value := pcommon.NewValueInt(123)
	destination := pcommon.NewMap()
	putAttribute(&destination, "int", value)
	inDest, exists := destination.Get("int")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, int64(123), inDest.Int(), "Attribute value should match")
}

func TestPutDoubleAttribute(t *testing.T) {
	value := pcommon.NewValueDouble(123.456)

	destination := pcommon.NewMap()
	putAttribute(&destination, "double", value)
	inDest, exists := destination.Get("double")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, 123.456, inDest.Double(), "Attribute value should match")
}

func TestPutBytesAttribute(t *testing.T) {
	value := pcommon.NewValueBytes()
	value.SetEmptyBytes().Append('1', '2')

	destination := pcommon.NewMap()
	putAttribute(&destination, "bytes", value)
	inDest, exists := destination.Get("bytes")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, "12", string(inDest.Bytes().AsRaw()), "Attribute value should match")
}

func TestCreateEntityEvent(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("attr1", "attrvalue1")

	entity := config.Entity{
		Type:       "KubernetesCluster",
		IDs:        []string{"id1"},
		Attributes: []string{"attr1"},
	}

	// Create the event builder with a new logs instance
	logs := plog.NewLogs()
	eventBuilder := NewEventBuilder(nil, "", "", &logs, nil)

	logRecord, err := eventBuilder.createEntityEvent(resourceAttrs, entity)
	assert.Nil(t, err)
	assert.Equal(t, 3, logRecord.Attributes().Len())

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

func TestCreateEntityEventWithNoAttributes(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")

	entity := config.Entity{
		Type: "KubernetesCluster",
		IDs:  []string{},
	}

	logs := plog.NewLogs()
	eventBuilder := NewEventBuilder(nil, "", "", &logs, nil)
	_, err := eventBuilder.createEntityEvent(resourceAttrs, entity)
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
		Source:      "KubernetesCluster",
		Destination: "KubernetesNode",
	}

	logs := plog.NewLogs()
	eventBuilder := NewEventBuilder(
		map[string]config.Entity{
			"KubernetesCluster": srcEntity,
			"KubernetesNode":    destEntity,
		},
		"",
		"",
		&logs,
		nil,
	)

	logRecord, err := eventBuilder.createRelationshipEvent(relationship, resourceAttrs)
	assert.Nil(t, err)
	assert.Equal(t, 5, logRecord.Attributes().Len())

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
		Source:      "KubernetesCluster",
		Destination: "KubernetesNode",
	}

	logs := plog.NewLogs()
	eventBuilder := NewEventBuilder(
		map[string]config.Entity{
			"KubernetesCluster": srcEntity,
			"KubernetesNode":    destEntity,
		},
		"",
		"",
		&logs,
		nil,
	)

	_, err := eventBuilder.createRelationshipEvent(relationship, resourceAttrs)
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

	logs := plog.NewLogs()
	eventBuilder := NewEventBuilder(
		map[string]config.Entity{
			"KubernetesCluster": srcEntity,
			"KubernetesNode":    destEntity,
		},
		"",
		"",
		&logs,
		nil,
	)

	_, err := eventBuilder.createRelationshipEvent(relationship, resourceAttrs)
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

	logs := plog.NewLogs()
	eventBuilder := NewEventBuilder(
		map[string]config.Entity{
			"KubernetesCluster": entity,
		},
		"src.",
		"dst.",
		&logs,
		nil,
	)

	logRecord, err := eventBuilder.createRelationshipEvent(relationship, resourceAttrs)
	assert.Nil(t, err)
	assert.Equal(t, 5, logRecord.Attributes().Len())

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

	logs := plog.NewLogs()
	eventBuilder := NewEventBuilder(
		map[string]config.Entity{
			"KubernetesCluster": entity,
		},
		"src.",
		"dst.",
		&logs,
		nil,
	)

	_, err := eventBuilder.createRelationshipEvent(relationship, resourceAttrs)
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

	logs := plog.NewLogs()
	eventBuilder := NewEventBuilder(
		map[string]config.Entity{
			"KubernetesCluster": entity,
		},
		"src.",
		"dst.",
		&logs,
		nil,
	)

	_, err := eventBuilder.createRelationshipEvent(relationship, resourceAttrs)
	assert.NotNil(t, err)
}
