package config

import "errors"

type ExpirationPolicy struct {
	Enabled            bool                `mapstructure:"enabled"`
	Interval           string              `mapstructure:"interval"`
	CacheConfiguration *CacheConfiguration `mapstructure:"cache_configuration"`
}

type CacheConfiguration struct {
	MaxCapacity        *int64  `mapstructure:"max_capacity"`
	TTLCleanupInterval *string `mapstructure:"ttl_cleanup_interval"`
}

func (e *ExpirationPolicy) IsValid() error {
	if e.Interval == "" {
		return errors.New("empty expiration interval")
	}

	if e.CacheConfiguration != nil {
		maxCapacity := e.CacheConfiguration.MaxCapacity
		if maxCapacity != nil && *maxCapacity <= 0 {
			return errors.New("invalid max capacity")
		}

		ttlCleanupInterval := e.CacheConfiguration.TTLCleanupInterval
		if ttlCleanupInterval != nil && *ttlCleanupInterval != "" {
			return errors.New("invalid TTL cleanup interval")
		}
	}

	return nil
}
