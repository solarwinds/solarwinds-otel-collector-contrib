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
	Entity     string   `mapstructure:"entity"`
	IDs        []string `mapstructure:"id"`
	Attributes []string `mapstructure:"attributes"`
}

func (e *Entity) Validate(index int) error {
	var allErrs error

	if e.Entity == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.entities[%d].entity is mandatory", index))
	}

	if len(e.IDs) == 0 {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.entities[%d].id is mandatory and must contain at least 1 item", index))
	}

	return allErrs
}

type Events struct {
	Relationships []RelationshipEvent `mapstructure:"relationships"`
	Entities      []EntityEvent       `mapstructure:"entities"`
}

func (e *Events) Validate(entities []Entity) error {
	var allErrs error

	if len(e.Entities) == 0 {
		allErrs = errors.Join(allErrs, errors.New("schema.events.entities is mandatory and must contain at least 1 item"))
	}

	entityTypes := make(map[string]bool)
	for _, entity := range entities {
		entityTypes[entity.Entity] = true
	}

	for i, entityEvent := range e.Entities {
		if err := entityEvent.Validate(i, entityTypes); err != nil {
			allErrs = errors.Join(allErrs, err)
		}
	}

	for i, relationshipEvent := range e.Relationships {
		if err := relationshipEvent.Validate(i, entityTypes); err != nil {
			allErrs = errors.Join(allErrs, err)
		}
	}

	return allErrs
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

func (r *RelationshipEvent) Validate(index int, entityTypes map[string]bool) error {
	var allErrs error

	if r.Action == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].action is mandatory", index))
	} else if r.Action != "update" && r.Action != "delete" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].action must be 'update' or 'delete', got '%s'", index, r.Action))
	}

	if r.Context == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].context is mandatory", index))
	} else if r.Context != ottllog.ContextName && r.Context != ottlmetric.ContextName {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].context must be '%s' or '%s', got '%s'",
			index, ottllog.ContextName, ottlmetric.ContextName, r.Context))
	}

	if r.Type == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].type is mandatory", index))
	}

	if r.Source == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].source_entity is mandatory", index))
	} else if !entityTypes[r.Source] {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].source_entity '%s' must be defined in schema.entities", index, r.Source))
	}

	if r.Destination == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].destination_entity is mandatory", index))
	} else if !entityTypes[r.Destination] {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.relationships[%d].destination_entity '%s' must be defined in schema.entities", index, r.Destination))
	}

	return allErrs
}

type EntityEvent struct {
	Context    string   `mapstructure:"context"`
	Conditions []string `mapstructure:"conditions"`
	Entity     string   `mapstructure:"entity"`
	Action     string   `mapstructure:"action"`
}

func (e *EntityEvent) Validate(index int, entityTypes map[string]bool) error {
	var allErrs error

	if e.Action == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.entities[%d].action is mandatory", index))
	} else if e.Action != "update" && e.Action != "delete" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.entities[%d].action must be 'update' or 'delete', got '%s'", index, e.Action))
	}

	if e.Context == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.entities[%d].context is mandatory", index))
	} else if e.Context != ottllog.ContextName && e.Context != ottlmetric.ContextName {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.entities[%d].context must be '%s' or '%s', got '%s'",
			index, ottllog.ContextName, ottlmetric.ContextName, e.Context))
	}

	if e.Entity == "" {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.entities[%d].type is mandatory", index))
	} else if !entityTypes[e.Entity] {
		allErrs = errors.Join(allErrs, fmt.Errorf("schema.events.entities[%d].type '%s' must be defined in schema.entities", index, e.Entity))
	}

	return allErrs
}
