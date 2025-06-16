package internal

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"time"
)

type Entity struct {
	Action     string
	Type       string
	IDs        pcommon.Map
	Attributes pcommon.Map
}

func (e Entity) Delete(logRecords *plog.LogRecordSlice) {
	logRecord := logRecords.AppendEmpty()
	logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	attrs := logRecord.Attributes()
	attrs.PutStr(entityEventType, entityDeleteEventType)
	attrs.PutStr(entityDeleteTypeAttrName, entitySoftDelete)
	attrs.PutStr(entityType, e.Type)

	// Copy id and entity attributes as pcommon.Map to the log record
	entityIdsMap := attrs.PutEmptyMap(entityIds)
	entityAttrs := attrs.PutEmptyMap(entityAttributes)

	e.IDs.CopyTo(entityIdsMap)
	e.Attributes.CopyTo(entityAttrs)
}

// Update adds a log record for the entity update event.
// Log record is decorated by following attributes:
//   - Entity type
//   - Entity ids
//   - Entity attributes
//   - timestamp
//   - event type set to update action
func (e Entity) Update(logRecords *plog.LogRecordSlice) {
	logRecord := logRecords.AppendEmpty()
	logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	attrs := logRecord.Attributes()
	attrs.PutStr(entityEventType, entityUpdateEventType)
	attrs.PutStr(entityType, e.Type)

	// Copy id and entity attributes as pcommon.Map to the log record
	entityIdsMap := attrs.PutEmptyMap(entityIds)
	entityAttrs := attrs.PutEmptyMap(entityAttributes)

	e.IDs.CopyTo(entityIdsMap)
	e.Attributes.CopyTo(entityAttrs)
}

var _ Event = (*Entity)(nil)

func (e Entity) GetActionType() string {
	return e.Action
}
