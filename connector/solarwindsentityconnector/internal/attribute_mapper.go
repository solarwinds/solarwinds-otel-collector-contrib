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
	"errors"
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

func (e *AttributeMapper) getEntities(entityType string, attrs Attributes) (entities []Entity, err error) {
	entity, ok := e.entityConfigs[entityType]
	if !ok {
		return nil, fmt.Errorf("entity type %s not found", entityType)
	}

	if len(entity.IDs) == 0 {
		return nil, fmt.Errorf("entity type %s has no IDs configured", entityType)
	}

	if attrs.Source.IsSubsetOf(entity.IDs) {
		newEntity, err := createEntity(entity, attrs.Common, attrs.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to create entity for type %s: %w", entityType, err)
		}
		entities = append(entities, newEntity)
	}

	if attrs.Destination.IsSubsetOf(entity.IDs) {
		newEntity, err := createEntity(entity, attrs.Common, attrs.Destination)
		if err != nil {
			return nil, fmt.Errorf("failed to create entity for type %s: %w", entityType, err)
		}
		entities = append(entities, newEntity)
	}

	if attrs.Common.ContainsAll(entity.IDs) {
		newEntity, err := createEntity(entity, attrs.Common)
		if err != nil {
			return nil, fmt.Errorf("failed to create entity for type %s: %w", entityType, err)
		}
		entities = append(entities, newEntity)
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("no entityConfigs found for entity type %s with attributes %v", entityType, attrs)
	}

	return entities, nil
}

func (e *AttributeMapper) getRelationship(relationship *config.RelationshipEvent, attrs Attributes) (*Relationship, error) {
	source, ok := e.entityConfigs[relationship.Source]
	if !ok {
		return nil, fmt.Errorf("unexpected source entity type")
	}

	dest, ok := e.entityConfigs[relationship.Destination]
	if !ok {
		return nil, fmt.Errorf("unexpected destination entity type")
	}

	if dest.Type == source.Type {
		if len(attrs.Source) == 0 && len(attrs.Destination) == 0 {
			return nil, fmt.Errorf("source and destination attributes are empty for same type relationship %s", relationship.Type)
		}
	}

	sourceIds, err := getRequiredAttributes(source.IDs, attrs.Common, attrs.Source)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to create source entity %s", source.Type), err)

	}

	destIds, err := getRequiredAttributes(dest.IDs, attrs.Common, attrs.Destination)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to create destination entity %s", dest.Type), err)
	}

	r := Relationship{
		Type: relationship.Type,
		Source: RelationshipEntity{
			Type: source.Type,
			IDs:  sourceIds,
		},
		Destination: RelationshipEntity{
			Type: dest.Type,
			IDs:  destIds,
		},
		Attributes: getOptionalAttributes(relationship.Attributes, attrs.Common),
	}

	return &r, nil
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

func getRequiredAttributes(configuredAttrs []string, actualAttrs ...map[string]pcommon.Value) (pcommon.Map, error) {
	if len(configuredAttrs) == 0 {
		return pcommon.NewMap(), fmt.Errorf("required attributes not configured")
	}
	required := pcommon.NewMap()
	allAttrs := mergeMaps(actualAttrs...)

	for _, requiredAttr := range configuredAttrs {
		value, exists := allAttrs[requiredAttr]
		if !exists {
			return pcommon.NewMap(), fmt.Errorf("required attribute %s not found in actual attributes", requiredAttr)
		}
		putAttribute(&required, requiredAttr, value)
	}
	return required, nil
}

func getOptionalAttributes(configuredAttrs []string, actualAttrs ...map[string]pcommon.Value) pcommon.Map {
	optional := pcommon.NewMap()
	allAttrs := mergeMaps(actualAttrs...)

	for _, optionalAttr := range configuredAttrs {
		value, exists := allAttrs[optionalAttr]
		if exists {
			putAttribute(&optional, optionalAttr, value)
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
