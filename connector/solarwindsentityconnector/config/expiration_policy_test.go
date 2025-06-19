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
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseReturnsWithoutErrors(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := "5s"
	expirationPolicy := ExpirationPolicy{
		Enabled:  true,
		Interval: "5s",
		CacheConfiguration: &CacheConfiguration{
			MaxCapacity:        &capacity,
			TTLCleanupInterval: &cleanupInterval,
		},
	}

	expectedExpirationPolicy := ExpirationSettings{
		Enabled:            true,
		Interval:           5 * time.Second,
		TTLCleanupInterval: 5 * time.Second,
		MaxCapacity:        10,
	}
	result, err := expirationPolicy.Parse()
	assert.Nil(t, err)
	assert.Equal(t, result, &expectedExpirationPolicy)
}

func TestParseWhenDisabledReturnsError(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := "5s"
	expirationPolicy := ExpirationPolicy{
		Enabled:  false,
		Interval: "1m",
		CacheConfiguration: &CacheConfiguration{
			MaxCapacity:        &capacity,
			TTLCleanupInterval: &cleanupInterval,
		},
	}

	_, err := expirationPolicy.Parse()
	assert.ErrorContains(t, err, "expiration policy is not enabled")
}

func TestParseWithEmptyIntervalReturnsError(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := "5s"
	expirationPolicy := ExpirationPolicy{
		Enabled:  true,
		Interval: "",
		CacheConfiguration: &CacheConfiguration{
			MaxCapacity:        &capacity,
			TTLCleanupInterval: &cleanupInterval,
		},
	}

	_, err := expirationPolicy.Parse()
	assert.ErrorContains(t, err, "expiration interval is not set")
}

func TestParseWithEmptyCleanupIntervalReturnsError(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := ""
	expirationPolicy := ExpirationPolicy{
		Enabled:  true,
		Interval: "5s",
		CacheConfiguration: &CacheConfiguration{
			MaxCapacity:        &capacity,
			TTLCleanupInterval: &cleanupInterval,
		},
	}

	_, err := expirationPolicy.Parse()
	assert.ErrorContains(t, err, "invalid TTL cleanup interval format")
}

func TestParseWithZeroCleanupIntervalReturnsError(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := "0"
	expirationPolicy := ExpirationPolicy{
		Enabled:  true,
		Interval: "5s",
		CacheConfiguration: &CacheConfiguration{
			MaxCapacity:        &capacity,
			TTLCleanupInterval: &cleanupInterval,
		},
	}

	_, err := expirationPolicy.Parse()
	assert.ErrorContains(t, err, "ttl cleanup interval must be longer than 0 seconds")
}

func TestParseMaxCapacityLessThenZeroReturnError(t *testing.T) {
	capacity := int64(-10)
	cleanupInterval := "5s"
	expirationPolicy := ExpirationPolicy{
		Enabled:  true,
		Interval: "5s",
		CacheConfiguration: &CacheConfiguration{
			MaxCapacity:        &capacity,
			TTLCleanupInterval: &cleanupInterval,
		},
	}

	_, err := expirationPolicy.Parse()
	assert.ErrorContains(t, err, "max capacity must be greater than zero")
}
