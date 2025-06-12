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
	"sync"
	"time"
)

const (
	itemCost = 1
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

type ttlCache struct {
	entities      *ristretto.Cache[string, storedEntity]
	relationships *ristretto.Cache[string, storedRelationship]
	ttl           time.Duration
	interval      time.Duration
	mu            sync.Mutex
	logger        *zap.Logger
}

func newTTLCache(cfg *config.ExpirationSettings, logger *zap.Logger, em chan<- internal.Subject) *ttlCache {
	var err error

	entityCache, err := ristretto.NewCache(&ristretto.Config[string, storedEntity]{
		NumCounters: cfg.MaxCapacity * 10,
		MaxCost:     cfg.MaxCapacity,
		// TtlTickerDurationInSec is set to 10 times the cleanup interval to ensure that entities are cleaned up
		// in a timely manner, but not too frequently.
		TtlTickerDurationInSec: int64(cfg.TTLCleanupInterval.Seconds()) * 10,
		BufferItems:            64,
	})

	relationshipCache, err := ristretto.NewCache(&ristretto.Config[string, storedRelationship]{
		NumCounters:            cfg.MaxCapacity * 10,
		MaxCost:                cfg.MaxCapacity,
		TtlTickerDurationInSec: int64(cfg.TTLCleanupInterval.Seconds()),
		BufferItems:            64,
		OnEvict: func(item *ristretto.Item[storedRelationship]) {
			logger.Info("ttlCache relationship evicted")

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

	return &ttlCache{
		entities:      entityCache,
		relationships: relationshipCache,
		ttl:           cfg.Interval,
		logger:        logger,
		mu:            sync.Mutex{},
	}
}

func (c *ttlCache) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("ttlCache stopped")
			return
		}
	}
}

func (c *ttlCache) Update(relationship *internal.Relationship) {
	sourceHash := buildKey(relationship.Source)
	destHash := buildKey(relationship.Destination)

	c.entities.SetWithTTL(sourceHash, storedEntity{
		Type: relationship.Source.Type,
		IDs:  relationship.Source.IDs,
	}, itemCost, c.ttl*10)

	c.entities.SetWithTTL(destHash, storedEntity{
		Type: relationship.Destination.Type,
		IDs:  relationship.Destination.IDs,
	}, itemCost, c.ttl*10)

	relationshipKey := fmt.Sprintf("%s:%s:%s", relationship.Type, sourceHash, destHash)
	relationshipValue := storedRelationship{
		sourceHash:       sourceHash,
		destHash:         destHash,
		relationshipType: relationship.Type,
	}
	c.relationships.SetWithTTL(relationshipKey, relationshipValue, itemCost, c.ttl)
}

// buildKey constructs a unique key for the cache based on the relationship event.
// The key is composition of relationship type, source and destination entity types
// and their ID values.
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
