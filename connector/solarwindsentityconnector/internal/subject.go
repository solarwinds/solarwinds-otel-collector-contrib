package internal

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"time"
)

type Subject interface {
	Update(logRecord *plog.LogRecord)
	Delete(logRecord *plog.LogRecordSlice)
}

type RelationshipEntity struct {
	Type string
	IDs  pcommon.Map
}

type Relationship struct {
	Type        string
	Source      RelationshipEntity
	Destination RelationshipEntity
	Attributes  pcommon.Map
}

func (r Relationship) Delete(logRecords *plog.LogRecordSlice) {
	logRecord := logRecords.AppendEmpty()
	logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	attrs := logRecord.Attributes()
	attrs.PutStr(entityEventType, relationshipDeleteEventType)

	// Copy id, entity and relationship attributes as pcommon.Map to the log record
	srcIds := attrs.PutEmptyMap(relationshipSrcEntityIds)
	destIds := attrs.PutEmptyMap(relationshipDestEntityIds)

	r.Source.IDs.CopyTo(srcIds)
	r.Destination.IDs.CopyTo(destIds)

	attrs.PutStr(relationshipType, r.Type)
	attrs.PutStr(srcEntityType, r.Source.Type)
	attrs.PutStr(destEntityType, r.Destination.Type)
}

type Entity struct {
	Type       string
	IDs        pcommon.Map
	Attributes pcommon.Map
}

func (e Entity) Delete(_ *plog.LogRecordSlice) {
	// TODO: Implement delete logic for Entity in the following task.
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
	attrs.PutStr(entityType, e.Type)

	// Copy id and entity attributes as pcommon.Map to the log record
	entityIdsMap := attrs.PutEmptyMap(entityIds)
	entityAttrs := attrs.PutEmptyMap(entityAttributes)

	e.IDs.CopyTo(entityIdsMap)
	e.Attributes.CopyTo(entityAttrs)
}

var _ Subject = (*Relationship)(nil)
var _ Subject = (*Entity)(nil)
