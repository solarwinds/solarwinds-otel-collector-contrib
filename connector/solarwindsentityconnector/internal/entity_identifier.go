package internal

import (
	"fmt"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type EntityIdentifier struct {
	entities map[string]config.Entity
}

func (e *EntityIdentifier) getEntities(entityType string, attrs Attributes) (entities []Entity, err error) {
	entity, ok := e.entities[entityType]
	if !ok {
		return nil, fmt.Errorf("entity type %s not found", entityType)
	}

	if len(entity.IDs) == 0 {
		return nil, fmt.Errorf("entity type %s has no IDs configured", entityType)
	}

	if isSubset(entity.IDs, attrs.Source) {
		newEntity, created := createEntity(entity, attrs.Source, attrs.Common)
		if created {
			entities = append(entities, newEntity)
		}
	}

	if isSubset(entity.IDs, attrs.Destination) {
		newEntity, created := createEntity(entity, attrs.Destination, attrs.Common)
		if created {
			entities = append(entities, newEntity)
		}
	}

	if isSuperSet(entity.IDs, attrs.Common) {
		newEntity, created := createEntity(entity, attrs.Common)
		if created {
			entities = append(entities, newEntity)
		}
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("no entities found for entity type %s with attributes %v", entityType, attrs)
	}

	return entities, nil
}

func createEntity(entity config.Entity, attrs ...map[string]pcommon.Value) (Entity, bool) {
	ids, allAttributesFound := getEntityAttributes(entity.IDs, attrs...)
	if allAttributesFound {
		ea, _ := getEntityAttributes(entity.Attributes, attrs...)
		return Entity{
			IDs:        ids,
			Attributes: ea,
			Type:       entity.Type,
		}, true
	}

	return Entity{}, false
}

func isSubset(superset []string, subset map[string]pcommon.Value) bool {
	if subset == nil || len(subset) == 0 {
		return false
	}

	for key, _ := range subset {
		found := false
		for _, entityId := range superset {
			if entityId == key {
				found = true
				break
			}
		}
		if found == false {
			return false
		}
	}

	return true
}

func isSuperSet(subset []string, superset map[string]pcommon.Value) bool {
	for _, attribute := range subset {
		if _, exists := superset[attribute]; !exists {
			return false
		}
	}
	return true
}

func getEntityAttributes(configuredAttrs []string, actualAttrs ...map[string]pcommon.Value) (pcommon.Map, bool) {
	ids := pcommon.NewMap()
	for _, attrsMap := range actualAttrs {
		putAttributes(configuredAttrs, attrsMap, &ids)
	}
	return ids, true
}
