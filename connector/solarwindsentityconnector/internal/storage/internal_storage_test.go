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
	"context"
	"fmt"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
	"testing"
	"time"
)

var sourceEntity = internal.RelationshipEntity{
	Type: "service",
	IDs: func() pcommon.Map {
		m := pcommon.NewMap()
		m.PutStr("name", "frontend")
		return m
	}(),
}

var destEntity = internal.RelationshipEntity{
	Type: "database",
	IDs: func() pcommon.Map {
		m := pcommon.NewMap()
		m.PutStr("name", "userdb")
		return m
	}(),
}

var relationship = &internal.Relationship{
	Type:        "dependsOn",
	Source:      sourceEntity,
	Destination: destEntity,
}

func TestNewInternalStorage(t *testing.T) {
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)

	tests := []struct {
		name        string
		cfg         *config.ExpirationSettings
		expectError bool
	}{
		{
			name: "Valid configuration",
			cfg: &config.ExpirationSettings{
				Interval:           10 * time.Second,
				MaxCapacity:        1000,
				TTLCleanupInterval: 1 * time.Second,
			},
			expectError: false,
		},
		{
			name: "Zero MaxCapacity",
			cfg: &config.ExpirationSettings{
				Interval:           10 * time.Second,
				MaxCapacity:        0,
				TTLCleanupInterval: 1 * time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := newInternalStorage(tt.cfg, logger, eventsChan)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, storage)
			} else {
				require.NoError(t, err)
				require.NotNil(t, storage)
				require.Equal(t, tt.cfg.Interval, storage.ttl)
				require.NotNil(t, storage.entities)
				require.NotNil(t, storage.relationships)
			}
		})
	}
}

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
	for _, tt := range tests {
		if tt.addToControlGroup {
			key, err := buildKey(tt.entity)
			require.NoError(t, err)
			referenceKeys[key] = tt.name
		}
	}

	// Then run the actual tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := buildKey(tt.entity)
			require.NoError(t, err)

			// There should be a refference key for this test case
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

func TestBuildKeySameEntityDifferentIdsOrderHaveSameKeys(t *testing.T) {
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

	key1, err1 := buildKey(entity)
	require.NoError(t, err1)

	key2, err2 := buildKey(entityDifferentOrder)
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
	for i := 0; i < 5; i++ {
		key, err := buildKey(entity)
		require.NoError(t, err)
		keys[i] = key
	}

	// Verify all keys are identical
	for i := 1; i < 5; i++ {
		require.Equal(t, keys[0], keys[i], "Keys should be consistent across multiple calls")
	}
}

func TestUpdate(t *testing.T) {
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)
	cfg := &config.ExpirationSettings{
		Interval:           1 * time.Second,
		MaxCapacity:        100000,
		TTLCleanupInterval: 100 * time.Millisecond,
	}

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	// Calculate the source and destination hashes
	sourceHash, err := buildKey(relationship.Source)
	require.NoError(t, err)
	destHash, err := buildKey(relationship.Destination)
	require.NoError(t, err)

	// First update should succeed
	err = storage.update(relationship)
	require.NoError(t, err)

	// Ensure entities are in cache
	storage.entities.Wait()
	storage.relationships.Wait()

	// Simulate some time passing
	time.Sleep(200 * time.Millisecond)

	// Get the TTLs after first update
	srcTtl1, found := storage.entities.GetTTL(sourceHash)
	require.True(t, found, "Source entity should be in cache")
	destTtl1, found := storage.entities.GetTTL(destHash)
	require.True(t, found, "Destination entity should be in cache")

	// Update with the same data should also succeed
	// (this tests the TTL refresh logic)
	err = storage.update(relationship)
	require.NoError(t, err)

	// Ensure entities are updated in cache
	storage.entities.Wait()
	storage.relationships.Wait()

	// Get the TTLs after second update
	sourceTtl2, found := storage.entities.GetTTL(sourceHash)
	require.True(t, found, "Source entity should still be in cache")
	destTtl2, found := storage.entities.GetTTL(destHash)
	require.True(t, found, "Destination entity should still be in cache")

	// Verify that the TTLs were updated (should be higher after refresh)
	require.Greater(t, sourceTtl2, srcTtl1,
		"Source TTL should be refreshed after update")
	require.Greater(t, destTtl2, destTtl1,
		"Destination TTL should be refreshed after update")
}

func TestTtlExpiration(t *testing.T) {
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)

	// Create the source and destination hashes
	sourceHash, err := buildKey(sourceEntity)
	require.NoError(t, err)
	destHash, err := buildKey(destEntity)
	require.NoError(t, err)

	// Set up a cache with the entities
	ttlTime := 100 * time.Millisecond
	cfg := &config.ExpirationSettings{
		Enabled:            true,
		Interval:           ttlTime,
		MaxCapacity:        1000000,
		TTLCleanupInterval: ttlTime / 10,
	}

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	// Run the storage in a goroutine
	done := make(chan struct{})
	go func() {
		storage.run(ctx)
		close(done)
	}()

	relationship := &internal.Relationship{
		Type:        "dependsOn",
		Source:      sourceEntity,
		Destination: destEntity,
	}

	// Update the relationship in storage
	err = storage.update(relationship)
	storage.entities.Wait()
	storage.relationships.Wait()
	require.NoError(t, err, "Failed to insert relationship")

	// Verify entities are in the cache before TTL expiration
	_, srcFound := storage.entities.Get(sourceHash)
	_, dstFound := storage.entities.Get(destHash)
	require.True(t, srcFound, "Source entity should be in cache")
	require.True(t, dstFound, "Destination entity should be in cache")

	time.Sleep(2 * ttlTime) // Wait for TTL to expire
	// Check that an event was sent
	select {
	case event := <-eventsChan:
		rel, ok := event.(*internal.Relationship)
		require.True(t, ok, "Event should be a Relationship")
		assert.Equal(t, "dependsOn", rel.Type)
		assert.Equal(t, "service", rel.Source.Type)
		assert.Equal(t, "database", rel.Destination.Type)

		relationshipKey := fmt.Sprintf("%s:%s:%s", relationship.Type, sourceHash, destHash)
		_, relFound := storage.relationships.Get(relationshipKey)
		storage.relationships.Wait()
		time.Sleep(100 * time.Millisecond) // Wait for eviction to happen
		require.False(t, relFound, "Relationship should be gone from cache")

		// Verify entities are gone after eviction
		_, srcFound = storage.entities.Get(sourceHash)
		_, dstFound = storage.entities.Get(destHash)
		require.True(t, srcFound, "Source entity should be in cache for a little longer")
		require.True(t, dstFound, "Destination entity should be in cache for a little longer")

		// Check the IDs
		nameVal, exists := rel.Source.IDs.Get("name")
		require.True(t, exists)
		assert.Equal(t, "frontend", nameVal.AsString())

		nameVal, exists = rel.Destination.IDs.Get("name")
		require.True(t, exists)
		assert.Equal(t, "userdb", nameVal.AsString())
		cancel()
	case <-time.After(10 * ttlTime):
		relationshipKey := fmt.Sprintf("%s:%s:%s", relationship.Type, sourceHash, destHash)
		_, relFound := storage.relationships.Get(relationshipKey)
		assert.False(t, relFound, "Relationship should be gone from cache")
		cancel()
		t.Fatal("No event was received, but relationship has gone from cache")
	}

}

func TestRunAndShutdown(t *testing.T) {
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)
	cfg := &config.ExpirationSettings{
		Interval:           1 * time.Second,
		MaxCapacity:        100,
		TTLCleanupInterval: 100 * time.Millisecond,
	}

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Run the storage in a goroutine
	done := make(chan struct{})
	go func() {
		storage.run(ctx)
		close(done)
	}()

	// Cancel should terminate the run loop
	cancel()

	// Wait for the goroutine to exit
	select {
	case <-done:
		// Success - the run method exited
	case <-time.After(1 * time.Second):
		t.Fatal("Storage run did not exit after context cancellation")
	}
}
