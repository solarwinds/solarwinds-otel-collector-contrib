package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
	"time"
)

const (
	// ristretto cache provides cost management, but we need to set the same cost for all items.
	// Additionally, the MaxCost is then equal to the maximum number of items in the cache.
	itemCost = 1
	// When dealing with TTL for entities we do not need to be as precise as with relationships.
	// Thus cleaning up the cache every 10 times the interval is sufficient.
	entityExpirationFactor = 10
	// The value is recommended by ristretto documentation but not set to any default value.
	bufferItems = 64
)

type storedRelationship struct {
	sourceHash       string
	destHash         string
	relationshipType string
}

type storedEntity struct {
	Type string
	IDs  pcommon.Map
}

type internalStorage struct {
	entities      *ristretto.Cache[string, storedEntity]
	relationships *ristretto.Cache[string, storedRelationship]
	ttl           time.Duration
	interval      time.Duration
	logger        *zap.Logger

	// TODO: Introduce mutex to protect concurrent access to the cache when parallelization is used
	// in the upper layers (NH-112603).
	// mu sync.Mutex
}

func newInternalStorage(cfg *config.ExpirationSettings, logger *zap.Logger, em chan<- internal.Subject) *internalStorage {
	var err error

	entityCache, err := ristretto.NewCache(&ristretto.Config[string, storedEntity]{
		NumCounters: cfg.MaxCapacity * 10,
		MaxCost:     cfg.MaxCapacity,
		// TtlTickerDurationInSec is set to 10 times the cleanup interval to ensure that entities are cleaned up
		// in a timely manner, but not too frequently.
		TtlTickerDurationInSec: int64(cfg.TTLCleanupInterval.Seconds()) * entityExpirationFactor,
		BufferItems:            bufferItems,
	})

	relationshipCache, err := ristretto.NewCache(&ristretto.Config[string, storedRelationship]{
		NumCounters:            cfg.MaxCapacity * 10,
		MaxCost:                cfg.MaxCapacity,
		TtlTickerDurationInSec: int64(cfg.TTLCleanupInterval.Seconds()),
		BufferItems:            bufferItems,
		OnEvict: func(item *ristretto.Item[storedRelationship]) {
			logger.Info("internalStorage relationship evicted")

			if item != nil {
				source, sourceExists := entityCache.Get(item.Value.sourceHash)
				if !sourceExists {
					logger.Warn("source entity not found in cache", zap.String("hash", item.Value.sourceHash))
					return
				}

				dest, destExists := entityCache.Get(item.Value.destHash)
				if !destExists {
					logger.Warn("destination entity not found in cache", zap.String("hash", item.Value.sourceHash))
				}

				em <- internal.Relationship{
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
		},
	})

	if err != nil {
		panic("Failed to create cache: " + err.Error())
	}

	return &internalStorage{
		entities:      entityCache,
		relationships: relationshipCache,
		ttl:           cfg.Interval,
		logger:        logger,
	}
}

func (c *internalStorage) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("internalStorage stopped")
			return
		}
	}
}

// Reset TTL for existing entries, or creates a new entries with default TTL, for given relationship
// as well as source and destination entities.
func (c *internalStorage) update(relationship *internal.Relationship) {
	sourceHash := buildKey(relationship.Source)
	destHash := buildKey(relationship.Destination)

	c.entities.SetWithTTL(sourceHash, storedEntity{
		Type: relationship.Source.Type,
		IDs:  relationship.Source.IDs,
	}, itemCost, c.ttl*entityExpirationFactor)

	c.entities.SetWithTTL(destHash, storedEntity{
		Type: relationship.Destination.Type,
		IDs:  relationship.Destination.IDs,
	}, itemCost, c.ttl*entityExpirationFactor)

	relationshipKey := fmt.Sprintf("%s:%s:%s", relationship.Type, sourceHash, destHash)
	relationshipValue := storedRelationship{
		sourceHash:       sourceHash,
		destHash:         destHash,
		relationshipType: relationship.Type,
	}
	c.relationships.SetWithTTL(relationshipKey, relationshipValue, itemCost, c.ttl)
}

// buildKey constructs a unique key for the entity referenced in the relationship.
// The key is composition of entity type and its ID attributes.
func buildKey(entity internal.RelationshipEntity) string {
	e := storedEntity{
		Type: entity.Type,
		IDs:  entity.IDs,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		panic(fmt.Sprintf("Failed to encode cache value: %v", err))
	}

	key := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", key)
}
