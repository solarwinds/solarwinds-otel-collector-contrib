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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestSetIdAttributesDefaultEmpty(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")

	_, err := getIDAttributes([]string{}, resourceAttrs)
	assert.NotNil(t, err)
}

func TestSetIdAttributesSameTypeEmpty(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	destination := plog.NewLogRecord()
	hasPrefix, err := setIdAttributesForRelationships(destination.Attributes(), []string{}, resourceAttrs, entityIds, "src.")
	assert.NotNil(t, err)
	assert.False(t, hasPrefix)
}

func TestSetIdAttributesDefaultNoMatch(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")

	_, err := getIDAttributes([]string{"id3"}, resourceAttrs)
	assert.NotNil(t, err)
}

func TestSetIdAttributesSameTypetIdNoMatch(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	destination := plog.NewLogRecord()
	hasPrefix, err := setIdAttributesForRelationships(destination.Attributes(), []string{"id2"}, resourceAttrs, entityIds, "src.")
	assert.False(t, hasPrefix)
	assert.NotNil(t, err)
	ids, exists := destination.Attributes().Get(entityIds)
	assert.True(t, exists)
	assert.Equal(t, 0, ids.Map().Len())
}

func TestSetIdAttributesSameTypePrefixNoMatch(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	destination := plog.NewLogRecord()
	hasPrefix, err := setIdAttributesForRelationships(destination.Attributes(), []string{"id"}, resourceAttrs, entityIds, "prefix.")
	assert.False(t, hasPrefix)
	assert.NotNil(t, err)
	ids, exists := destination.Attributes().Get(entityIds)
	assert.True(t, exists)
	assert.Equal(t, 0, ids.Map().Len())
}

func TestSetIdAttributesDefaultMultiple(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("id1", "idvalue1")
	resourceAttrs.PutStr("id2", "idvalue2")

	ids, err := getIDAttributes([]string{"id1"}, resourceAttrs)
	assert.Nil(t, err)
	assert.Equal(t, 1, ids.Len())
	id, exists := ids.Get("id1")
	assert.True(t, exists)
	assert.Equal(t, "idvalue1", id.Str())
}

func TestSetIdAttributesSameTypeMultiple(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "idvalue1")
	resourceAttrs.PutStr("dst.id", "idvalue2")

	destination := plog.NewLogRecord()
	hasPrefix, err := setIdAttributesForRelationships(destination.Attributes(), []string{"id"}, resourceAttrs, entityIds, "src.")
	assert.True(t, hasPrefix)
	assert.Nil(t, err)
	ids, exists := destination.Attributes().Get(entityIds)
	assert.True(t, exists)
	assert.Equal(t, 1, ids.Map().Len())
	id, exists := ids.Map().Get("id")
	assert.True(t, exists)
	assert.Equal(t, "idvalue1", id.Str())
}

func TestSetAttributesSingle(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("attr1", "attrvalue1")
	resourceAttrs.PutStr("attr2", "attrvalue2")

	attributes := getAttributes([]string{"attr1"}, resourceAttrs)
	assert.Equal(t, 1, attributes.Len())
}

func TestSetAttributesEmpty(t *testing.T) {
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("attr1", "attrvalue1")
	resourceAttrs.PutStr("attr2", "attrvalue2")

	attributes := getAttributes([]string{}, resourceAttrs)
	assert.NotNil(t, attributes)
}

func TestPutStringAttribute(t *testing.T) {
	value := pcommon.NewValueStr("stringValue")

	destination := pcommon.NewMap()
	putAttribute(&destination, "string", value)
	inDest, exists := destination.Get("string")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, "stringValue", inDest.Str(), "Attribute value should match")
}

func TestPutBoolAttribute(t *testing.T) {
	value := pcommon.NewValueBool(true)

	destination := pcommon.NewMap()
	putAttribute(&destination, "bool", value)
	inDest, exists := destination.Get("bool")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, true, inDest.Bool(), "Attribute value should match")
}

func TestPutIntAttribute(t *testing.T) {
	value := pcommon.NewValueInt(123)
	destination := pcommon.NewMap()
	putAttribute(&destination, "int", value)
	inDest, exists := destination.Get("int")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, int64(123), inDest.Int(), "Attribute value should match")
}

func TestPutDoubleAttribute(t *testing.T) {
	value := pcommon.NewValueDouble(123.456)

	destination := pcommon.NewMap()
	putAttribute(&destination, "double", value)
	inDest, exists := destination.Get("double")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, 123.456, inDest.Double(), "Attribute value should match")
}

func TestPutBytesAttribute(t *testing.T) {
	value := pcommon.NewValueBytes()
	value.SetEmptyBytes().Append('1', '2')

	destination := pcommon.NewMap()
	putAttribute(&destination, "bytes", value)
	inDest, exists := destination.Get("bytes")
	assert.True(t, exists, "Attribute should exist in destination")
	assert.Equal(t, "12", string(inDest.Bytes().AsRaw()), "Attribute value should match")
}
