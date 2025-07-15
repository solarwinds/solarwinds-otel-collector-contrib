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
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_config",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name", "k8s.namespace.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty_entities",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.entities is mandatory and must contain at least 1 item",
		},
		{
			name: "entity_missing_type",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.entities[0].entity is mandatory",
		},
		{
			name: "entity_missing_id",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.entities[0].id is mandatory and must contain at least 1 item",
		},
		{
			name: "empty_entity_events",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.entities is mandatory and must contain at least 1 item",
		},
		{
			name: "entity_event_missing_action",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.entities[0].action is mandatory",
		},
		{
			name: "entity_event_invalid_action",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "invalid",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.entities[0].action must be 'update' or 'delete', got 'invalid'",
		},
		{
			name: "entity_event_missing_context",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.entities[0].context is mandatory",
		},
		{
			name: "entity_event_invalid_context",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "invalid",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.entities[0].context must be 'log' or 'metric', got 'invalid'",
		},
		{
			name: "entity_event_missing_type",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.entities[0].type is mandatory",
		},
		{
			name: "entity_event_type_not_in_entities",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "NonExistentType",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.entities[0].type 'NonExistentType' must be defined in schema.entities",
		},
		{
			name: "valid_config_with_relationships",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
						{
							Entity: "KubernetesDeployment",
							IDs:    []string{"k8s.deployment.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
						Relationships: []RelationshipEvent{
							{
								Action:      "delete",
								Context:     "metric",
								Type:        "CommunicatesWith",
								Source:      "KubernetesPod",
								Destination: "KubernetesDeployment",
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "relationship_event_missing_action",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
						Relationships: []RelationshipEvent{
							{
								Action:      "",
								Context:     "metric",
								Type:        "CommunicatesWith",
								Source:      "KubernetesPod",
								Destination: "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.relationships[0].action is mandatory",
		},
		{
			name: "relationship_event_missing_source_entity",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
						Relationships: []RelationshipEvent{
							{
								Action:      "delete",
								Context:     "metric",
								Type:        "CommunicatesWith",
								Source:      "",
								Destination: "KubernetesPod",
							},
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "schema.events.relationships[0].source_entity is mandatory",
		},
		{
			name: "expiration_policy_enabled_missing_interval",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
				Expiration: ExpirationPolicy{
					Enabled:  true,
					Interval: "",
				},
			},
			expectError: true,
			errorMsg:    "expiration_policy.interval is mandatory when expiration_policy.enabled is true",
		},
		{
			name: "expiration_policy_invalid_interval",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
				Expiration: ExpirationPolicy{
					Enabled:  true,
					Interval: "invalid",
				},
			},
			expectError: true,
			errorMsg:    "expiration_policy.interval must be a valid duration",
		},
		{
			name: "valid_config_with_expiration_policy",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "update",
								Context: "log",
								Entity:  "KubernetesPod",
							},
						},
					},
				},
				Expiration: ExpirationPolicy{
					Enabled:  true,
					Interval: "5m",
				},
			},
			expectError: false,
		},
		{
			name: "multiple_validation_errors",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "",         // Missing entity type
							IDs:    []string{}, // Missing IDs
						},
						{
							Entity: "KubernetesPod",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "",                // Missing action
								Context: "invalid_context", // Invalid context
								Entity:  "NonExistentType", // Type not in entities
							},
						},
						Relationships: []RelationshipEvent{
							{
								Action:      "invalid_action",  // Invalid action
								Context:     "",                // Missing context
								Type:        "",                // Missing type
								Source:      "",                // Missing source
								Destination: "NonExistentType", // Type not in entities
							},
						},
					},
				},
				Expiration: ExpirationPolicy{
					Enabled:  true,
					Interval: "invalid_duration", // Invalid duration
				},
			},
			expectError: true,
			errorMsg:    "schema.entities[0].entity is mandatory", // Should contain multiple errors
		},
		{
			name: "multiple_validation_errors_detailed",
			config: &Config{
				Schema: Schema{
					Entities: []Entity{
						{
							Entity: "",         // Missing entity type
							IDs:    []string{}, // Missing IDs
						},
						{
							Entity: "ValidEntity",
							IDs:    []string{"k8s.pod.name"},
						},
					},
					Events: Events{
						Entities: []EntityEvent{
							{
								Action:  "invalid_action",  // Invalid action
								Context: "invalid_context", // Invalid context
								Entity:  "",                // Missing type
							},
							{
								Action:  "", // Missing action
								Context: "log",
								Entity:  "NonExistentType", // Type not in entities
							},
						},
						Relationships: []RelationshipEvent{
							{
								Action:      "", // Missing action
								Context:     "", // Missing context
								Type:        "", // Missing type
								Source:      "", // Missing source
								Destination: "", // Missing destination
							},
						},
					},
				},
				Expiration: ExpirationPolicy{
					Enabled:  true,
					Interval: "invalid_duration", // Invalid duration
				},
			},
			expectError: true,
			errorMsg:    "schema validation failed", // Should contain multiple nested errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
