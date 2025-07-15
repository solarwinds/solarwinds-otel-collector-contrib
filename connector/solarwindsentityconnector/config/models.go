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
	"errors"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
)

type Entity struct {
	Type       string   `mapstructure:"entity"`
	IDs        []string `mapstructure:"id"`
	Attributes []string `mapstructure:"attributes"`
}

// Validate validates an individual entity configuration
func (e *Entity) Validate(index int) error {
	// Entity type is mandatory
	if e.Type == "" {
		return fmt.Errorf("schema.entities[%d].entity is mandatory", index)
	}

	// ID is mandatory with at least 1 item
	if len(e.IDs) == 0 {
		return fmt.Errorf("schema.entities[%d].id is mandatory and must contain at least 1 item", index)
	}

	return nil
}

type Events struct {
	Relationships []RelationshipEvent `mapstructure:"relationships"`
	Entities      []EntityEvent       `mapstructure:"entities"`
}

// Validate validates the events section of the configuration
func (e *Events) Validate(entities []Entity) error {
	// Events.entities is mandatory with at least 1 item
	if len(e.Entities) == 0 {
		return errors.New("schema.events.entities is mandatory and must contain at least 1 item")
	}

	// Create a map for quick entity type lookup
	entityTypes := make(map[string]bool)
	for _, entity := range entities {
		entityTypes[entity.Type] = true
	}

	// Validate entity events
	for i, entityEvent := range e.Entities {
		if err := entityEvent.Validate(i, entityTypes); err != nil {
			return err
		}
	}

	// Validate relationship events (optional section)
	for i, relationshipEvent := range e.Relationships {
		if err := relationshipEvent.Validate(i, entityTypes); err != nil {
			return err
		}
	}

	return nil
}

type Relationship struct {
	Type        string `mapstructure:"type"`
	Source      string `mapstructure:"source_entity"`
	Destination string `mapstructure:"destination_entity"`
}

type RelationshipEvent struct {
	Type        string   `mapstructure:"type"`
	Source      string   `mapstructure:"source_entity"`
	Destination string   `mapstructure:"destination_entity"`
	Attributes  []string `mapstructure:"attributes"`
	Conditions  []string `mapstructure:"conditions"`
	Context     string   `mapstructure:"context"`
	Action      string   `mapstructure:"action"`
}

// Validate validates an individual relationship event configuration
func (r *RelationshipEvent) Validate(index int, entityTypes map[string]bool) error {
	// Action is mandatory
	if r.Action == "" {
		return fmt.Errorf("schema.events.relationships[%d].action is mandatory", index)
	}

	// Action must be "update" or "delete"
	if r.Action != "update" && r.Action != "delete" {
		return fmt.Errorf("schema.events.relationships[%d].action must be 'update' or 'delete', got '%s'", index, r.Action)
	}

	// Context is mandatory
	if r.Context == "" {
		return fmt.Errorf("schema.events.relationships[%d].context is mandatory", index)
	}

	// Context must be "log" or "metric"
	if r.Context != ottllog.ContextName && r.Context != ottlmetric.ContextName {
		return fmt.Errorf("schema.events.relationships[%d].context must be '%s' or '%s', got '%s'",
			index, ottllog.ContextName, ottlmetric.ContextName, r.Context)
	}

	// Type is mandatory
	if r.Type == "" {
		return fmt.Errorf("schema.events.relationships[%d].type is mandatory", index)
	}

	// Source entity is mandatory
	if r.Source == "" {
		return fmt.Errorf("schema.events.relationships[%d].source_entity is mandatory", index)
	}

	// Destination entity is mandatory
	if r.Destination == "" {
		return fmt.Errorf("schema.events.relationships[%d].destination_entity is mandatory", index)
	}

	// Validate that source and destination entity types exist in schema.entities
	if !entityTypes[r.Source] {
		return fmt.Errorf("schema.events.relationships[%d].source_entity '%s' must be defined in schema.entities", index, r.Source)
	}

	if !entityTypes[r.Destination] {
		return fmt.Errorf("schema.events.relationships[%d].destination_entity '%s' must be defined in schema.entities", index, r.Destination)
	}

	return nil
}

type EntityEvent struct {
	Context    string   `mapstructure:"context"`
	Conditions []string `mapstructure:"conditions"`
	Type       string   `mapstructure:"type"`
	Action     string   `mapstructure:"action"`
}

// Validate validates an individual entity event configuration
func (e *EntityEvent) Validate(index int, entityTypes map[string]bool) error {
	// Action is mandatory
	if e.Action == "" {
		return fmt.Errorf("schema.events.entities[%d].action is mandatory", index)
	}

	// Action must be "update" or "delete"
	if e.Action != "update" && e.Action != "delete" {
		return fmt.Errorf("schema.events.entities[%d].action must be 'update' or 'delete', got '%s'", index, e.Action)
	}

	// Context is mandatory
	if e.Context == "" {
		return fmt.Errorf("schema.events.entities[%d].context is mandatory", index)
	}

	// Context must be "log" or "metric"
	if e.Context != ottllog.ContextName && e.Context != ottlmetric.ContextName {
		return fmt.Errorf("schema.events.entities[%d].context must be '%s' or '%s', got '%s'",
			index, ottllog.ContextName, ottlmetric.ContextName, e.Context)
	}

	// Type is mandatory
	if e.Type == "" {
		return fmt.Errorf("schema.events.entities[%d].type is mandatory", index)
	}

	// Validate that the entity type exists in schema.entities
	if !entityTypes[e.Type] {
		return fmt.Errorf("schema.events.entities[%d].type '%s' must be defined in schema.entities", index, e.Type)
	}

	return nil
}
