package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	itemCost = 1
)

type Cache struct {
	data     *ristretto.Cache[string, relationship]
	ttl      time.Duration
	interval time.Duration
	mu       sync.Mutex
	logger   *zap.Logger
}

func NewCache(cfg *config.ExpirationSettings, logger *zap.Logger) *Cache {
	var err error

	cache, err := ristretto.NewCache(&ristretto.Config[string, relationship]{
		NumCounters:            cfg.MaxCapacity * 10,
		MaxCost:                cfg.MaxCapacity,
		TtlTickerDurationInSec: int64(cfg.TTLCleanupInterval.Seconds()),
		BufferItems:            64,
		OnEvict: func(item *ristretto.Item[relationship]) {
			if item == nil {
				return
			}
			cacheValue := item.Value
			logger.Info(fmt.Sprintf("Cache evicting %s", item.Key),
				zap.String("relationship_type", cacheValue.RelationshipType),
				zap.String("source_entity_type", cacheValue.SourceEntityType),
				zap.String("destination_entity_type", cacheValue.DestinationEntityType),
				zap.Strings("source_entity_ids", cacheValue.SourceEntityID),
				zap.Strings("destination_entity_ids", cacheValue.DestinationEntityID))
			logger.Debug("Cache item evicted")
		},
	})

	if err != nil {
		panic("Failed to create cache: " + err.Error())
	}

	return &Cache{
		data:   cache,
		ttl:    cfg.Interval,
		logger: logger,
		mu:     sync.Mutex{},
	}
}

func (c *Cache) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Cache stopped")
			return
		}
	}
}

func (c *Cache) Update(relationship config.RelationshipEvent, sourceIds, destinationIds pcommon.Map) {
	key, value := buildKeyValue(relationship, sourceIds, destinationIds)

	c.data.SetWithTTL(key, value, itemCost, c.ttl)
}

// buildKey constructs a unique key for the cache based on the relationship event.
// The key is composition of relationship type, source and destination entity types
// and their ID values.
func buildKeyValue(relationship config.RelationshipEvent, sourceIds, destIds pcommon.Map) (string, relationship) {
	r := newRelationship(relationship, sourceIds, destIds)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		panic(fmt.Sprintf("Failed to encode cache value: %v", err))
	}

	key := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", key), r
}
