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
)

// Tests for Entity

func TestEntity_Validate_ValidEntity(t *testing.T) {
	entity := Entity{Entity: "entity", IDs: []string{"id1"}}

	err := entity.Validate()
	assert.NoError(t, err)
}

func TestEntity_Validate_EmptyEntityName(t *testing.T) {
	entity := Entity{Entity: "", IDs: []string{"id1"}}

	err := entity.Validate()
	assert.ErrorContains(t, err, "entity is mandatory")
}

func TestEntity_Validate_EmptyIDs(t *testing.T) {
	entity := Entity{Entity: "entity", IDs: []string{}}

	err := entity.Validate()
	assert.ErrorContains(t, err, "id is mandatory and must contain at least 1 item")
}

func TestEvent_validateActionAndContext_ValidUpdateActionLogContext(t *testing.T) {
	event := Event{Action: EventUpdateAction, Context: "log"}

	err := event.validateActionAndContext()
	assert.NoError(t, err)
}

func TestEvent_validateActionAndContext_ValidDeleteActionMetricContext(t *testing.T) {
	event := Event{Action: EventDeleteAction, Context: "metric"}

	err := event.validateActionAndContext()
	assert.NoError(t, err)
}

func TestEvent_validateActionAndContext_EmptyAction(t *testing.T) {
	event := Event{Action: "", Context: "log"}

	err := event.validateActionAndContext()
	assert.ErrorContains(t, err, "action is mandatory")
}

func TestEvent_validateActionAndContext_InvalidAction(t *testing.T) {
	event := Event{Action: "invalid", Context: "log"}

	err := event.validateActionAndContext()
	assert.ErrorContains(t, err, "action must be 'update' or 'delete', got 'invalid'")
}

func TestEvent_validateActionAndContext_EmptyContext(t *testing.T) {
	event := Event{Action: EventUpdateAction, Context: ""}

	err := event.validateActionAndContext()
	assert.ErrorContains(t, err, "context is mandatory")
}

func TestEvent_validateActionAndContext_InvalidContext(t *testing.T) {
	event := Event{Action: EventUpdateAction, Context: "invalid"}

	err := event.validateActionAndContext()
	assert.ErrorContains(t, err, "context must be 'log' or 'metric', got 'invalid'")
}

func TestRelationshipEvent_Validate_ValidRelationship(t *testing.T) {
	relationshipEvent := RelationshipEvent{
		Event:       Event{Action: EventUpdateAction, Context: "log"},
		Type:        "dependency",
		Source:      "a",
		Destination: "b",
	}

	err := relationshipEvent.Validate()
	assert.NoError(t, err)
}

func TestRelationshipEvent_Validate_EmptyType(t *testing.T) {
	relationshipEvent := RelationshipEvent{
		Event:       Event{Action: EventUpdateAction, Context: "log"},
		Type:        "",
		Source:      "a",
		Destination: "b",
	}

	err := relationshipEvent.Validate()
	assert.ErrorContains(t, err, "type is mandatory")
}

func TestRelationshipEvent_Validate_EmptySource(t *testing.T) {
	relationshipEvent := RelationshipEvent{
		Event:       Event{Action: EventUpdateAction, Context: "log"},
		Type:        "dependency",
		Source:      "",
		Destination: "b",
	}

	err := relationshipEvent.Validate()
	assert.ErrorContains(t, err, "source_entity is mandatory")
}

func TestRelationshipEvent_Validate_EmptyDestination(t *testing.T) {
	relationshipEvent := RelationshipEvent{
		Event:       Event{Action: EventUpdateAction, Context: "log"},
		Type:        "dependency",
		Source:      "a",
		Destination: "",
	}

	err := relationshipEvent.Validate()
	assert.ErrorContains(t, err, "destination_entity is mandatory")
}

func TestRelationshipEvent_Validate_InvalidEventFields(t *testing.T) {
	relationshipEvent := RelationshipEvent{
		Event:       Event{Action: "", Context: ""},
		Type:        "dependency",
		Source:      "a",
		Destination: "b",
	}

	err := relationshipEvent.Validate()
	assert.ErrorContains(t, err, "action is mandatory")
	assert.ErrorContains(t, err, "context is mandatory")
}

func TestEntityEvent_Validate_ValidEntityEvent(t *testing.T) {
	entityEvent := EntityEvent{
		Event:  Event{Action: EventDeleteAction, Context: "log"},
		Entity: "entity",
	}

	err := entityEvent.Validate()
	assert.NoError(t, err)
}

func TestEntityEvent_Validate_EmptyEntity(t *testing.T) {
	entityEvent := EntityEvent{
		Event:  Event{Action: EventUpdateAction, Context: "log"},
		Entity: "",
	}

	err := entityEvent.Validate()
	assert.ErrorContains(t, err, "entity is mandatory")
}

func TestEntityEvent_Validate_InvalidEventFields(t *testing.T) {
	entityEvent := EntityEvent{
		Event:  Event{Action: "", Context: ""},
		Entity: "entity",
	}

	err := entityEvent.Validate()
	assert.ErrorContains(t, err, "action is mandatory")
	assert.ErrorContains(t, err, "context is mandatory")
}
