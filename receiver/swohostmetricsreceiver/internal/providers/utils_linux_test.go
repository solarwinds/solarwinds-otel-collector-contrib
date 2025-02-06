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

package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseKeyValues_ReturnsExpectedResult(t *testing.T) {
	expectedResult := map[string]string{
		"Time zone": "Expected value",
	}
	text := "Time zone: Expected value"
	actualResult := ParseKeyValue(text, ": ", []string{"Time zone"})

	assert.Equal(t, expectedResult, actualResult)
}

func Test_ParseKeyValues_ReturnsEmptyMapForInvalidString(t *testing.T) {
	expectedResult := map[string]string{}
	text := "Time zone: Invalid: value"
	actualResult := ParseKeyValue(text, ": ", []string{"Time zone"})

	assert.Equal(t, expectedResult, actualResult)
}

func Test_ParseKeyValues_ReturnsEmptyMapForEmptyInput(t *testing.T) {
	expectedResult := map[string]string{}
	text := ""
	actualResult := ParseKeyValue(text, ": ", []string{"Time zone"})

	assert.Equal(t, expectedResult, actualResult)
}

func Test_ParseKeyValues_ReturnsEmptyMapForInvalidSeparator(t *testing.T) {
	expectedResult := map[string]string{}
	text := "Time zone: Invalid: value"
	actualResult := ParseKeyValue(text, "invalid separator", []string{"Time zone"})

	assert.Equal(t, expectedResult, actualResult)
}

func Test_ParseKeyValues_ReturnsExpectedMapForAllAttributes(t *testing.T) {
	expectedResult := map[string]string{
		"First param":  "value",
		"Second param": "value2",
		"Third param":  "value3",
	}
	text := `First param: value
Second param: value2
Third param: value3`
	actualResult := ParseKeyValue(text, ": ", []string{"First param", "Second param", "Third param"})

	assert.Equal(t, expectedResult, actualResult)
}

func Test_ParseKeyValues_FiltersInvalidLinesAndReturnsExpectedMap(t *testing.T) {
	expectedResult := map[string]string{
		"First param":  "value",
		"Second param": "value2",
		"Third param":  "value3",
	}
	text := `First param: value

Second param: value2
Invalid param=HELLO
Third param: value3`
	actualResult := ParseKeyValue(text, ": ", []string{"First param", "Second param", "Third param"})

	assert.Equal(t, expectedResult, actualResult)
}
