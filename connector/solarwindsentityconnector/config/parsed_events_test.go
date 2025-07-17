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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
)

func TestCreateParsedEvents(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Type:       "test-entity",
					Context:    "log",
					Action:     "update",
					Conditions: []string{"true"},
				},
				{
					Type:       "test-entity-metric",
					Context:    "metric",
					Action:     "update",
					Conditions: []string{"true"},
				},
			},
			Relationships: []RelationshipEvent{
				{
					Type:        "test-relationship",
					Source:      "source-entity",
					Destination: "dest-entity",
					Context:     "log",
					Action:      "update",
					Conditions:  []string{"true"},
				},
				{
					Type:        "test-relationship-metric",
					Source:      "source-entity",
					Destination: "dest-entity",
					Context:     "metric",
					Action:      "update",
					Conditions:  []string{"true"},
				},
			},
		},
	}

	parsedEvents := CreateParsedEvents(schema, settings)

	// Test that parsers are created
	assert.NotNil(t, parsedEvents.LogEvents.Parser)
	assert.NotNil(t, parsedEvents.MetricEvents.Parser)

	// Test log events
	assert.Len(t, parsedEvents.LogEvents.Entities, 1)
	assert.Equal(t, "test-entity", parsedEvents.LogEvents.Entities[0].Definition.Type)
	assert.NotNil(t, parsedEvents.LogEvents.Entities[0].ConditionSeq)

	assert.Len(t, parsedEvents.LogEvents.Relationships, 1)
	assert.Equal(t, "test-relationship", parsedEvents.LogEvents.Relationships[0].Definition.Type)
	assert.NotNil(t, parsedEvents.LogEvents.Relationships[0].ConditionSeq)

	// Test metric events
	assert.Len(t, parsedEvents.MetricEvents.Entities, 1)
	assert.Equal(t, "test-entity-metric", parsedEvents.MetricEvents.Entities[0].Definition.Type)
	assert.NotNil(t, parsedEvents.MetricEvents.Entities[0].ConditionSeq)

	assert.Len(t, parsedEvents.MetricEvents.Relationships, 1)
	assert.Equal(t, "test-relationship-metric", parsedEvents.MetricEvents.Relationships[0].Definition.Type)
	assert.NotNil(t, parsedEvents.MetricEvents.Relationships[0].ConditionSeq)
}

func TestCreateParsedEventsWithConverters(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	// Create parser for log context
	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Type:       "test-entity",
					Context:    "log",
					Action:     "update",
					Conditions: []string{"Len(\"test\") != 0"},
				},
				{
					Type:       "test-entity-metric",
					Context:    "metric",
					Action:     "update",
					Conditions: []string{"MD5(\"test\") != \"expected-value\""},
				},
			},
			Relationships: []RelationshipEvent{
				{
					Type:        "test-relationship",
					Source:      "source-entity",
					Destination: "dest-entity",
					Context:     "log",
					Action:      "update",
					Conditions:  []string{"SHA1(\"test\") != \"expected-value\""},
				},
				{
					Type:        "test-relationship-metric",
					Source:      "source-entity",
					Destination: "dest-entity",
					Context:     "metric",
					Action:      "update",
					Conditions:  []string{"Hex(\"test\") != \"1A6B32A\""},
				},
			},
		},
	}
	parsedEvents := CreateParsedEvents(schema, settings)

	assert.NotNil(t, parsedEvents.LogEvents.Parser)
	assert.NotNil(t, parsedEvents.MetricEvents.Parser)

	// Test log events
	assert.Len(t, parsedEvents.LogEvents.Entities, 1)
	assert.Equal(t, "test-entity", parsedEvents.LogEvents.Entities[0].Definition.Type)
	assert.NotNil(t, parsedEvents.LogEvents.Entities[0].ConditionSeq)

	assert.Len(t, parsedEvents.LogEvents.Relationships, 1)
	assert.Equal(t, "test-relationship", parsedEvents.LogEvents.Relationships[0].Definition.Type)
	assert.NotNil(t, parsedEvents.LogEvents.Relationships[0].ConditionSeq)

	// Test metric events
	assert.Len(t, parsedEvents.MetricEvents.Entities, 1)
	assert.Equal(t, "test-entity-metric", parsedEvents.MetricEvents.Entities[0].Definition.Type)
	assert.NotNil(t, parsedEvents.MetricEvents.Entities[0].ConditionSeq)

	assert.Len(t, parsedEvents.MetricEvents.Relationships, 1)
	assert.Equal(t, "test-relationship-metric", parsedEvents.MetricEvents.Relationships[0].Definition.Type)
	assert.NotNil(t, parsedEvents.MetricEvents.Relationships[0].ConditionSeq)
}

func TestCreateParsedEventsEmptyConditions(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Type:    "test-entity",
					Context: "log",
					Action:  "update",
				},
			},
		},
	}

	parsedEvents := CreateParsedEvents(schema, settings)
	require.Len(t, parsedEvents.LogEvents.Entities, 1)
}

func TestCreateParsedEventsUnknownContext(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Type:       "test-entity",
					Context:    "unknown-context", // This should be ignored
					Action:     "update",
					Conditions: []string{"true"},
				},
				{
					Type:       "test-entity-log",
					Context:    "log",
					Action:     "update",
					Conditions: []string{"true"},
				},
			},
		},
	}

	parsedEvents := CreateParsedEvents(schema, settings)

	// Only log event should be parsed, unknown context should be ignored
	assert.Len(t, parsedEvents.LogEvents.Entities, 1)
	assert.Len(t, parsedEvents.MetricEvents.Entities, 0)
	assert.Equal(t, "test-entity-log", parsedEvents.LogEvents.Entities[0].Definition.Type)
}

func TestCreateParsedEventsWithInvalidConditions(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Type:       "invalid-condition",
					Context:    "log",
					Action:     "update",
					Conditions: []string{"InvalidFunction()"},
				},
			},
		},
	}

	// This should panic or handle error gracefully
	assert.Panics(t, func() {
		CreateParsedEvents(schema, settings)
	})
}

func TestCreateParsedEventsEmptySchema(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities:      []EntityEvent{},
			Relationships: []RelationshipEvent{},
		},
	}

	parsedEvents := CreateParsedEvents(schema, settings)

	assert.NotNil(t, parsedEvents.LogEvents.Parser)
	assert.NotNil(t, parsedEvents.MetricEvents.Parser)
	assert.Len(t, parsedEvents.LogEvents.Entities, 0)
	assert.Len(t, parsedEvents.LogEvents.Relationships, 0)
	assert.Len(t, parsedEvents.MetricEvents.Entities, 0)
	assert.Len(t, parsedEvents.MetricEvents.Relationships, 0)
}
