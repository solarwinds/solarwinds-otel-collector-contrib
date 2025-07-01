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

package attributesdecorator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestDecorateResourceAttributesForMetrics(t *testing.T) {
	additionalAttributes := map[string]string{
		"service.name": "test-service",
		"environment":  "test-env",
	}

	ms := pmetric.NewMetrics()
	ms.ResourceMetrics().AppendEmpty().Resource().Attributes().PutStr("existing.attribute", "test-value")

	DecorateResourceAttributes(ms.ResourceMetrics(), additionalAttributes)

	expectedAttributes := pcommon.NewMap()
	err := expectedAttributes.FromRaw(map[string]any{
		"existing.attribute": "test-value",
		"service.name":       "test-service",
		"environment":        "test-env",
	})
	assert.NoError(t, err)

	actualAttributes := ms.ResourceMetrics().At(0).Resource().Attributes()

	aaRaw := actualAttributes.AsRaw()
	eaRaw := expectedAttributes.AsRaw()
	assert.Equal(t, eaRaw, aaRaw, "Expected and actual attributes should match")
}

func TestDecorateResourceAttributesForLogs(t *testing.T) {
	additionalAttributes := map[string]string{
		"service.name": "test-service",
		"environment":  "test-env",
	}

	ls := plog.NewLogs()
	ls.ResourceLogs().AppendEmpty().Resource().Attributes().PutStr("existing.attribute", "test-value")

	DecorateResourceAttributes(ls.ResourceLogs(), additionalAttributes)

	expectedAttributes := pcommon.NewMap()
	err := expectedAttributes.FromRaw(map[string]any{
		"existing.attribute": "test-value",
		"service.name":       "test-service",
		"environment":        "test-env",
	})
	assert.NoError(t, err)

	actualAttributes := ls.ResourceLogs().At(0).Resource().Attributes()

	aaRaw := actualAttributes.AsRaw()
	eaRaw := expectedAttributes.AsRaw()
	assert.Equal(t, eaRaw, aaRaw, "Expected and actual attributes should match")
}

func TestDecorateResourceAttributesForTraces(t *testing.T) {
	additionalAttributes := map[string]string{
		"service.name": "test-service",
		"environment":  "test-env",
	}

	ts := ptrace.NewTraces()
	ts.ResourceSpans().AppendEmpty().Resource().Attributes().PutStr("existing.attribute", "test-value")

	DecorateResourceAttributes(ts.ResourceSpans(), additionalAttributes)

	expectedAttributes := pcommon.NewMap()
	err := expectedAttributes.FromRaw(map[string]any{
		"existing.attribute": "test-value",
		"service.name":       "test-service",
		"environment":        "test-env",
	})
	assert.NoError(t, err)

	actualAttributes := ts.ResourceSpans().At(0).Resource().Attributes()

	aaRaw := actualAttributes.AsRaw()
	eaRaw := expectedAttributes.AsRaw()
	assert.Equal(t, eaRaw, aaRaw, "Expected and actual attributes should match")
}
