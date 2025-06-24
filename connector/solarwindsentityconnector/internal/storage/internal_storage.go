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
	"errors"
	"fmt"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.uber.org/zap"
	"time"
)

const (
	// ristretto cache provides cost management, but we need to set the same cost for all items.
	// Additionally, when itemCost is set to 1 the MaxCost is equal to the maximum number of items in the cache.
	itemCost = 1
	// When dealing with TTL for entities we do not need to be as precise as with relationships.
	// Thus cleaning up the cache every 10 times the interval.
	entityTTLCleanupFactor = 10
	// entityTTLFactor is a factor that controls how long the entity will be kept in the cache,
	// after the relationship is updated.
	entityTTLFactor = 5
	// bufferItems is a configuration parameter that controls how many key-value operations
	// are buffered before being processed by the internal ristretto eviction logic.
	// The value is recommended by ristretto documentation but not set to any default value.
	bufferItems = 64
)

// InternalCache defines the interface for cache operations used by Manager
type InternalCache interface {
	delete(relationship *internal.Relationship) error
	update(relationship *internal.Relationship) error
	run(ctx context.Context)
}

// internalStorage implements InternalCache interface.
var _ InternalCache = (*internalStorage)(nil)

type storedRelationship struct {
	sourceHash       string
	destHash         string
	relationshipType string
}

type internalStorage struct {
	entities                  *ristretto.Cache[string, internal.RelationshipEntity]
	relationships             *ristretto.Cache[string, storedRelationship]
	ttl                       time.Duration
	ttlCleanUpIntervalSeconds time.Duration
	logger                    *zap.Logger
	keyBuilder                KeyBuilder

	// TODO: Introduce mutex to protect concurrent access to the cache when parallelization is used
	// in the upper layers (NH-112603).
	// mu sync.Mutex
}

func newInternalStorage(cfg *config.ExpirationSettings, logger *zap.Logger, em chan<- internal.Event) (*internalStorage, error) {
	var err error

	// maxCost sets the maximum number of items, when itemCost is set to 1
	maxCost := cfg.MaxCapacity
	// numCounters are internal setting for ristretto cache that helps to manage eviction mechanism.
	// It is recommended to set it to 10 times the maximum capacity for most of the use cases.
	numCounters := cfg.MaxCapacity * 10
	// ttlCleanupSeconds sets the interval for sequence scan of the evicted items. When item is evicted,
	// it is not immediately removed from the cache.
	ttlCleanupSeconds := int64(cfg.TTLCleanupIntervalSeconds.Seconds())
	// One second is the minimum value for ttlCleanupSeconds, as it is used to control the eviction process for the two caches.
	if ttlCleanupSeconds <= 0 {
		return nil, fmt.Errorf("ttlCleanupSeconds has to be bigger than 0")
	}

	if (cfg.Interval * 2) > time.Duration(ttlCleanupSeconds)*time.Second {
		return nil, fmt.Errorf("ttlCleanupSeconds (%s) has to be at minimum twice the value of cfg.Interval (%s)", cfg.Interval, cfg.TTLCleanupIntervalSeconds)
	}

	entityCache, err := ristretto.NewCache(&ristretto.Config[string, internal.RelationshipEntity]{
		NumCounters:            numCounters,
		MaxCost:                maxCost,
		TtlTickerDurationInSec: ttlCleanupSeconds * entityTTLCleanupFactor,
		BufferItems:            bufferItems,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create entity cache: %w", err)
	}

	relationshipCache, err := ristretto.NewCache(&ristretto.Config[string, storedRelationship]{
		NumCounters:            numCounters,
		MaxCost:                maxCost,
		TtlTickerDurationInSec: ttlCleanupSeconds,
		BufferItems:            bufferItems,
		OnEvict: func(item *ristretto.Item[storedRelationship]) {
			onRelationshipEvict(item, entityCache, logger, em)
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create relaitonship cache: %w", err)
	}

	return &internalStorage{
		entities:                  entityCache,
		relationships:             relationshipCache,
		ttl:                       cfg.Interval,
		ttlCleanUpIntervalSeconds: cfg.TTLCleanupIntervalSeconds,
		logger:                    logger,
		keyBuilder:                NewDefaultKeyBuilder(),
	}, nil
}

// onRelationshipEvict is a callback function that is called when a relationship item is evicted from the cache.
// It retrieves the source and destination entities from the entity cache and sends a relationship event.
func onRelationshipEvict(
	item *ristretto.Item[storedRelationship],
	entityCache *ristretto.Cache[string, internal.RelationshipEntity],
	logger *zap.Logger,
	em chan<- internal.Event) {

	logger.Debug("relationship item evicted", zap.String("relationshipType", item.Value.relationshipType))
	source, sourceExists := entityCache.Get(item.Value.sourceHash)
	if !sourceExists {
		logger.Warn("source entity not found in cache", zap.String("hash", item.Value.sourceHash))
		return
	}

	dest, destExists := entityCache.Get(item.Value.destHash)
	if !destExists {
		logger.Warn("destination entity not found in cache", zap.String("hash", item.Value.destHash))
		return
	}

	em <- &internal.Relationship{
		Type: item.Value.relationshipType,
		Source: internal.RelationshipEntity{
			Type: source.Type,
			IDs:  source.IDs,
		},
		Destination: internal.RelationshipEntity{
			Type: dest.Type,
			IDs:  dest.IDs,
		},
	}
}

func (c *internalStorage) run(ctx context.Context) {
	defer func() {
		c.logger.Info("closing ristretto caches")
		c.entities.Close()
		c.relationships.Close()
		c.logger.Info("internalStorage stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

// Delete removes a relationship from the internal storage.
// Entities are not removed, they will be removed when they expire.
func (c *internalStorage) delete(relationship *internal.Relationship) error {
	c.logger.Debug("deleting relationship from internal storage", zap.String("relationshipType", relationship.Type))

	sourceHash, err := c.keyBuilder.BuildEntityKey(relationship.Source)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to hash key for source entity: %s", relationship.Source.Type))
	}
	destHash, err := c.keyBuilder.BuildEntityKey(relationship.Destination)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to hash key for destination entity: %s", relationship.Destination.Type))
	}

	// Remove the relationship from the cache
	relationshipKey, err := c.keyBuilder.BuildRelationshipKey(relationship.Type, sourceHash, destHash)
	if err != nil {
		return err
	}
	c.relationships.Del(relationshipKey)
	return nil
}

// Reset TTL for existing entries, or creates a new entries with default TTL, for given relationship
// as well as source and destination entities.
// Entities have minimum TTL of ttlCleanUpIntervalSeconds * entityTTLFactor, which is minimum 5 seconds as ttlCleanUpIntervalSeconds has 1s minimum.
// Relationships have TTL which can be anything, even milliseconds.
func (c *internalStorage) update(relationship *internal.Relationship) error {
	c.logger.Debug("updating relationship in internal storage", zap.String("relationshipType", relationship.Type))

	sourceHash, err := c.keyBuilder.BuildEntityKey(relationship.Source)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to hash key for source entity: %s", relationship.Source.Type))
	}
	destHash, err := c.keyBuilder.BuildEntityKey(relationship.Destination)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to hash key for destination entity: %s", relationship.Destination.Type))
	}

	sourceUpdated := c.entities.SetWithTTL(sourceHash, relationship.Source, itemCost, c.ttlCleanUpIntervalSeconds*entityTTLFactor)
	if !sourceUpdated {
		return fmt.Errorf("failed to update source entity: %s", relationship.Source.Type)
	}

	destUpdated := c.entities.SetWithTTL(destHash, relationship.Destination, itemCost, c.ttlCleanUpIntervalSeconds*entityTTLFactor)
	if !destUpdated {
		return fmt.Errorf("failed to update destination entity: %s", relationship.Destination.Type)
	}

	relationshipKey, err := c.keyBuilder.BuildRelationshipKey(relationship.Type, sourceHash, destHash)
	if err != nil {
		return err
	}

	relationshipValue := storedRelationship{
		sourceHash:       sourceHash,
		destHash:         destHash,
		relationshipType: relationship.Type,
	}

	relationshipUpdated := c.relationships.SetWithTTL(relationshipKey, relationshipValue, itemCost, c.ttl)
	if !relationshipUpdated {
		return fmt.Errorf("failed to update relationship: %s", relationship.Type)
	}
	return nil
}
