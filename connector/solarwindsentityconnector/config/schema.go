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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/xconfmap"
)

type Schema struct {
	Entities []Entity `mapstructure:"entities"`
	Events   Events   `mapstructure:"events"`
}

// By implementing the xconfmap.Validator, we ensure it's validated by the collector automatically
var _ xconfmap.Validator = (*Schema)(nil)

func (s *Schema) Validate() error {
	var errs error

	entityTypes := make(map[string]struct{})
	for _, entity := range s.Entities {
		entityTypes[entity.Entity] = struct{}{}
	}

	for i, e := range s.Events.Entities {
		if _, exists := entityTypes[e.Entity]; !exists {
			errs = errors.Join(errs, fmt.Errorf("the '%s' name for events::entities::%d::entity must be defined in 'entities'", e.Entity, i))
		}
	}

	for i, r := range s.Events.Relationships {
		if _, exists := entityTypes[r.Source]; !exists {
			errs = errors.Join(errs, fmt.Errorf("the '%s' name for events::relationships::%d::source_entity must be defined in 'entities'", r.Source, i))
		}

		if _, exists := entityTypes[r.Destination]; !exists {
			errs = errors.Join(errs, fmt.Errorf("the '%s' name for events::relationships::%d::destination_entity must be defined in 'entities'", r.Destination, i))
		}
	}

	return errs
}

type ParsedSchema struct {
	Entities map[string]Entity
	Events   ParsedEvents
}

func (s *Schema) Unmarshal(settings component.TelemetrySettings) (*ParsedSchema, error) {
	events, err := createParsedEvents(*s, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create events: %w", err)
	}

	return &ParsedSchema{
		Entities: s.newEntities(),
		Events:   events,
	}, nil
}

func (s *Schema) newEntities() map[string]Entity {
	entities := make(map[string]Entity, len(s.Entities))
	for _, entity := range s.Entities {
		entities[entity.Entity] = entity
	}

	return entities
}
