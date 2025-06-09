package storage

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type relationship struct {
	RelationshipType      string
	SourceEntityType      string
	DestinationEntityType string
	SourceEntityID        []string
	DestinationEntityID   []string
}

func newRelationship(rEvent config.RelationshipEvent, sourceIds, destIds pcommon.Map) relationship {
	value := relationship{
		RelationshipType:      rEvent.Type,
		SourceEntityType:      rEvent.Source,
		DestinationEntityType: rEvent.Destination,
		SourceEntityID:        make([]string, 0, sourceIds.Len()),
		DestinationEntityID:   make([]string, 0, destIds.Len()),
	}

	sourceIds.Range(func(k string, v pcommon.Value) bool {
		value.SourceEntityID = append(value.SourceEntityID, k)
		return true
	})

	destIds.Range(func(k string, v pcommon.Value) bool {
		value.DestinationEntityID = append(value.DestinationEntityID, k)
		return true
	})

	return value
}
