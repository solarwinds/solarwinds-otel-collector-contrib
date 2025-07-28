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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
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
					Entity: "test-entity",
					Event: Event{
						Context:    "log",
						Action:     "update",
						Conditions: []string{"true"},
					},
				},
				{
					Entity: "test-entity-metric",
					Event: Event{
						Context:    "metric",
						Action:     "update",
						Conditions: []string{"true"},
					},
				},
			},
			Relationships: []RelationshipEvent{
				{
					Type:        "test-relationship",
					Source:      "source-entity",
					Destination: "dest-entity",
					Event: Event{
						Context:    "log",
						Action:     "update",
						Conditions: []string{"true"},
					},
				},
				{
					Type:        "test-relationship-metric",
					Source:      "source-entity",
					Destination: "dest-entity",
					Event: Event{
						Context:    "metric",
						Action:     "update",
						Conditions: []string{"true"},
					},
				},
			},
		},
	}

	parsedEvents, err := createParsedEvents(schema, settings)
	assert.NoError(t, err)

	// Test log events - should have 2 events total (1 entity + 1 relationship)
	assert.Len(t, parsedEvents.LogEvents.Events, 2)

	// Find and test entity event
	var logEntityEvent ParsedEventInterface[ottllog.TransformContext]
	var logRelationshipEvent ParsedEventInterface[ottllog.TransformContext]
	for _, event := range parsedEvents.LogEvents.Events {
		if event.IsEntityEvent() {
			logEntityEvent = event
		} else {
			logRelationshipEvent = event
		}
	}

	assert.NotNil(t, logEntityEvent)
	assert.Equal(t, "test-entity", logEntityEvent.GetEntityEvent().Entity)
	assert.NotNil(t, logEntityEvent.GetConditionSeq())

	assert.NotNil(t, logRelationshipEvent)
	assert.Equal(t, "test-relationship", logRelationshipEvent.GetRelationshipEvent().Type)
	assert.NotNil(t, logRelationshipEvent.GetConditionSeq())

	// Test metric events - should have 2 events total (1 entity + 1 relationship)
	assert.Len(t, parsedEvents.MetricEvents.Events, 2)

	// Find and test entity event
	var metricEntityEvent ParsedEventInterface[ottlmetric.TransformContext]
	var metricRelationshipEvent ParsedEventInterface[ottlmetric.TransformContext]
	for _, event := range parsedEvents.MetricEvents.Events {
		if event.IsEntityEvent() {
			metricEntityEvent = event
		} else {
			metricRelationshipEvent = event
		}
	}

	assert.NotNil(t, metricEntityEvent)
	assert.Equal(t, "test-entity-metric", metricEntityEvent.GetEntityEvent().Entity)
	assert.NotNil(t, metricEntityEvent.GetConditionSeq())

	assert.NotNil(t, metricRelationshipEvent)
	assert.Equal(t, "test-relationship-metric", metricRelationshipEvent.GetRelationshipEvent().Type)
	assert.NotNil(t, metricRelationshipEvent.GetConditionSeq())
}

func TestCreateParsedEventsWithConverters(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	// Create parser for log context
	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Entity: "test-entity",
					Event: Event{
						Context:    "log",
						Action:     "update",
						Conditions: []string{"Len(\"test\") != 0"},
					},
				},
				{
					Entity: "test-entity-metric",
					Event: Event{
						Context:    "metric",
						Action:     "update",
						Conditions: []string{"MD5(\"test\") != \"expected-value\""},
					},
				},
			},
			Relationships: []RelationshipEvent{
				{
					Type:        "test-relationship",
					Source:      "source-entity",
					Destination: "dest-entity",
					Event: Event{
						Context:    "log",
						Action:     "update",
						Conditions: []string{"SHA1(\"test\") != \"expected-value\""},
					},
				},
				{
					Type:        "test-relationship-metric",
					Source:      "source-entity",
					Destination: "dest-entity",
					Event: Event{
						Context:    "metric",
						Action:     "update",
						Conditions: []string{"Hex(\"test\") != \"1A6B32A\""},
					},
				},
			},
		},
	}
	parsedEvents, err := createParsedEvents(schema, settings)
	assert.NoError(t, err)

	// Test log events - should have 2 events total (1 entity + 1 relationship)
	assert.Len(t, parsedEvents.LogEvents.Events, 2)

	// Find and test entity event
	var logEntityEvent ParsedEventInterface[ottllog.TransformContext]
	var logRelationshipEvent ParsedEventInterface[ottllog.TransformContext]
	for _, event := range parsedEvents.LogEvents.Events {
		if event.IsEntityEvent() {
			logEntityEvent = event
		} else {
			logRelationshipEvent = event
		}
	}

	assert.NotNil(t, logEntityEvent)
	assert.Equal(t, "test-entity", logEntityEvent.GetEntityEvent().Entity)
	assert.NotNil(t, logEntityEvent.GetConditionSeq())

	assert.NotNil(t, logRelationshipEvent)
	assert.Equal(t, "test-relationship", logRelationshipEvent.GetRelationshipEvent().Type)
	assert.NotNil(t, logRelationshipEvent.GetConditionSeq())

	// Test metric events - should have 2 events total (1 entity + 1 relationship)
	assert.Len(t, parsedEvents.MetricEvents.Events, 2)

	// Find and test entity event
	var metricEntityEvent ParsedEventInterface[ottlmetric.TransformContext]
	var metricRelationshipEvent ParsedEventInterface[ottlmetric.TransformContext]
	for _, event := range parsedEvents.MetricEvents.Events {
		if event.IsEntityEvent() {
			metricEntityEvent = event
		} else {
			metricRelationshipEvent = event
		}
	}

	assert.NotNil(t, metricEntityEvent)
	assert.Equal(t, "test-entity-metric", metricEntityEvent.GetEntityEvent().Entity)
	assert.NotNil(t, metricEntityEvent.GetConditionSeq())

	assert.NotNil(t, metricRelationshipEvent)
	assert.Equal(t, "test-relationship-metric", metricRelationshipEvent.GetRelationshipEvent().Type)
	assert.NotNil(t, metricRelationshipEvent.GetConditionSeq())
}

func TestCreateParsedEventsEmptyConditions(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Entity: "test-entity",
					Event: Event{
						Context: "log",
						Action:  "update",
					},
				},
			},
		},
	}

	parsedEvents, err := createParsedEvents(schema, settings)
	require.NoError(t, err)
	require.Len(t, parsedEvents.LogEvents.Events, 1)

	// Verify it's an entity event
	event := parsedEvents.LogEvents.Events[0]
	require.True(t, event.IsEntityEvent())
	require.NotNil(t, event.GetEntityEvent())
}

func TestCreateParsedEventsUnknownContext(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Entity: "test-entity",
					Event: Event{
						Context:    "unknown-context",
						Action:     "update",
						Conditions: []string{"true"},
					},
				},
			},
		},
	}

	_, err := createParsedEvents(schema, settings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported context: unknown-context")
}

func TestCreateParsedEventsWithInvalidConditions(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities: []EntityEvent{
				{
					Entity: "invalid-condition",
					Event: Event{
						Context:    "log",
						Action:     "update",
						Conditions: []string{"InvalidFunction()"},
					},
				},
			},
		},
	}

	_, err := createParsedEvents(schema, settings)
	require.Error(t, err)
}

func TestCreateParsedEventsEmptySchema(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	schema := Schema{
		Events: Events{
			Entities:      []EntityEvent{},
			Relationships: []RelationshipEvent{},
		},
	}

	parsedEvents, err := createParsedEvents(schema, settings)
	require.NoError(t, err)

	assert.Len(t, parsedEvents.LogEvents.Events, 0)
	assert.Len(t, parsedEvents.MetricEvents.Events, 0)
}
