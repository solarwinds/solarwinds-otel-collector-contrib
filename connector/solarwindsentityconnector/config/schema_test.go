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
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func TestValidate_ValidSchema(t *testing.T) {
	schema := Schema{
		Entities: []Entity{{Entity: "a"}, {Entity: "b"}},
		Events: Events{
			Entities:      []EntityEvent{{Entity: "a"}},
			Relationships: []RelationshipEvent{{Source: "a", Destination: "b"}},
		},
	}

	err := schema.Validate()
	assert.NoError(t, err)
}

func TestValidate_EmptySchemaIsValid(t *testing.T) {
	schema := Schema{
		Entities: []Entity{},
		Events:   Events{Entities: []EntityEvent{}, Relationships: []RelationshipEvent{}},
	}

	err := schema.Validate()
	assert.NoError(t, err)
}

func TestValidate_MixedValidAndInvalidReferences(t *testing.T) {
	schema := Schema{
		Entities: []Entity{
			{Entity: "a"},
			{Entity: "b"},
		},
		Events: Events{
			Entities: []EntityEvent{
				{Entity: "a"},         // Valid
				{Entity: "undefined"}, // Invalid
				{Entity: "b"},         // Valid
			},
			Relationships: []RelationshipEvent{
				{Source: "a", Destination: "b"},         // Valid
				{Source: "undefined", Destination: "b"}, // Invalid source
				{Source: "a", Destination: "undefined"}, // Invalid destination
			},
		},
	}

	err := schema.Validate()
	assert.ErrorContains(t, err, "events::entities::1::entity: 'undefined' must be defined in 'entities'")
	assert.ErrorContains(t, err, "events::relationships::1::source_entity: 'undefined' must be defined in 'entities'")
	assert.ErrorContains(t, err, "events::relationships::2::destination_entity: 'undefined' must be defined in 'entities'")
	assert.NotContains(t, err.Error(), "'a' must be defined in 'entities'")
	assert.NotContains(t, err.Error(), "'b' must be defined in 'entities'")
}

func TestUnmarshal_ValidSchemaWithSimpleConditions(t *testing.T) {

	testEntity := Entity{Entity: "a", IDs: []string{"id1"}}
	schema := Schema{
		Entities: []Entity{testEntity},
		Events: Events{
			Entities: []EntityEvent{
				{Entity: "a", Event: Event{Action: EventUpdateAction, Context: "log"}},
			},
			Relationships: []RelationshipEvent{
				{Source: "a", Destination: "b", Type: "childOf", Event: Event{Action: EventUpdateAction, Context: "metric"}},
			},
		},
	}

	settings := component.TelemetrySettings{
		Logger: zap.NewNop(),
	}
	parsedSchema, err := schema.Unmarshal(settings)

	assert.NoError(t, err)
	assert.NotNil(t, parsedSchema)

	// Check entities map
	expectedEntities := map[string]Entity{"a": testEntity}
	assert.Equal(t, expectedEntities, parsedSchema.Entities)

	// Check that events structure is created
	assert.Len(t, parsedSchema.Events.LogEvents.Entities, 1)
	assert.Len(t, parsedSchema.Events.MetricEvents.Relationships, 1)
}
