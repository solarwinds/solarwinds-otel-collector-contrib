package models

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

const (
	entityEventType = "otel.entity.event.type"

	// Event type values
	entityUpdateEventType       = "entity_state"
	relationshipUpdateEventType = "entity_relationship_state"

	// Entity properties
	entityType       = "otel.entity.type"
	entityIdsKey     = "otel.entity.id"
	entityAttributes = "otel.entity.attributes"

	// Relationship properties
	relationshipSrcEntityIds  = "otel.entity_relationship.source_entity.id"
	relationshipDestEntityIds = "otel.entity_relationship.destination_entity.id"
	relationshipAttributes    = "otel.entity_relationship.attributes"
	relationshipType          = "otel.entity_relationship.type"
	srcEntityType             = "otel.entity_relationship.source_entity.type"
	destEntityType            = "otel.entity_relationship.destination_entity.type"
)

type Subject interface {
	Update(logRecord *plog.LogRecord)
}

type Action string

const (
	Update Action = "update"
)

type RelationshipEntity struct {
	Type string
	IDs  pcommon.Map
}

type Relationship struct {
	Action
	Type        string
	Source      RelationshipEntity
	Destination RelationshipEntity
	Attributes  pcommon.Map
}

type Entity struct {
	Action
	Type       string
	IDs        pcommon.Map
	Attributes pcommon.Map
}

func (r Relationship) Update(logRecord *plog.LogRecord) {
	attrs := logRecord.Attributes()
	attrs.PutStr(entityEventType, relationshipUpdateEventType)

	// Copy id, entity and relationship attributes as pcommon.Map to the log record
	srcIds := attrs.PutEmptyMap(relationshipSrcEntityIds)
	destIds := attrs.PutEmptyMap(relationshipDestEntityIds)
	relationshipAttrs := attrs.PutEmptyMap(relationshipAttributes)

	r.Source.IDs.CopyTo(srcIds)
	r.Destination.IDs.CopyTo(destIds)
	r.Attributes.CopyTo(relationshipAttrs)

	attrs.PutStr(relationshipType, r.Type)
	attrs.PutStr(srcEntityType, r.Source.Type)
	attrs.PutStr(destEntityType, r.Destination.Type)
}

func (e Entity) Update(logRecord *plog.LogRecord) {
	attrs := logRecord.Attributes()
	attrs.PutStr(entityEventType, entityUpdateEventType)

	// Copy id and entity attributes as pcommon.Map to the log record
	entityIds := attrs.PutEmptyMap(entityIdsKey)
	entityAttrs := attrs.PutEmptyMap(entityAttributes)

	e.IDs.CopyTo(entityIds)
	e.Attributes.CopyTo(entityAttrs)
}

var _ Subject = (*Relationship)(nil)
var _ Subject = (*Entity)(nil)
