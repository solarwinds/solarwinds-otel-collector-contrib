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

func TestGetEntities_SourceEntityNotFound(t *testing.T) {
	attributeMapper := NewAttributeMapper(map[string]config.Entity{})
	attrs := Attributes{
		Source: map[string]pcommon.Value{"id": pcommon.NewValueStr("value")},
		Common: map[string]pcommon.Value{"commonAttr": pcommon.NewValueStr("commonValue")},
	}

	entities, err := attributeMapper.getEntities("nonexistentEntityType", attrs)
	assert.Nil(t, entities)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity type nonexistentEntityType not found")
}

func TestGetEntities_DestinationEntityNotFound(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"sourceEntityType": {Type: "sourceEntityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Destination: map[string]pcommon.Value{"id": pcommon.NewValueStr("value")},
		Common:      map[string]pcommon.Value{"commonAttr": pcommon.NewValueStr("commonValue")},
	}

	entities, err := attributeMapper.getEntities("nonexistentEntityType", attrs)
	assert.Nil(t, entities)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity type nonexistentEntityType not found")
}

func TestGetEntities_SourceAttributesSubsetOfEntityIDs(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"sourceEntityType": {Type: "sourceEntityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source: map[string]pcommon.Value{"id": pcommon.NewValueStr("value")},
		Common: map[string]pcommon.Value{"commonAttr": pcommon.NewValueStr("commonValue")},
	}

	entities, err := attributeMapper.getEntities("sourceEntityType", attrs)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "sourceEntityType", entities[0].Type)
	value, _ := entities[0].IDs.Get("id")
	assert.Equal(t, "value", value.Str())
}

func TestGetEntities_DestinationAttributesSubsetOfEntityIDs(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"destEntityType": {Type: "destEntityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Destination: map[string]pcommon.Value{"id": pcommon.NewValueStr("value")},
		Common:      map[string]pcommon.Value{"commonAttr": pcommon.NewValueStr("commonValue")},
	}

	entities, err := attributeMapper.getEntities("destEntityType", attrs)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "destEntityType", entities[0].Type)
	value, _ := entities[0].IDs.Get("id")
	assert.Equal(t, "value", value.Str())
}

func TestGetEntities_AttributesInCommonAreSuperset(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"commonEntityType": {Type: "commonEntityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Common: map[string]pcommon.Value{"id": pcommon.NewValueStr("value"), "extraAttr": pcommon.NewValueStr("extraValue")},
	}

	entities, err := attributeMapper.getEntities("commonEntityType", attrs)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "commonEntityType", entities[0].Type)
	value, _ := entities[0].IDs.Get("id")
	assert.Equal(t, "value", value.Str())
}

func TestGetEntities_NoEntityCreated(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source:      map[string]pcommon.Value{"otherId": pcommon.NewValueStr("value")},
		Destination: map[string]pcommon.Value{"otherId": pcommon.NewValueStr("value")},
		Common:      map[string]pcommon.Value{"otherId": pcommon.NewValueStr("value")},
	}

	entities, err := attributeMapper.getEntities("entityType", attrs)
	assert.Nil(t, entities)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no entityConfigs found for entity type entityType")
}

func TestGetEntities_SourceAndCommonAttributesOutputTwoEntities(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Source: map[string]pcommon.Value{"id": pcommon.NewValueStr("sourceValue")},
		Common: map[string]pcommon.Value{"id": pcommon.NewValueStr("commonValue")},
	}

	entities, err := attributeMapper.getEntities("entityType", attrs)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entities))

	// Verify first entity from Source attributes
	assert.Equal(t, "entityType", entities[0].Type)
	sourceValue, _ := entities[0].IDs.Get("id")
	assert.Equal(t, "sourceValue", sourceValue.Str())

	// Verify second entity from Common attributes
	assert.Equal(t, "entityType", entities[1].Type)
	commonValue, _ := entities[1].IDs.Get("id")
	assert.Equal(t, "commonValue", commonValue.Str())
}

func TestGetEntities_DestinationAndCommonAttributesOutputTwoEntities(t *testing.T) {
	entityConfigs := map[string]config.Entity{
		"entityType": {Type: "entityType", IDs: []string{"id"}},
	}
	attributeMapper := NewAttributeMapper(entityConfigs)
	attrs := Attributes{
		Destination: map[string]pcommon.Value{"id": pcommon.NewValueStr("destinationValue")},
		Common:      map[string]pcommon.Value{"id": pcommon.NewValueStr("commonValue")},
	}

	entities, err := attributeMapper.getEntities("entityType", attrs)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entities))

	// Verify first entity from Destination attributes
	assert.Equal(t, "entityType", entities[0].Type)
	destinationValue, _ := entities[0].IDs.Get("id")
	assert.Equal(t, "destinationValue", destinationValue.Str())

	// Verify second entity from Common attributes
	assert.Equal(t, "entityType", entities[1].Type)
	commonValue, _ := entities[1].IDs.Get("id")
	assert.Equal(t, "commonValue", commonValue.Str())
}
