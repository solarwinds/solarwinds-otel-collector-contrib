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
)

func TestGetEntity_WithoutConfiguredIds(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"testEntityType": {Type: "testEntityType", IDs: []string{}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Common: map[string]pcommon.Value{"commonAttr": pcommon.NewValueStr("commonValue")},
	}

	_, err := attributeMapper.getEntity("testEntityType", attrs)
	assert.Error(t, err, "failed to get ID attributes for entity testEntityType: required attributes not configured")
}

func TestGetEntities_EntityNotFound(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"testEntityType": {Type: "testEntityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Common: map[string]pcommon.Value{"commonAttr": pcommon.NewValueStr("commonValue")},
	}

	entity, err := attributeMapper.getEntity("nonexistentEntityType", attrs)
	assert.Nil(t, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity type nonexistentEntityType not found")
}

func TestGetEntities_CommonContainsRequiredKeysEntityIsCreated(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"testEntityType": {Type: "testEntityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Common: map[string]pcommon.Value{
			"id":        pcommon.NewValueStr("id-value"),
			"extraAttr": pcommon.NewValueStr("extraValue")},
	}

	entity, err := attributeMapper.getEntity("testEntityType", attrs)
	assert.NoError(t, err)
	assert.Equal(t, "testEntityType", entity.Type)
	value, _ := entity.IDs.Get("id")
	assert.Equal(t, "id-value", value.Str())
}

func TestGetEntities_NoEntityCreatedWhenCommonIdIsMissing(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source:      map[string]pcommon.Value{"id": pcommon.NewValueStr("not-used")},
		Destination: map[string]pcommon.Value{"id": pcommon.NewValueStr("not-used")},
		Common:      map[string]pcommon.Value{"otherId": pcommon.NewValueStr("is-not-the-required-id")},
	}

	entity, err := attributeMapper.getEntity("entityType", attrs)
	assert.Nil(t, entity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create entity for type")
}

func TestGetRelationshipEntities_SourceEntityDoesNotExist(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"destType": {Type: "destType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Destination: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("destValue"),
		},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("nonExistentType", "destType", attrs)
	assert.Nil(t, src)
	assert.Nil(t, dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source entity type nonExistentType not found")
}

func TestGetRelationshipEntities_DestinationEntityDoesNotExist(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"sourceType": {Type: "sourceType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source: map[string]pcommon.Value{
			"id": pcommon.NewValueStr("sourceValue"),
		},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("sourceType", "nonExistentType", attrs)
	assert.Nil(t, src)
	assert.Nil(t, dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destination entity type nonExistentType not found")
}

func TestGetRelationshipEntities_SourceAndCommonAttributesOutputRelationshipEntities(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}

	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source: map[string]pcommon.Value{"id": pcommon.NewValueStr("sourceValue")},
		Common: map[string]pcommon.Value{"id": pcommon.NewValueStr("commonValue")},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("entityType", "entityType", attrs)
	assert.NoError(t, err)

	// Verify first entity from Source attributes
	assert.Equal(t, "entityType", src.Type)
	sourceValue, _ := src.IDs.Get("id")
	assert.Equal(t, "sourceValue", sourceValue.Str())

	// Verify second entity from Common attributes
	assert.Equal(t, "entityType", src.Type)
	commonValue, _ := dest.IDs.Get("id")
	assert.Equal(t, "commonValue", commonValue.Str())
}

func TestGetRelationshipEntities_DestinationAndCommonAttributesOutputRelationshipEntities(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Destination: map[string]pcommon.Value{"id": pcommon.NewValueStr("destinationValue")},
		Common:      map[string]pcommon.Value{"id": pcommon.NewValueStr("commonValue")},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("entityType", "entityType", attrs)
	assert.NoError(t, err)

	// Verify second entity from Common attributes
	assert.Equal(t, "entityType", src.Type)
	commonValue, _ := src.IDs.Get("id")
	assert.Equal(t, "commonValue", commonValue.Str())

	// Verify first entity from Destination attributes
	assert.Equal(t, "entityType", dest.Type)
	destinationValue, _ := dest.IDs.Get("id")
	assert.Equal(t, "destinationValue", destinationValue.Str())
}

func TestGetRelationshipEntities_SourceAndDestinationAttributesOutputRelationship(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"sourceType": {Type: "sourceType", IDs: []string{"id"}},
		"destType":   {Type: "destType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source:      map[string]pcommon.Value{"id": pcommon.NewValueStr("sourceValue")},
		Destination: map[string]pcommon.Value{"id": pcommon.NewValueStr("destValue")},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("sourceType", "destType", attrs)
	assert.NoError(t, err)

	// Verify source entity
	assert.Equal(t, "sourceType", src.Type)
	sourceValue, _ := src.IDs.Get("id")
	assert.Equal(t, "sourceValue", sourceValue.Str())

	// Verify destination entity
	assert.Equal(t, "destType", dest.Type)
	destValue, _ := dest.IDs.Get("id")
	assert.Equal(t, "destValue", destValue.Str())
}

func TestGetRelationshipEntities_AttributesSubsetUsedForRelationship(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"sourceType": {Type: "sourceType", IDs: []string{"id"}},
		"destType":   {Type: "destType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source:      map[string]pcommon.Value{"id": pcommon.NewValueStr("sourceValue")},
		Destination: map[string]pcommon.Value{"id": pcommon.NewValueStr("destValue")},
		Common: map[string]pcommon.Value{
			"attr1": pcommon.NewValueStr("value1"),
			"attr2": pcommon.NewValueStr("value2"),
		},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("sourceType", "destType", attrs)
	assert.NoError(t, err)

	// Verify source entity
	assert.Equal(t, "sourceType", src.Type)
	assert.Equal(t, 1, src.IDs.Len())
	sourceValue, _ := src.IDs.Get("id")
	assert.Equal(t, "sourceValue", sourceValue.Str())

	// Verify destination entity
	assert.Equal(t, "destType", dest.Type)
	assert.Equal(t, 1, dest.IDs.Len())
	destValue, _ := dest.IDs.Get("id")
	assert.Equal(t, "destValue", destValue.Str())
}

func TestGetRelationshipEntities_NoAttributesForSourceAndDestination(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"sourceType": {Type: "sourceType", IDs: []string{"id"}},
		"destType":   {Type: "destType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Common: map[string]pcommon.Value{"otherAttr": pcommon.NewValueStr("value")},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("sourceType", "destType", attrs)
	assert.Nil(t, src)
	assert.Nil(t, dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create source entity for type")
}

func TestGetRelationshipEntities_WithoutSourceIds(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"sourceType": {Type: "sourceType", IDs: []string{}},
		"destType":   {Type: "destType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Common: map[string]pcommon.Value{"otherAttr": pcommon.NewValueStr("value")},
	}

	_, _, err := attributeMapper.getRelationshipEntities("sourceType", "destType", attrs)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to create source entity for type")
}

func TestGetRelationshipEntities_WithoutDestIds(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"sourceType": {Type: "sourceType", IDs: []string{"id"}},
		"destType":   {Type: "destType", IDs: []string{}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source: map[string]pcommon.Value{"id": pcommon.NewValueStr("value")},
	}

	_, _, err := attributeMapper.getRelationshipEntities("sourceType", "destType", attrs)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to create destination entity for type")
}

func TestGetRelationshipEntities_SameTypeRelationshipWithoutPrefixes(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Common: map[string]pcommon.Value{"id": pcommon.NewValueStr("random")},
	}

	_, _, err := attributeMapper.getRelationshipEntities("entityType", "entityType", attrs)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "source and destination attributes are empty for same type relationship")
}

func TestGetRelationshipEntities_SameTypeRelationshipWithSourcePrefix(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source: map[string]pcommon.Value{"id": pcommon.NewValueStr("sourceValue")},
		Common: map[string]pcommon.Value{"id": pcommon.NewValueStr("commonValue")},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("entityType", "entityType", attrs)
	assert.Nil(t, err)
	assert.Equal(t, "entityType", src.Type)
	srcId, _ := src.IDs.Get("id")
	assert.Equal(t, "sourceValue", srcId.Str())

	assert.Equal(t, "entityType", dest.Type)
	destId, _ := dest.IDs.Get("id")
	assert.Equal(t, "commonValue", destId.Str())
}

func TestGetRelationshipEntities_SameTypeRelationshipWithDestinationPrefix(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Destination: map[string]pcommon.Value{"id": pcommon.NewValueStr("destValue")},
		Common:      map[string]pcommon.Value{"id": pcommon.NewValueStr("commonValue")},
	}

	src, dest, err := attributeMapper.getRelationshipEntities("entityType", "entityType", attrs)
	assert.Nil(t, err)

	assert.Equal(t, "entityType", src.Type)
	srcId, _ := src.IDs.Get("id")
	assert.Equal(t, "commonValue", srcId.Str())

	assert.Equal(t, "entityType", dest.Type)
	destId, _ := dest.IDs.Get("id")
	assert.Equal(t, "destValue", destId.Str())
}
