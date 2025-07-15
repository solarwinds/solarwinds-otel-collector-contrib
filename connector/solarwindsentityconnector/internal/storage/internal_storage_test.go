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
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
	"testing"
	"time"
)

var ttl = 1000 * time.Millisecond
var ttlCleanupInterval = 2 * time.Second
var cfg = &config.ExpirationSettings{
	Enabled:                   true,
	Interval:                  ttl,
	MaxCapacity:               1000,
	TTLCleanupIntervalSeconds: ttlCleanupInterval,
}

var sourceEntity = internal.Entity{
	Type: "service",
	IDs: func() pcommon.Map {
		m := pcommon.NewMap()
		m.PutStr("name", "frontend")
		return m
	}(),
}

var destEntity = internal.Entity{
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
				Interval:                  10 * time.Second,
				MaxCapacity:               1000,
				TTLCleanupIntervalSeconds: 20 * time.Second,
			},
			expectError: false,
		},
		{
			name: "Zero MaxCapacity throws error",
			cfg: &config.ExpirationSettings{
				Interval:                  10 * time.Second,
				MaxCapacity:               0,
				TTLCleanupIntervalSeconds: 1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "TTl is less than 1 second, throws error",
			cfg: &config.ExpirationSettings{
				Interval:                  1 * time.Millisecond,
				MaxCapacity:               1000,
				TTLCleanupIntervalSeconds: 10 * time.Millisecond,
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
				require.Equal(t, tt.cfg.Interval, storage.relationshipTtl)
				require.NotNil(t, storage.entities)
				require.NotNil(t, storage.relationships)
			}
		})
	}
}

func TestUpdate_RelationshipUpdate_UpdatesTtl(t *testing.T) {
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	// First update should succeed
	err = storage.update(relationship)
	require.NoError(t, err)
	// Ensure entities are in cache
	storage.entities.Wait()
	storage.relationships.Wait()
	time.Sleep(200 * time.Millisecond)

	// Calculate the source and destination hashes
	sourceHash, err := storage.cacheKeyBuilder.BuildEntityKey(relationship.Source)
	require.NoError(t, err)
	destHash, err := storage.cacheKeyBuilder.BuildEntityKey(relationship.Destination)
	require.NoError(t, err)
	relationshipKey := fmt.Sprintf("%s:%s:%s", relationship.Type, sourceHash, destHash)

	// Get the TTLs after first update
	srcTtl1, found := storage.entities.GetTTL(sourceHash)
	require.True(t, found, "Source entity should be in cache")
	destTtl1, found := storage.entities.GetTTL(destHash)
	require.True(t, found, "Destination entity should be in cache")
	relTtl1, found := storage.relationships.GetTTL(relationshipKey)
	require.True(t, found, "Relationship should be in cache")

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
	relTtl2, found := storage.relationships.GetTTL(relationshipKey)
	require.True(t, found, "Relationship should still be in cache")

	// Verify that the TTLs were updated (should be higher after refresh)
	require.Greater(t, sourceTtl2, srcTtl1,
		"Source TTL should be refreshed after update")
	require.Greater(t, destTtl2, destTtl1,
		"Destination TTL should be refreshed after update")
	require.Greater(t, relTtl2, relTtl1,
		"Relationship TTL should be refreshed after update")
}

func TestTtlExpiration_TenIdenticalUpdates_ResultInOneExpiryEvent(t *testing.T) {
	logger := zap.NewNop()
	// Increase channel buffer to handle 10 events
	eventsChan := make(chan internal.Event, 20)

	// Set up a cache with the entities
	ttl := 1000 * time.Millisecond
	ttlCleanupInterval := 2 * time.Second
	cfg := &config.ExpirationSettings{
		Enabled:                   true,
		Interval:                  ttl,
		MaxCapacity:               1000,
		TTLCleanupIntervalSeconds: ttlCleanupInterval,
	}

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	relationship := &internal.Relationship{
		Type:        "dependsOn",
		Source:      sourceEntity,
		Destination: destEntity,
	}

	insertTime := time.Now()
	t.Logf("Inserting 10 identical relationships at %v", insertTime.Format(time.RFC3339Nano))

	// Send 10 updates
	for i := 0; i < 10; i++ {
		err = storage.update(relationship)
		t.Logf("Inserting identical relationships at %v", insertTime.Format(time.RFC3339Nano))
		time.Sleep(10 * time.Millisecond) // Sleep to simulate time passing between updates
		storage.entities.Wait()
		storage.relationships.Wait()
		require.NoError(t, err, "Failed to insert relationship %d", i)
	}

	// Wait for 10 events
	eventsReceived := 0
	maxWait := ttlCleanupInterval * 10
	deadline := time.After(maxWait)
	select {
	case event := <-eventsChan:
		eventTime := time.Now()
		delta := eventTime.Sub(insertTime)
		t.Logf("Event %d arrived at %v (delta: %v)", eventsReceived+1, eventTime.Format(time.RFC3339Nano), delta)

		rel, ok := event.(*internal.Relationship)
		require.True(t, ok, "Event should be a Relationship")
		assert.Equal(t, "dependsOn", rel.Type)
		assert.Equal(t, "service", rel.Source.Type)
		assert.Equal(t, "database", rel.Destination.Type)

		nameVal, exists := rel.Source.IDs.Get("name")
		require.True(t, exists)
		assert.Equal(t, "frontend", nameVal.AsString())

		nameVal, exists = rel.Destination.IDs.Get("name")
		require.True(t, exists)
		assert.Equal(t, "userdb", nameVal.AsString())

		eventsReceived++
	case <-deadline:
		assert.Equal(t, eventsReceived, 1, "Only one event should be received, even though 10 updates were sent")
	}
}

func TestTtlExpiration_TenDifferentUpdates_ResultInTenExpiryEvents(t *testing.T) {
	logger := zap.NewNop()
	// Increase channel buffer to handle 10 events
	eventsChan := make(chan internal.Event, 20)

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	// Create and insert 10 different relationships
	insertTime := time.Now()
	t.Logf("Inserting 10 different relationships at %v", insertTime.Format(time.RFC3339Nano))

	for i := 0; i < 10; i++ {
		// Create a unique relationship by varying the source entity's ID
		sourceEntity := internal.Entity{
			Type: "service",
			IDs: func() pcommon.Map {
				m := pcommon.NewMap()
				m.PutStr("name", fmt.Sprintf("frontend-%d", i))
				return m
			}(),
		}

		relationship := &internal.Relationship{
			Type:        "dependsOn",
			Source:      sourceEntity,
			Destination: destEntity, // All relationships share the same destination
		}

		err = storage.update(relationship)
		time.Sleep(10 * time.Millisecond) // Sleep to simulate time passing between updates
		storage.entities.Wait()
		storage.relationships.Wait()
		require.NoError(t, err, "Failed to insert relationship %d", i)
		t.Logf("Relationship %d inserted at %v", i+1, time.Now().Format(time.RFC3339Nano))
	}

	// Wait for and verify all 10 events
	eventsReceived := 0
	maxWait := ttlCleanupInterval * 10
	deadline := time.After(maxWait)

	for eventsReceived < 10 {
		select {
		case event := <-eventsChan:
			eventTime := time.Now()
			delta := eventTime.Sub(insertTime)
			t.Logf("Event %d arrived at %v (delta: %v from insertion)",
				eventsReceived+1, eventTime.Format(time.RFC3339Nano), delta)

			rel, ok := event.(*internal.Relationship)
			require.True(t, ok, "Event should be a Relationship")
			assert.Equal(t, "dependsOn", rel.Type)
			assert.Equal(t, "service", rel.Source.Type)
			assert.Equal(t, "database", rel.Destination.Type)

			// Source name should match one of our generated frontends
			nameVal, exists := rel.Source.IDs.Get("name")
			require.True(t, exists)
			assert.Contains(t, nameVal.AsString(), "frontend-", "Source name should contain 'frontend-'")

			nameVal, exists = rel.Destination.IDs.Get("name")
			require.True(t, exists)
			assert.Equal(t, "userdb", nameVal.AsString())

			eventsReceived++

		case <-deadline:
			t.Fatalf("Timed out waiting for events: received %d of 10 expected events", eventsReceived)
			return
		}
	}
	assert.Equal(t, 10, eventsReceived, "Should receive all 10 events for 10 different relationships")
}

func TestTtlExpiration_RelationshipIsRemovedFirst_EntitiesSecond(t *testing.T) {
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	err = storage.update(relationship)
	require.NoError(t, err)
	storage.entities.Wait()
	storage.relationships.Wait()

	kb := internal.NewKeyBuilder()
	sourceHash, err := kb.BuildEntityKey(relationship.Source)
	require.NoError(t, err)
	destHash, err := kb.BuildEntityKey(relationship.Destination)
	require.NoError(t, err)
	relationshipKey := fmt.Sprintf("%s:%s:%s", relationship.Type, sourceHash, destHash)

	// Wait for the expiry event (relationship should expire first)
	select {
	case <-eventsChan:
		// After event, relationship should be gone, but entities should still exist
		t.Logf("Received expiry event for relationship")
		_, found := storage.relationships.Get(relationshipKey)
		assert.False(t, found, "Relationship should be removed after expiry event")
		t.Logf("relationship is gone from cache")
		_, found = storage.entities.Get(sourceHash)
		assert.True(t, found, "Source entity should still be present after relationship expiry")
		_, found = storage.entities.Get(destHash)
		assert.True(t, found, "Destination entity should still be present after relationship expiry")
		t.Logf("entities are still in cache")
	case <-time.After(10 * time.Second):
		t.Fatal("Timed out waiting for relationship expiry event")
	}

	maxWait := 30 * ttlCleanupInterval

	// Set up ticker for polling and deadline for timeout
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(maxWait)

	t.Logf("Waiting for entities to expire...")
	// Use a single select for both deadline and polling checks
	for {
		select {
		case <-ticker.C:
			// Check entity status on each tick
			_, foundSrc := storage.entities.Get(sourceHash)
			_, foundDst := storage.entities.Get(destHash)

			if !foundSrc && !foundDst {
				t.Logf("Both entities have been removed from cache")
				return
			}

			// Log progress details
			t.Logf("Entities still in cache: source=%v, destination=%v",
				foundSrc, foundDst)

		case <-deadline:
			t.Fatalf("Entities not expired after waiting %v", maxWait)
		}
	}
}

func TestInternalStorage_Delete(t *testing.T) {
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)

	tests := []struct {
		name         string
		relationship *internal.Relationship
		setupCache   func(*internalStorage)
		expectError  bool
	}{
		{
			name:         "Successfully delete relationship",
			relationship: relationship,
			setupCache: func(s *internalStorage) {
				// First add the relationship to the cache
				err := s.update(relationship)
				require.NoError(t, err)

				// Verify it exists in the cache
				sourceHash, _ := s.cacheKeyBuilder.BuildEntityKey(sourceEntity)
				destHash, _ := s.cacheKeyBuilder.BuildEntityKey(destEntity)
				relationshipKey := fmt.Sprintf("%s:%s:%s", relationship.Type, sourceHash, destHash)
				s.entities.Wait()
				s.relationships.Wait()

				_, found := s.relationships.Get(relationshipKey)
				require.True(t, found, "relationship not found in cache before deletion")
			},
			expectError: false,
		},
		{
			name: "Error building relationship hash key",
			relationship: &internal.Relationship{
				Type: "", // Empty type should cause an error when building the key
				Source: internal.Entity{
					Type: "service",
					IDs:  pcommon.NewMap(),
				},
				Destination: destEntity,
			},
			setupCache:  func(s *internalStorage) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new storage instance for each test case
			storage, err := newInternalStorage(cfg, logger, eventsChan)
			require.NoError(t, err)
			defer storage.entities.Close()
			defer storage.relationships.Close()

			// Setup cache with test data if needed
			tt.setupCache(storage)

			// Execute the delete operation
			err = storage.delete(tt.relationship)
			storage.entities.Wait()
			storage.relationships.Wait()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDelete_DoesNotTriggerOnRelationshipEvict(t *testing.T) {
	t.Skip("this test must be run manually only")
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)

	// Create test storage with short TTL for quick testing
	ttl := 60 * time.Second
	cleanupInterval := 2 * time.Second
	cfg := &config.ExpirationSettings{
		Enabled:                   true,
		Interval:                  ttl,
		MaxCapacity:               1000,
		TTLCleanupIntervalSeconds: cleanupInterval,
	}

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	// First, add the relationship to the cache
	err = storage.update(relationship)
	require.NoError(t, err)
	storage.entities.Wait()
	storage.relationships.Wait()

	// Get the keys to verify presence in cache
	sourceHash, err := storage.cacheKeyBuilder.BuildEntityKey(relationship.Source)
	require.NoError(t, err)
	destHash, err := storage.cacheKeyBuilder.BuildEntityKey(relationship.Destination)
	require.NoError(t, err)
	relationshipKey, err := storage.cacheKeyBuilder.BuildRelationshipKey(relationship.Type, sourceHash, destHash)
	require.NoError(t, err)

	// Verify relationship is in cache
	_, found := storage.relationships.Get(relationshipKey)
	require.True(t, found, "Relationship should be in cache before deletion")

	// Now delete the relationship
	err = storage.delete(relationship)
	require.NoError(t, err)
	storage.relationships.Wait()

	// Verify relationship is no longer in cache
	_, found = storage.relationships.Get(relationshipKey)
	assert.False(t, found, "Relationship should no longer be in cache after deletion")

	// Check that no event was sent to the channel (meaning onRelationshipEvict wasn't called)
	select {
	case event := <-eventsChan:
		t.Fatalf("Unexpected event received: %+v", event)
	case <-time.After(10 * time.Second):
		// This is good - no event was received
	}

	// Verify entities are still in cache (delete only removes relationships, not entities)
	_, foundSource := storage.entities.Get(sourceHash)
	assert.True(t, foundSource, "Source entity should still be in cache after relationship deletion")
	_, foundDest := storage.entities.Get(destHash)
	assert.True(t, foundDest, "Destination entity should still be in cache after relationship deletion")
}

// Ttl expiration should trigger an event, but deletion should not.
func TestDelete_ComparedToTTLExpiration(t *testing.T) {
	logger := zap.NewNop()
	eventsChan := make(chan internal.Event, 10)

	// Create test storage with very short TTL for quick testing
	ttl := 1 * time.Second
	cleanupInterval := 2 * time.Second
	cfg := &config.ExpirationSettings{
		Enabled:                   true,
		Interval:                  ttl,
		MaxCapacity:               1000,
		TTLCleanupIntervalSeconds: cleanupInterval,
	}

	storage, err := newInternalStorage(cfg, logger, eventsChan)
	require.NoError(t, err)

	// Create two identical relationships - one to delete manually, one to expire via TTL
	createTestRelationship := func(suffix string) *internal.Relationship {
		return &internal.Relationship{
			Type: "testRelation" + suffix,
			Source: internal.Entity{
				Type: "sourceType",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("id", "source-"+suffix)
					return m
				}(),
			},
			Destination: internal.Entity{
				Type: "destType",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("id", "dest-"+suffix)
					return m
				}(),
			},
		}
	}

	relToDelete := createTestRelationship("delete")
	relToExpire := createTestRelationship("expire")

	// Add both relationships to the cache
	err = storage.update(relToDelete)
	require.NoError(t, err)
	err = storage.update(relToExpire)
	require.NoError(t, err)
	storage.entities.Wait()
	storage.relationships.Wait()

	// Delete one relationship manually
	err = storage.delete(relToDelete)
	require.NoError(t, err)
	storage.relationships.Wait()

	// Wait for the other relationship to expire via TTL
	waitTime := 10 * time.Second
	t.Logf("Waiting %v for TTL expiration...", waitTime)

	// Use a select to wait for eviction event
	select {
	case event := <-eventsChan:
		// We should receive exactly one event from the TTL expired relationship
		rel, ok := event.(*internal.Relationship)
		require.True(t, ok, "Event should be a Relationship")

		// Verify it's the expired relationship, not the deleted one
		sourceIDVal, exists := rel.Source.IDs.Get("id")
		require.True(t, exists)
		assert.Equal(t, "source-expire", sourceIDVal.AsString(),
			"Should receive event for the expired relationship, not the deleted one")

		// Check no more events
		time.Sleep(500 * time.Millisecond)
		assert.Equal(t, 0, len(eventsChan), "Should not receive additional events")

	case <-time.After(waitTime):
		t.Fatal("Timed out waiting for TTL expiration event")
	}
}
