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
	"go.opentelemetry.io/collector/pdata/plog"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

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
