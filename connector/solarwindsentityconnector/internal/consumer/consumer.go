package consumer

import (
	"context"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	otelConsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"time"
)

type Relationship struct {
	config.Relationship
	SourceEntityIDs      []string
	DestinationEntityIDs []string
}

type Consumer interface {
	SendExpiredRelationships(ctx context.Context, relationships []Relationship)
}

const (
	// Log properties
	entityEventAsLog = "otel.entity.event_as_log"
	entityEventType  = "otel.entity.event.type"

	// Relationship properties
	relationshipSrcEntityIds  = "otel.entity_relationship.source_entity.id"
	relationshipDestEntityIds = "otel.entity_relationship.destination_entity.id"
	relationshipType          = "otel.entity_relationship.type"
	srcEntityType             = "otel.entity_relationship.source_entity.type"
	destEntityType            = "otel.entity_relationship.destination_entity.type"
)

type consumer struct {
	logsConsumer otelConsumer.Logs
	entities     map[string]config.Entity
}

var _ Consumer = (*consumer)(nil)

func NewConsumer(logsConsumer otelConsumer.Logs, entities map[string]config.Entity) Consumer {
	return &consumer{
		logsConsumer: logsConsumer,
		entities:     entities,
	}
}

func (c *consumer) SendExpiredRelationships(ctx context.Context, relationships []Relationship) {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr(entityEventAsLog, "true")
	sl := rl.ScopeLogs().AppendEmpty()

	timestamp := pcommon.NewTimestampFromTime(time.Now())

	for _, relationship := range relationships {
		lr := sl.LogRecords().AppendEmpty()
		lr.SetObservedTimestamp(timestamp)
		attrs := lr.Attributes()
		attrs.PutStr(relationshipType, relationship.Type)
		attrs.PutStr(srcEntityType, relationship.Source)
		attrs.PutStr(destEntityType, relationship.Destination)
		attrs.PutStr(entityEventType, "set_unknown")

		srcIds := attrs.PutEmptyMap(relationshipSrcEntityIds)
		srcEntity := c.entities[relationship.Source]
		for i := 0; i < len(relationship.SourceEntityIDs); i++ {
			srcIds.PutStr(srcEntity.IDs[i], relationship.SourceEntityIDs[i])
		}

		destIds := attrs.PutEmptyMap(relationshipDestEntityIds)
		destEntity := c.entities[relationship.Destination]
		for i := 0; i < len(relationship.DestinationEntityIDs); i++ {
			destIds.PutStr(destEntity.IDs[i], relationship.DestinationEntityIDs[i])
		}
	}

	_ = c.logsConsumer.ConsumeLogs(ctx, logs)

}
