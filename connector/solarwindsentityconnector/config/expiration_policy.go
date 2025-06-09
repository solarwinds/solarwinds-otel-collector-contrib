package config

import (
	"errors"
	"time"
)

type ExpirationPolicy struct {
	Enabled            bool                `mapstructure:"enabled"`
	Interval           string              `mapstructure:"interval"`
	CacheConfiguration *CacheConfiguration `mapstructure:"cache_configuration"`
}

type CacheConfiguration struct {
	MaxCapacity        *int64  `mapstructure:"max_capacity"`
	TTLCleanupInterval *string `mapstructure:"ttl_cleanup_interval"`
}

type ExpirationSettings struct {
	Enabled            bool
	Interval           time.Duration
	TTLCleanupInterval time.Duration
	MaxCapacity        int64
}

func (e *ExpirationPolicy) Parse() *ExpirationSettings {
	if !e.Enabled {
		return &ExpirationSettings{
			Enabled: false,
		}
	}

	var interval time.Duration

	if e.Interval == "" {
		return nil
	} else {
		interval, _ = time.ParseDuration(e.Interval)
	}

	maxCapacity, err := e.getMaxCapacity()
	if err != nil {
		return nil // Invalid max capacity
	}
	ttlCleanupInterval, err := e.getTTLCleanupInterval()
	if err != nil {
		return nil // Invalid TTL cleanup interval
	}

	return &ExpirationSettings{
		Enabled:            true,
		Interval:           interval,
		TTLCleanupInterval: ttlCleanupInterval,
		MaxCapacity:        maxCapacity,
	}
}

func (e *ExpirationPolicy) getMaxCapacity() (int64, error) {
	defaultMaxCapacity := int64(1_000_000) // Default to 1 million items
	if e.CacheConfiguration != nil && e.CacheConfiguration.MaxCapacity != nil {
		if *e.CacheConfiguration.MaxCapacity <= 0 {
			return 0, errors.New("max capacity must be greater than zero")
		}

		return *e.CacheConfiguration.MaxCapacity, nil
	}

	return defaultMaxCapacity, nil
}

func (e *ExpirationPolicy) getTTLCleanupInterval() (time.Duration, error) {
	if e.CacheConfiguration != nil && e.CacheConfiguration.TTLCleanupInterval != nil {
		parsedCleanupInterval, err := time.ParseDuration(*e.CacheConfiguration.TTLCleanupInterval)
		if err != nil {
			return time.Duration(0), errors.New("invalid TTL cleanup interval format")
		}
		if parsedCleanupInterval <= 0 {
			return time.Duration(0), errors.New("ttl cleanup interval must be longer than 0 seconds")
		}

		return parsedCleanupInterval, nil
	}

	defaultTTLCleanupInterval := 10 * time.Second
	return defaultTTLCleanupInterval, nil
}
