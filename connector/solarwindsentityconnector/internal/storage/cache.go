package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	itemCost = 1
)

type ttlCache struct {
	data     *ristretto.Cache[string, internal.Subject]
	ttl      time.Duration
	interval time.Duration
	mu       sync.Mutex
	logger   *zap.Logger
}

func NewTTLCache(cfg *config.ExpirationSettings, logger *zap.Logger, em chan<- internal.Subject) *ttlCache {
	var err error

	cache, err := ristretto.NewCache(&ristretto.Config[string, internal.Subject]{
		NumCounters:            cfg.MaxCapacity * 10,
		MaxCost:                cfg.MaxCapacity,
		TtlTickerDurationInSec: int64(cfg.TTLCleanupInterval.Seconds()),
		BufferItems:            64,
		OnEvict: func(item *ristretto.Item[internal.Subject]) {
			logger.Info("ttlCache item evicted")
			if item != nil {
				em <- item.Value
			}
		},
	})

	if err != nil {
		panic("Failed to create cache: " + err.Error())
	}

	return &ttlCache{
		data:   cache,
		ttl:    cfg.Interval,
		logger: logger,
		mu:     sync.Mutex{},
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

func (c *ttlCache) Update(relationship internal.Subject) {
	key := buildKey(relationship)

	c.data.SetWithTTL(key, relationship, itemCost, c.ttl)
}

// buildKey constructs a unique key for the cache based on the relationship event.
// The key is composition of relationship type, source and destination entity types
// and their ID values.
func buildKey(relationship internal.Subject) string {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(relationship); err != nil {
		panic(fmt.Sprintf("Failed to encode cache value: %v", err))
	}

	key := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", key)
}
