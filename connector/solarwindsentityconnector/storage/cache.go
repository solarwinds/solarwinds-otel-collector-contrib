package storage

import (
	"github.com/dgraph-io/ristretto/v2"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"time"
)

var (
	cache       *ristretto.Cache[string, string]
	ttlInterval time.Duration
)

const (
	itemCost = 1
)

func StartCache(cfg config.ExpirationPolicy) {
	if cache == nil {
		initCache(cfg)
	} else {
		cache.Clear()
	}
}

func initCache(cfg config.ExpirationPolicy) {
	var err error

	var ttlCleanupInterval time.Duration
	var maxCapacity int64

	if cfg.CacheConfiguration.MaxCapacity != nil {
		maxCapacity = *cfg.CacheConfiguration.MaxCapacity
	} else {
		maxCapacity = 1_000_000 // Defaults to 1 million items
	}

	if cfg.CacheConfiguration.TTLCleanupInterval != nil {
		ttlCleanupInterval, err = time.ParseDuration(*cfg.CacheConfiguration.TTLCleanupInterval)
		if err != nil {
			panic("Failed to parse TTL cleanup interval: " + err.Error())
		}
	} else {
		ttlCleanupInterval = 120 * time.Second // Default to 120 seconds
	}

	cache, err = ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters:            maxCapacity * 10,
		MaxCost:                maxCapacity,
		TtlTickerDurationInSec: int64(ttlCleanupInterval.Seconds()),
	})

	if err != nil {
		panic("Failed to create cache: " + err.Error())
	}

	defer cache.Close()
}

func Update(key, value string) {
	cache.SetWithTTL(key, value, itemCost, ttlInterval)
}
