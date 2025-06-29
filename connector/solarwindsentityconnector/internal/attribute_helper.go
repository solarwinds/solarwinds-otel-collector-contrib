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
	"go.opentelemetry.io/collector/pdata/plog"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

// getIDAttributes sets the entity id attributes in the log record as needed by SWO.
// Attributes are used to infer the entity in the system.
//
// Returns error if any of the attributes are missing in the resourceAttrs.
// If any ID attribute is missing, the entity would not be inferred.
func getIDAttributes(entityIds []string, resourceAttrs pcommon.Map) (pcommon.Map, error) {
	if len(entityIds) == 0 {
		return pcommon.Map{}, fmt.Errorf("entity ID attributes are empty")
	}

	attrs := pcommon.NewMap()
	for _, id := range entityIds {
		value, exists := findAttribute(id, resourceAttrs)

		if !exists {
			return pcommon.Map{}, fmt.Errorf("missing entity ID attribute: %s", id)
		}

		putAttribute(&attrs, id, value)
	}
	return attrs, nil
}

// setIdAttributesForRelationships sets the entity id attributes in the log record as needed by SWO for relationships.
// prefix - mandatory for same type relationships, but optional for relationships between different entity types.
// Returns bool indicating if prefixed attributes were found (if called with empty prefix, you might need to ignore it), and an error.
func setIdAttributesForRelationships(attrs pcommon.Map, entityIds []string, resourceAttrs pcommon.Map, name, prefix string) (bool, error) {
	if len(entityIds) == 0 {
		return false, fmt.Errorf("entity ID attributes are empty")
	}

	hasPrefix := false
	logIds := attrs.PutEmptyMap(name)
	for _, id := range entityIds {

		value, exists := findAttribute(prefix+id, resourceAttrs)
		if exists {
			putAttribute(&logIds, id, value)
			hasPrefix = true
			continue
		}

		value, exists = findAttribute(id, resourceAttrs)
		if exists {
			putAttribute(&logIds, id, value)
			continue
		}

		return false, fmt.Errorf("missing entity ID attribute: %s", id)
	}
	return hasPrefix, nil
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

// CreateEventLog prepares a clean LogRecordSlice, where log records representing events should be appended.
// Creates a resource log in input plog.Logs with single scope log decorated with attributes necessary for proper SWO ingestion.
func CreateEventLog(logs *plog.Logs) *plog.LogRecordSlice {
	resourceLog := logs.ResourceLogs().AppendEmpty()
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().Attributes().PutBool(entityEventAsLog, true)
	lrs := scopeLog.LogRecords()

	return &lrs
}
