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
)

const (
	defaultTTLCleanupInterval = 5 * time.Second
	defaultMaxCapacity        = int64(1_000_000) // 1 million items in the cache
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

func (e *ExpirationPolicy) Parse() (*ExpirationSettings, error) {
	if !e.Enabled {
		return nil, fmt.Errorf("expiration policy is not enabled")
	}

	var interval time.Duration

	if e.Interval == "" {
		return nil, fmt.Errorf("expiration interval is not set")
	} else {
		interval, _ = time.ParseDuration(e.Interval)
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
		Enabled:            true,
		Interval:           interval,
		TTLCleanupInterval: ttlCleanupInterval,
		MaxCapacity:        maxCapacity,
	}, nil
}

func (e *ExpirationPolicy) getMaxCapacity() (int64, error) {
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

	return defaultTTLCleanupInterval, nil
}
