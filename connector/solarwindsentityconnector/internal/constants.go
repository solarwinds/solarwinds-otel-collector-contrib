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

const (
	// Log properties
	EntityEventAsLog = "otel.entity.event_as_log"
	entityEventType  = "otel.entity.event.type"

	// Event type values
	entityUpdateEventType       = "entity_state"
	relationshipUpdateEventType = "entity_relationship_state"
	relationshipDeleteEventType = "entity_relationship_delete"

	// Entity properties
	entityType       = "otel.entity.type"
	entityIds        = "otel.entity.id"
	entityAttributes = "otel.entity.attributes"

	// Relationship properties
	relationshipSrcEntityIds  = "otel.entity_relationship.source_entity.id"
	relationshipDestEntityIds = "otel.entity_relationship.destination_entity.id"
	relationshipAttributes    = "otel.entity_relationship.attributes"
	relationshipType          = "otel.entity_relationship.type"
	srcEntityType             = "otel.entity_relationship.source_entity.type"
	destEntityType            = "otel.entity_relationship.destination_entity.type"
)
