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

var _ xconfmap.Validator = &Schema{}

func (s *Schema) Validate() error {
	var allErrs error

	for i, entity := range s.Entities {
		if err := entity.Validate(i); err != nil {
			allErrs = errors.Join(allErrs, err)
		}
	}

	if err := s.Events.Validate(s.Entities); err != nil {
		allErrs = errors.Join(allErrs, err)
	}

	if allErrs != nil {
		return fmt.Errorf("schema validation failed: %w", allErrs)
	}

	return nil
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
