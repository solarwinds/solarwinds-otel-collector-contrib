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
	"fmt"
	"github.com/solarwinds/solarwinds-otel-collector-releases/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"time"
)

// BuildEventLog prepares a clean LogRecordSlice, where log records representing events should be appended.
// In given plog.Logs creates a resource log with one scope log and set attributes needed by SWO ingestion.
func BuildEventLog(logs *plog.Logs) *plog.LogRecordSlice {
	resourceLog := logs.ResourceLogs().AppendEmpty()
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().Attributes().PutBool(entityEventAsLog, true)
	lrs := scopeLog.LogRecords()

	return &lrs
}

// setIdAttributes sets the entity id attributes in the log record as needed by SWO.
// Attributes are used to infer the entity in the system.
//
// Returns error if any of the attributes are missing in the resourceAttrs.
// If any ID attribute is missing, the entity would not be inferred.
func setIdAttributes(attrs pcommon.Map, entityIds []string, resourceAttrs pcommon.Map, name string) error {
	if len(entityIds) == 0 {
		return fmt.Errorf("entity id attributes are empty")
	}

	logIds := attrs.PutEmptyMap(name)
	for _, id := range entityIds {
		value, exists := findAttribute(id, resourceAttrs)
		if !exists {
			return fmt.Errorf("failed to find entity id attribute %s", id)
		}
		putAttribute(&logIds, id, value)
	}
	return nil
}

// setEntityAttributes sets the entity attributes in the log record as needed by SWO.
// Attributes are used to update the entity.
func setAttributes(attrs pcommon.Map, entityAttrs []string, resourceAttrs pcommon.Map, name string) {
	if len(entityAttrs) == 0 {
		return
	}

	logIds := attrs.PutEmptyMap(name)
	for _, attr := range entityAttrs {
		value, exists := findAttribute(attr, resourceAttrs)
		if !exists {
			continue
		}
		putAttribute(&logIds, attr, value)
	}
}

// findAttribute checks if the attribute identified as key exists in the source pcommon.Map.
func findAttribute(key string, src pcommon.Map) (pcommon.Value, bool) {
	attrVal, ok := src.Get(key)
	return attrVal, ok
}

// putAttribute copies the value of attribute identified as key, to destination pcommon.Map.
func putAttribute(dest *pcommon.Map, key string, attrValue pcommon.Value) {
	switch typeAttr := attrValue.Type(); typeAttr {
	case pcommon.ValueTypeInt:
		dest.PutInt(key, attrValue.Int())
	case pcommon.ValueTypeDouble:
		dest.PutDouble(key, attrValue.Double())
	case pcommon.ValueTypeBool:
		dest.PutBool(key, attrValue.Bool())
	case pcommon.ValueTypeBytes:
		value := attrValue.Bytes().AsRaw()
		dest.PutEmptyBytes(key).FromRaw(value)
	default:
		dest.PutStr(key, attrValue.Str())
	}
}

func CreateEntityEvent(resourceAttrs pcommon.Map, entity config.Entity) (plog.LogRecord, error) {
	lr := plog.NewLogRecord()
	attrs := lr.Attributes()
	attrs.PutStr(entityType, entity.Type)

	if err := setIdAttributes(attrs, entity.IDs, resourceAttrs, entityIds); err != nil {
		return plog.LogRecord{}, err
	}

	setAttributes(attrs, entity.Attributes, resourceAttrs, entityAttributes)

	lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	return lr, nil
}

func CreateRelationshipEvent(resourceAttrs pcommon.Map, relationship config.Relationship, source, dest config.Entity) (plog.LogRecord, error) {
	lr := plog.NewLogRecord()
	attrs := lr.Attributes()

	attrs.PutStr(relationshipType, relationship.Type)
	attrs.PutStr(srcEntityType, source.Type)
	attrs.PutStr(destEntityType, dest.Type)

	if err := setIdAttributes(attrs, source.IDs, resourceAttrs, relationshipSrcEntityIds); err != nil {
		return plog.LogRecord{}, fmt.Errorf("source entity: %w", err)
	}

	if err := setIdAttributes(attrs, dest.IDs, resourceAttrs, relationshipDestEntityIds); err != nil {
		return plog.LogRecord{}, fmt.Errorf("destination entity: %w", err)
	}

	setAttributes(attrs, relationship.Attributes, resourceAttrs, relationshipAttributes)

	lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return lr, nil
}
