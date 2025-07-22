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

func TestConfig_EvaluateWarnings_NoEntitiesInSchema(t *testing.T) {
	cfg := &Config{
		Schema: Schema{
			Events: Events{
				Entities: []EntityEvent{{}},
			},
		},
	}

	warnings := cfg.EvaluateWarnings()

	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings, "no entities defined in schema::entities, there's nothing to do")
}

func TestConfig_EvaluateWarnings_NoEventsInSchema(t *testing.T) {
	cfg := &Config{
		Schema: Schema{
			Entities: []Entity{{}},
		},
	}

	warnings := cfg.EvaluateWarnings()

	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings, "no events defined in schema::events, there's nothing to do")
}

func TestConfig_EvaluateWarnings_ValidSchemaWithEntityEvents(t *testing.T) {
	cfg := &Config{
		Schema: Schema{
			Entities: []Entity{{}},
			Events: Events{
				Entities: []EntityEvent{{}},
			},
		},
	}

	warnings := cfg.EvaluateWarnings()

	assert.Empty(t, warnings)
}

func TestConfig_EvaluateWarnings_ValidSchemaWithRelationshipEvents(t *testing.T) {
	cfg := &Config{
		Schema: Schema{
			Entities: []Entity{{}},
			Events: Events{
				Relationships: []RelationshipEvent{{}},
			},
		},
	}

	warnings := cfg.EvaluateWarnings()

	assert.Empty(t, warnings)
}
