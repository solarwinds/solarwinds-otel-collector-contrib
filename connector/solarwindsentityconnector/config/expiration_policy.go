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
	"go.opentelemetry.io/collector/confmap/xconfmap"
	"time"
)

const (
	defaultInterval           = 5 * time.Minute
	defaultTTLCleanupInterval = 5 * time.Second
	defaultMaxCapacity        = 1_000_000 // 1 million items in the cache
)

type ExpirationPolicy struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
	// TTL of relationships in the format accepted by time.ParseDuration, e.g. "5s", "1m", etc.
	Interval           string              `mapstructure:"interval" yaml:"interval"`
	CacheConfiguration CacheConfiguration `mapstructure:"cache_configuration" yaml:"cache_configuration"`
}

var _ xconfmap.Validator = &Schema{}

type CacheConfiguration struct {
	MaxCapacity int64 `mapstructure:"max_capacity" yaml:"max_capacity"`
	// In the format accepted by time.ParseDuration, e.g. "5s", "1m", etc. Granularity is 1 second and it's also a minimal value.
	TTLCleanupInterval string `mapstructure:"ttl_cleanup_interval" yaml:"ttl_cleanup_interval"`
}

type ExpirationSettings struct {
	Enabled                   bool
	Interval                  time.Duration
	TTLCleanupIntervalSeconds time.Duration
	MaxCapacity               int64
}

func (e *ExpirationPolicy) Validate() error {
	if !e.Enabled {
		return nil
	}

	var allErrs error

	if _, err := e.getInterval(); err != nil {
		allErrs = errors.Join(allErrs, err)
	}

	if _, err := e.getMaxCapacity(); err != nil {
		allErrs = errors.Join(allErrs, err)
	}

	if _, err := e.getTTLCleanupInterval(); err != nil {
		allErrs = errors.Join(allErrs, err)
	}

	return allErrs
}

func (e *ExpirationPolicy) Unmarshal() (*ExpirationSettings, error) {
	if !e.Enabled {
		return &ExpirationSettings{Enabled: false}, nil
	}

	if e.Interval == "" {
		return nil, fmt.Errorf("expiration interval is not set")
	}

	interval, err := e.getInterval()
	if err != nil {
		return nil, err
	}

	maxCapacity, err := e.getMaxCapacity()
	if err != nil {
		return nil, err
	}
	ttlCleanupInterval, err := e.getTTLCleanupInterval()
	if err != nil {
		return nil, err
	}

	return &ExpirationSettings{
		Enabled:                   true,
		Interval:                  interval,
		TTLCleanupIntervalSeconds: ttlCleanupInterval,
		MaxCapacity:               maxCapacity,
	}, nil
}

func (e *ExpirationPolicy) getInterval() (time.Duration, error) {
	interval, err := time.ParseDuration(e.Interval)
	if err != nil {
		return time.Duration(0), fmt.Errorf("expiration_policy.interval must be a valid duration (e.g., '5m', '1h'): %w", err)
	}

	return interval, nil
}

func (e *ExpirationPolicy) getMaxCapacity() (int64, error) {
	if e.CacheConfiguration.MaxCapacity <= 0 {
		return 0, errors.New("expiration_policy.cache_configuration.max_capacity must be greater than zero")
	}

	return e.CacheConfiguration.MaxCapacity, nil
}

func (e *ExpirationPolicy) getTTLCleanupInterval() (time.Duration, error) {
	parsedCleanupInterval, err := time.ParseDuration(e.CacheConfiguration.TTLCleanupInterval)
	if err != nil {
		return time.Duration(0), errors.New("expiration_policy.cache_configuration.ttl_cleanup_interval: invalid format")
	}
	if parsedCleanupInterval < time.Second {
		return time.Duration(0), errors.New("expiration_policy.cache_configuration.ttl_cleanup_interval must be at least 1 second")
	}

	return parsedCleanupInterval, nil
}
