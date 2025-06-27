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

// Delete adds a log record for the entity delete event.
// Log record is decorated by following attributes:
//   - Entity type
//   - Entity ids
//   - Entity attributes
//   - timestamp
//   - event type set to delete action
//   - entity delete type set to soft_delete
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
	e.IDs.CopyTo(entityIdsMap)

	if e.Attributes.Len() != 0 {
		entityAttrs := attrs.PutEmptyMap(entityAttributes)
		e.Attributes.CopyTo(entityAttrs)
	}
}

var _ Event = (*Entity)(nil)
