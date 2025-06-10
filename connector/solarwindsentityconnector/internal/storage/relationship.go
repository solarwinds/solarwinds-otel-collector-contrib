package storage

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type relationship struct {
	config.Relationship
	SourceEntityIDs      []string
	DestinationEntityIDs []string
}

func newRelationship(rEvent config.RelationshipEvent, sourceIds, destIds pcommon.Map) relationship {
	value := relationship{
		Relationship: config.Relationship{
			Type:        rEvent.Type,
			Source:      rEvent.Source,
			Destination: rEvent.Destination,
		},
		SourceEntityIDs:      make([]string, 0, sourceIds.Len()),
		DestinationEntityIDs: make([]string, 0, destIds.Len()),
	}

	sourceIds.Range(func(k string, v pcommon.Value) bool {
		value.SourceEntityIDs = append(value.SourceEntityIDs, k)
		return true
	})

	destIds.Range(func(k string, v pcommon.Value) bool {
		value.DestinationEntityIDs = append(value.DestinationEntityIDs, k)
		return true
	})

	return value
}
