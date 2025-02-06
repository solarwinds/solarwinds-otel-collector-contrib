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

package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func RequireAttribute(t *testing.T, attributes pcommon.Map, attributeKey string, expectedValue string) {
	value, exists := attributes.Get(attributeKey)
	require.Truef(t, exists, "The %s attribute should exists", attributeKey)
	require.Equalf(t, expectedValue, value.AsString(), "The %s attribute should be %s", attributeKey, expectedValue)
}
