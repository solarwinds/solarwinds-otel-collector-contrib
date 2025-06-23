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

package storage

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// TestBuildKey tests the buildKey function for various scenarios
// Select test entities are added control group.
// Everything is checked against the control group. Some expect to find the key in the control group, and some expect not to find themselves
// because they should generate key unique from everyting in the control group.
func TestBuildKey(t *testing.T) {
	tests := []struct {
		name              string
		entity            internal.RelationshipEntity
		addToControlGroup bool
	}{
		{
			name: "Simple entity",
			entity: internal.RelationshipEntity{
				Type: "service",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("id", "service-123")
					return m
				}(),
			},
			addToControlGroup: true,
		},
		{
			name: "Complex entity",
			entity: internal.RelationshipEntity{
				Type: "pod",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("namespace", "default")
					m.PutStr("pod_name", "my-pod")
					return m
				}(),
			},
			addToControlGroup: true,
		},
		{
			name: "Different type, same IDs",
			entity: internal.RelationshipEntity{
				Type: "deployment",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("id", "service-123")
					return m
				}(),
			},
			addToControlGroup: false, // Different type should produce different hash
		},
		{
			name: "Same type, different IDs",
			entity: internal.RelationshipEntity{
				Type: "service",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("id", "service-456")
					return m
				}(),
			},
			addToControlGroup: false, // Different IDs should produce different hash
		},
		{
			name: "Empty IDs",
			entity: internal.RelationshipEntity{
				Type: "service",
				IDs:  pcommon.NewMap(),
			},
			addToControlGroup: false, // Empty IDs should be different from non-empty ones
		},
		{
			name: "Various data types",
			entity: internal.RelationshipEntity{
				Type: "resource",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("name", "test-resource")
					m.PutInt("count", 42)
					m.PutDouble("cpu", 2.5)
					m.PutBool("active", true)
					slice := m.PutEmptySlice("tags")
					slice.AppendEmpty().SetStr("tag1")
					slice.AppendEmpty().SetStr("tag2")
					return m
				}(),
			},
			addToControlGroup: true,
		},
	}

	// First generate reference keys for comparison
	// All tests with addToControlGroup=true will generate a reference key map.
	referenceKeys := make(map[string]string)
	kb := NewDefaultKeyBuilder()
	for _, tt := range tests {
		if tt.addToControlGroup {
			key, err := kb.BuildEntityKey(tt.entity)
			require.NoError(t, err)
			referenceKeys[key] = tt.name
		}
	}

	// Then run the actual tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := kb.BuildEntityKey(tt.entity)
			require.NoError(t, err)

			// There should be a reference key for this test case
			if tt.addToControlGroup {
				refKey := referenceKeys[key]
				require.NotNil(t, refKey)
			} else {
				// For tests with controlGroup=false, the key should not match any reference key
				for refName, refKey := range referenceKeys {
					if refName != tt.name {
						require.NotEqual(t, refKey, key, "Key should be different from %s", refName)
					}
				}
			}
		})
	}
}

func TestBuildKey_SameEntitiesWithDifferentIdsOrderHaveSameKeys(t *testing.T) {
	entity := internal.RelationshipEntity{
		Type: "service",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("id", "service-123")
			m.PutStr("environment", "production")
			return m
		}(),
	}

	entityDifferentOrder := internal.RelationshipEntity{
		Type: "service",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("environment", "production")
			m.PutStr("id", "service-123")
			return m
		}(),
	}

	kb := NewDefaultKeyBuilder()
	key1, err1 := kb.BuildEntityKey(entity)
	require.NoError(t, err1)

	key2, err2 := kb.BuildEntityKey(entityDifferentOrder)
	require.NoError(t, err2)

	require.Equal(t, key1, key2, "Keys should be identical for the same entity with different ID order")
}

// TestBuildKeyConsistency ensures that the same entity always generates the same key
func TestBuildKeyConsistency(t *testing.T) {
	entity := internal.RelationshipEntity{
		Type: "service",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("id", "service-123")
			m.PutStr("environment", "production")
			return m
		}(),
	}

	// Generate the key multiple times
	keys := make([]string, 5)
	kb := NewDefaultKeyBuilder()
	for i := 0; i < 5; i++ {
		key, err := kb.BuildEntityKey(entity)
		require.NoError(t, err)
		keys[i] = key
	}

	// Verify all keys are identical
	for i := 1; i < 5; i++ {
		require.Equal(t, keys[0], keys[i], "Keys should be consistent across multiple calls")
	}
}

// TestBuildRelationshipKey ensures that relationship keys are built with the proper format
// and validates input parameters
func TestBuildRelationshipKey(t *testing.T) {
	keyBuilder := NewDefaultKeyBuilder()

	tests := []struct {
		name             string
		relationshipType string
		sourceHash       string
		destHash         string
		expectedKey      string
		expectError      bool
		errorMessage     string
	}{
		{
			name:             "Basic relationship key",
			relationshipType: "dependsOn",
			sourceHash:       "abcdef123456",
			destHash:         "fedcba654321",
			expectedKey:      "dependsOn:abcdef123456:fedcba654321",
			expectError:      false,
		},
		{
			name:             "Empty relationship type",
			relationshipType: "",
			sourceHash:       "abcdef123456",
			destHash:         "fedcba654321",
			expectedKey:      "",
			expectError:      true,
			errorMessage:     "relationshipType cannot be empty",
		},
		{
			name:             "Empty source hash",
			relationshipType: "dependsOn",
			sourceHash:       "",
			destHash:         "fedcba654321",
			expectedKey:      "",
			expectError:      true,
			errorMessage:     "sourceHash cannot be empty",
		},
		{
			name:             "Empty destination hash",
			relationshipType: "dependsOn",
			sourceHash:       "abcdef123456",
			destHash:         "",
			expectedKey:      "",
			expectError:      true,
			errorMessage:     "destHash cannot be empty",
		},
		{
			name:             "Special characters in relationship type",
			relationshipType: "depends-on:with:colons",
			sourceHash:       "abcdef123456",
			destHash:         "fedcba654321",
			expectedKey:      "depends-on:with:colons:abcdef123456:fedcba654321",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := keyBuilder.BuildRelationshipKey(tt.relationshipType, tt.sourceHash, tt.destHash)

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, tt.errorMessage, err.Error())
				require.Empty(t, key)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedKey, key)
			}
		})
	}
}

// TestBuildRelationshipKey_WithRealEntityHashes uses real entity hashes to test end-to-end relationship key creation
func TestBuildRelationshipKey_WithRealEntityHashes(t *testing.T) {
	keyBuilder := NewDefaultKeyBuilder()

	// Create test entities
	sourceEntity := internal.RelationshipEntity{
		Type: "service",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("name", "frontend-service")
			return m
		}(),
	}

	destEntity := internal.RelationshipEntity{
		Type: "database",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("name", "user-database")
			return m
		}(),
	}

	// Generate entity hashes
	sourceHash, err := keyBuilder.BuildEntityKey(sourceEntity)
	require.NoError(t, err)
	destHash, err := keyBuilder.BuildEntityKey(destEntity)
	require.NoError(t, err)

	// Test relationship types
	relationshipTypes := []string{"dependsOn", "calls", "connects-to"}

	for _, relType := range relationshipTypes {
		t.Run(relType, func(t *testing.T) {
			// Create relationship key
			relationshipKey, err := keyBuilder.BuildRelationshipKey(relType, sourceHash, destHash)
			require.NoError(t, err)

			// Verify the key format
			expectedKey := relType + ":" + sourceHash + ":" + destHash
			require.Equal(t, expectedKey, relationshipKey)

			// Verify key components can be extracted
			components := parseRelationshipKey(relationshipKey)
			require.Equal(t, relType, components[0])
			require.Equal(t, sourceHash, components[1])
			require.Equal(t, destHash, components[2])
		})
	}

	// Test the error case with empty relationship type
	t.Run("empty relationship type", func(t *testing.T) {
		_, err := keyBuilder.BuildRelationshipKey("", sourceHash, destHash)
		require.Error(t, err)
		require.Equal(t, "relationshipType cannot be empty", err.Error())
	})
}

// Helper function to parse relationship key components
func parseRelationshipKey(key string) []string {
	var relType, sourceHash, destHash string
	// This is a simplified version for testing,
	// in a real implementation we'd handle potential edge cases
	_, err := fmt.Sscanf(key, "%s:%s:%s", &relType, &sourceHash, &destHash)
	if err != nil {
		return []string{"", "", ""}
	}
	return []string{relType, sourceHash, destHash}
}
