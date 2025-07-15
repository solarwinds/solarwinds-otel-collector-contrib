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
	"fmt"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type AttributeMapper struct {
	entityConfigs map[string]config.Entity
}

func NewAttributeMapper(entityConfigs map[string]config.Entity) AttributeMapper {
	return AttributeMapper{
		entityConfigs: entityConfigs,
	}
}

func (e *AttributeMapper) getEntity(entityType string, attrs Attributes) (*Entity, error) {
	entity, ok := e.entityConfigs[entityType]
	if !ok {
		return nil, fmt.Errorf("entity type %s not found", entityType)
	}

	newEntity, err := createEntity(entity, attrs.Common)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity for type %s: %w", entityType, err)
	}

	return &newEntity, nil
}

func (e *AttributeMapper) getRelationshipEntities(sourceEntityType, destEntityType string, attrs Attributes) (*Entity, *Entity, error) {
	if sourceEntityType == destEntityType {
		if len(attrs.Source) == 0 && len(attrs.Destination) == 0 {
			return nil, nil, fmt.Errorf("source and destination attributes are empty for same type relationship")
		}
	}

	sourceEntityConfiguration, ok := e.entityConfigs[sourceEntityType]
	if !ok {
		return nil, nil, fmt.Errorf("source entity type %s not found", sourceEntityType)
	}

	sourceEntity, err := createEntity(sourceEntityConfiguration, attrs.Common, attrs.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create source entity for type %s: %w", sourceEntityType, err)
	}

	destinationEntityConfiguration, ok := e.entityConfigs[destEntityType]
	if !ok {
		return nil, nil, fmt.Errorf("destination entity type %s not found", destEntityType)
	}
	destinationEntity, err := createEntity(destinationEntityConfiguration, attrs.Common, attrs.Destination)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create destination entity for type %s: %w", destEntityType, err)
	}

	return &sourceEntity, &destinationEntity, nil
}

func createEntity(entity config.Entity, attrs ...map[string]pcommon.Value) (Entity, error) {
	ids, err := getRequiredAttributes(entity.IDs, attrs...)
	if err != nil {
		return Entity{}, fmt.Errorf("failed to get ID attributes for entity %s: %w", entity.Type, err)
	}

	ea := getOptionalAttributes(entity.Attributes, attrs...)

	return Entity{
		IDs:        ids,
		Attributes: ea,
		Type:       entity.Type,
	}, nil
}

func getRequiredAttributes(requiredKeys []string, actualAttributes ...map[string]pcommon.Value) (pcommon.Map, error) {
	if len(requiredKeys) == 0 {
		return pcommon.NewMap(), fmt.Errorf("required attributes not configured")
	}
	required := pcommon.NewMap()
	mergedAttributes := mergeMaps(actualAttributes...)

	for _, requiredKey := range requiredKeys {
		value, exists := mergedAttributes[requiredKey]
		if !exists {
			return pcommon.NewMap(), fmt.Errorf("required attribute %s not found in actual attributes", requiredKey)
		}
		putAttribute(&required, requiredKey, value)
	}
	return required, nil
}

func getOptionalAttributes(optionalKeys []string, actualAttributes ...map[string]pcommon.Value) pcommon.Map {
	optional := pcommon.NewMap()
	mergedAttributes := mergeMaps(actualAttributes...)

	for _, optionalKey := range optionalKeys {
		value, exists := mergedAttributes[optionalKey]
		if exists {
			putAttribute(&optional, optionalKey, value)
		}
	}
	return optional
}

func mergeMaps(maps ...map[string]pcommon.Value) map[string]pcommon.Value {
	merged := make(map[string]pcommon.Value)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}
