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

type Schema struct {
	Entities []Entity `mapstructure:"entities"`
	Events   Events   `mapstructure:"events"`
}

func (s *Schema) NewEntities() map[string]Entity {
	entities := make(map[string]Entity, len(s.Entities))
	for _, entity := range s.Entities {
		entities[entity.Type] = entity
	}

	return entities
}

func (s *Schema) NewRelationships() []Relationship {
	return s.Events.Relationships
}
