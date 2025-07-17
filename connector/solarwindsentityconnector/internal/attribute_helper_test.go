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
)

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
