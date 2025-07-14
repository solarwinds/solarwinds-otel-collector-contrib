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

type RelationshipEntity struct {
	Type string
	IDs  pcommon.Map
}

type Relationship struct {
	Action      string
	Type        string
	Source      RelationshipEntity
	Destination RelationshipEntity
	Attributes  pcommon.Map
}

var _ Event = (*Relationship)(nil)

// Update adds a log record for the relationship update event.
// Log record is decorated by following attributes:
//   - Source entity type
//   - Source entity ids
//   - Destination entity type
//   - Destination entity ids
//   - Relationship type
//   - timestamp
//   - event type set to update action
func (r Relationship) Update(logRecords *plog.LogRecordSlice) {
	logRecord := logRecords.AppendEmpty()
	logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	attrs := logRecord.Attributes()
	attrs.PutStr(entityEventType, relationshipUpdateEventType)

	// Copy id, entity and relationship attributes as pcommon.Map to the log record
	srcIds := attrs.PutEmptyMap(relationshipSrcEntityIds)
	destIds := attrs.PutEmptyMap(relationshipDestEntityIds)

	r.Source.IDs.CopyTo(srcIds)
	r.Destination.IDs.CopyTo(destIds)
	if r.Attributes.Len() != 0 {
		relationshipAttrs := attrs.PutEmptyMap(relationshipAttributes)
		r.Attributes.CopyTo(relationshipAttrs)
	}

	attrs.PutStr(relationshipType, r.Type)
	attrs.PutStr(srcEntityType, r.Source.Type)
	attrs.PutStr(destEntityType, r.Destination.Type)
}

// Delete adds the relationship as a log record with attributes needed be SWO.
// Log record is decorated by following attributes:
//   - Source entity type
//   - Source entity ids
//   - Destination entity type
//   - Destination entity ids
//   - Relationship type
//   - timestamp
//   - event type set to delete action
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
