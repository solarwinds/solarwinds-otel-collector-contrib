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

package config

import (
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/confmap/xconfmap"
)

const (
	defaultInterval           = 5 * time.Minute
	defaultTTLCleanupInterval = 5 * time.Second
	defaultMaxCapacity        = 1_000_000 // 1 million items in the cache
)

type ExpirationSettings struct {
	Enabled bool `mapstructure:"enabled"`
	// TTL of relationships in the format accepted by time.ParseDuration, e.g. "5s", "1m", etc.
	Interval           string             `mapstructure:"interval"`
	CacheConfiguration CacheConfiguration `mapstructure:"cache_configuration"`
}

// By implementing the xconfmap.Validator, we ensure it's validated by the collector automatically
var _ xconfmap.Validator = (*ExpirationSettings)(nil)

type CacheConfiguration struct {
	MaxCapacity int64 `mapstructure:"max_capacity"`
	// In the format accepted by time.ParseDuration, e.g. "5s", "1m", etc. Granularity is 1 second and it's also a minimal value.
	TTLCleanupInterval string `mapstructure:"ttl_cleanup_interval"`
}

func (e *ExpirationSettings) Validate() error {
	if !e.Enabled {
		return nil
	}

	var errs error

	if _, err := e.getInterval(); err != nil {
		errs = errors.Join(errs, err)
	}

	if _, err := e.getMaxCapacity(); err != nil {
		errs = errors.Join(errs, err)
	}

	if _, err := e.getTTLCleanupInterval(); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}

type ExpirationPolicy struct {
	Enabled                   bool
	Interval                  time.Duration
	TTLCleanupIntervalSeconds time.Duration
	MaxCapacity               int64
}

func (e *ExpirationSettings) Unmarshal() (ExpirationPolicy, error) {
	if !e.Enabled {
		return ExpirationPolicy{Enabled: false}, nil
	}

	interval, err := e.getInterval()
	if err != nil {
		return ExpirationPolicy{}, err
	}

	maxCapacity, err := e.getMaxCapacity()
	if err != nil {
		return ExpirationPolicy{}, err
	}
	ttlCleanupInterval, err := e.getTTLCleanupInterval()
	if err != nil {
		return ExpirationPolicy{}, err
	}

	return ExpirationPolicy{
		Enabled:                   true,
		Interval:                  interval,
		TTLCleanupIntervalSeconds: ttlCleanupInterval,
		MaxCapacity:               maxCapacity,
	}, nil
}

func (e *ExpirationSettings) getInterval() (time.Duration, error) {
	interval, err := time.ParseDuration(e.Interval)
	if err != nil {
		return time.Duration(0), fmt.Errorf("the interval must be a valid duration (e.g., '5m', '1h'): %w", err)
	}

	return interval, nil
}

func (e *ExpirationSettings) getMaxCapacity() (int64, error) {
	if e.CacheConfiguration.MaxCapacity <= 0 {
		return 0, errors.New("the cache_configuration::max_capacity must be greater than zero")
	}

	return e.CacheConfiguration.MaxCapacity, nil
}

func (e *ExpirationSettings) getTTLCleanupInterval() (time.Duration, error) {
	parsedCleanupInterval, err := time.ParseDuration(e.CacheConfiguration.TTLCleanupInterval)
	if err != nil {
		return time.Duration(0), errors.New("the cache_configuration::ttl_cleanup_interval has invalid format")
	}
	if parsedCleanupInterval < time.Second {
		return time.Duration(0), errors.New("the cache_configuration::ttl_cleanup_interval must be at least 1 second")
	}

	return parsedCleanupInterval, nil
}
