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

func TestIdentifyAttributes_WithPrefixes(t *testing.T) {
	// Create a resource attributes map
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "sourceValue")
	resourceAttrs.PutStr("dst.id", "destinationValue")
	resourceAttrs.PutStr("commonAttr", "commonValue")

	// Call IdentifyAttributes with prefixes
	attributes := IdentifyAttributes(resourceAttrs, "src.", "dst.")

	// Verify Source attributes
	assert.Equal(t, 1, len(attributes.Source))
	assert.Equal(t, "sourceValue", attributes.Source["id"].Str())

	// Verify Destination attributes
	assert.Equal(t, 1, len(attributes.Destination))
	assert.Equal(t, "destinationValue", attributes.Destination["id"].Str())

	// Verify Common attributes
	assert.Equal(t, 1, len(attributes.Common))
	assert.Equal(t, "commonValue", attributes.Common["commonAttr"].Str())
}

func TestIdentifyAttributes_EmptySrcPrefix(t *testing.T) {
	// Create a resource attributes map
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "sourceValue")
	resourceAttrs.PutStr("dst.id", "destinationValue")
	resourceAttrs.PutStr("commonAttr", "commonValue")

	// Call IdentifyAttributes with an empty srcPrefix
	attributes := IdentifyAttributes(resourceAttrs, "", "dst.")

	// Verify Source attributes
	assert.Equal(t, 0, len(attributes.Source))

	// Verify Destination attributes
	assert.Equal(t, 1, len(attributes.Destination))
	assert.Equal(t, "destinationValue", attributes.Destination["id"].Str())

	// Verify Common attributes
	assert.Equal(t, 2, len(attributes.Common))
	assert.Equal(t, "sourceValue", attributes.Common["src.id"].Str())
	assert.Equal(t, "commonValue", attributes.Common["commonAttr"].Str())
}

func TestIdentifyAttributes_EmptyDestPrefix(t *testing.T) {
	// Create a resource attributes map
	resourceAttrs := pcommon.NewMap()
	resourceAttrs.PutStr("src.id", "sourceValue")
	resourceAttrs.PutStr("dst.id", "destinationValue")
	resourceAttrs.PutStr("commonAttr", "commonValue")

	attributes := IdentifyAttributes(resourceAttrs, "src.", "")

	// Verify Source attributes
	assert.Equal(t, 1, len(attributes.Source))
	assert.Equal(t, "sourceValue", attributes.Source["id"].Str())

	// Verify Destination attributes
	assert.Equal(t, 0, len(attributes.Destination))

	// Verify Common attributes
	assert.Equal(t, 2, len(attributes.Common))
	assert.Equal(t, "destinationValue", attributes.Common["dst.id"].Str())
	assert.Equal(t, "commonValue", attributes.Common["commonAttr"].Str())
}

func TestIdentifyAttributes_EmptyResourceAttributes(t *testing.T) {
	resourceAttrs := pcommon.NewMap()

	attributes := IdentifyAttributes(resourceAttrs, "src.", "dst.")

	assert.Equal(t, 0, len(attributes.Source))
	assert.Equal(t, 0, len(attributes.Destination))
	assert.Equal(t, 0, len(attributes.Common))
}
