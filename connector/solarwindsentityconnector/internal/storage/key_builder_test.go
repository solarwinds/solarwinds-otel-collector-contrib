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
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/stretchr/testify/require"
)

// TestBuildKey_DifferentTypesWithSameAttributes_HaveDifferentHashes tests that entities with the same
// attributes but different types produce different hash keys
func TestBuildKey_DifferentTypesWithSameAttributes_HaveDifferentHashes(t *testing.T) {
	entity1 := internal.RelationshipEntity{
		Type: "service",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("id", "service-123")
			return m
		}(),
	}

	entity2 := internal.RelationshipEntity{
		Type: "deployment",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("id", "service-123")
			return m
		}(),
	}
	kb := NewKeyBuilder()
	key1, err := kb.BuildEntityKey(entity1)
	require.NoError(t, err)

	key2, err := kb.BuildEntityKey(entity2)
	require.NoError(t, err)

	require.NotEqual(t, key1, key2, "Entities with different types but same attributes should have different keys")
}

func TestBuildKey_EmptyIds_HaveSameHashes(t *testing.T) {
	entityWithIds := internal.RelationshipEntity{
		Type: "service",
		IDs:  pcommon.NewMap(),
	}

	entityWithEmptyIds := internal.RelationshipEntity{
		Type: "service",
		IDs:  pcommon.NewMap(),
	}

	kb := NewKeyBuilder()
	keyWithIds, err := kb.BuildEntityKey(entityWithIds)
	require.NoError(t, err)

	keyWithEmptyIds, err := kb.BuildEntityKey(entityWithEmptyIds)
	require.NoError(t, err)

	require.Equal(t, keyWithIds, keyWithEmptyIds, "Hashes should be the same for entities with empty IDs")
}

// TestBuildKey_SameIdsButOfDifferentTypes_ResultInDifferentHashes tests that the same ID values
// but of different data types result in different hash keys
func TestBuildKey_SameIdsButOfDifferentTypes_ResultInDifferentHashes(t *testing.T) {
	entityWithStringId := internal.RelationshipEntity{
		Type: "resource",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("value", "42")
			return m
		}(),
	}

	entityWithIntId := internal.RelationshipEntity{
		Type: "resource",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutInt("value", 42)
			return m
		}(),
	}

	kb := NewKeyBuilder()
	keyWithStringId, err := kb.BuildEntityKey(entityWithStringId)
	require.NoError(t, err)

	keyWithIntId, err := kb.BuildEntityKey(entityWithIntId)
	require.NoError(t, err)

	require.NotEqual(t, keyWithStringId, keyWithIntId, "Same ID values but of different types should result in different hash keys")
}

func TestBuildKey_SameTypesWithDifferentlyOrderedIds_HaveSameKeys(t *testing.T) {
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

	kb := NewKeyBuilder()
	key1, err1 := kb.BuildEntityKey(entity)
	require.NoError(t, err1)

	key2, err2 := kb.BuildEntityKey(entityDifferentOrder)
	require.NoError(t, err2)

	require.Equal(t, key1, key2, "Keys should be identical for the same entity with different ID order")
}

// TestBuildKey_Consistency ensures that the same entity always generates the same key
func TestBuildKey_Consistency(t *testing.T) {
	entity := internal.RelationshipEntity{
		Type: "service",
		IDs: func() pcommon.Map {
			m := pcommon.NewMap()
			m.PutStr("id", "service-123")
			m.PutStr("environment", "production")
			return m
		}(),
	}

	kb := NewKeyBuilder()
	// Generate the key multiple times
	keys := make([]string, 5)
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
	keyBuilder := NewKeyBuilder()

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
