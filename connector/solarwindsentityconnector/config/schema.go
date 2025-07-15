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

	"go.opentelemetry.io/collector/component"
)

type Schema struct {
	Entities []Entity `mapstructure:"entities"`
	Events   Events   `mapstructure:"events"`
}

// Validate validates the schema section of the configuration
func (s *Schema) Validate() error {
	// Schema is mandatory
	if len(s.Entities) == 0 {
		return errors.New("schema.entities is mandatory and must contain at least 1 item")
	}

	// Validate entities
	for i, entity := range s.Entities {
		if err := entity.Validate(i); err != nil {
			return err
		}
	}

	// Events section is mandatory
	if err := s.Events.Validate(s.Entities); err != nil {
		return err
	}

	return nil
}

func (s *Schema) NewEntities() map[string]Entity {
	entities := make(map[string]Entity, len(s.Entities))
	for _, entity := range s.Entities {
		entities[entity.Type] = entity
	}

	return entities
}

func (s *Schema) NewEvents(settings component.TelemetrySettings) ParsedEvents {
	NewEvents := CreateParsedEvents(*s, settings)
	return NewEvents
}
