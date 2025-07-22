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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalReturnsWithoutErrors(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := "10s"
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "5s",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        capacity,
			TTLCleanupInterval: cleanupInterval,
		},
	}

	expectedExpirationPolicy := ExpirationPolicy{
		Enabled:                   true,
		Interval:                  5 * time.Second,
		TTLCleanupIntervalSeconds: 10 * time.Second,
		MaxCapacity:               10,
	}
	result, err := expirationPolicy.Unmarshal()
	assert.Nil(t, err)
	assert.Equal(t, result, &expectedExpirationPolicy)
}

func TestUnmarshalWhenDisabledReturnsValidResult(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := "5s"
	expirationPolicy := ExpirationSettings{
		Enabled:  false,
		Interval: "1m",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        capacity,
			TTLCleanupInterval: cleanupInterval,
		},
	}

	result, err := expirationPolicy.Unmarshal()
	assert.NoError(t, err)
	assert.False(t, result.Enabled)
}

func TestUnmarshalWithEmptyIntervalReturnsError(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := "5s"
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        capacity,
			TTLCleanupInterval: cleanupInterval,
		},
	}

	_, err := expirationPolicy.Unmarshal()
	assert.ErrorContains(t, err, "interval must be a valid duration")
}

func TestUnmarshalWithEmptyCleanupIntervalReturnsError(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := ""
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "5s",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        capacity,
			TTLCleanupInterval: cleanupInterval,
		},
	}

	_, err := expirationPolicy.Unmarshal()
	assert.ErrorContains(t, err, "cache_configuration::ttl_cleanup_interval: invalid format")
}

func TestUnmarshalWithZeroCleanupIntervalReturnsError(t *testing.T) {
	capacity := int64(10)
	cleanupInterval := "0"
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "5s",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        capacity,
			TTLCleanupInterval: cleanupInterval,
		},
	}

	_, err := expirationPolicy.Unmarshal()
	assert.ErrorContains(t, err, "cache_configuration::ttl_cleanup_interval must be at least 1 second")
}

func TestUnmarshalMaxCapacityLessThenZeroReturnError(t *testing.T) {
	capacity := int64(-10)
	cleanupInterval := "5s"
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "5s",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        capacity,
			TTLCleanupInterval: cleanupInterval,
		},
	}

	_, err := expirationPolicy.Unmarshal()
	assert.ErrorContains(t, err, "cache_configuration::max_capacity must be greater than zero")
}

func TestValidate_DisabledReturnsNil(t *testing.T) {
	expirationPolicy := ExpirationSettings{
		Enabled: false,
		// These values are invalid, but since Enabled is false, validation should pass
		Interval: "invalid",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        -1,
			TTLCleanupInterval: "invalid",
		},
	}

	err := expirationPolicy.Validate()
	assert.NoError(t, err)
}

func TestValidate_EnabledValidConfigurationReturnsNil(t *testing.T) {
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "1s",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        1,
			TTLCleanupInterval: "1s",
		},
	}

	err := expirationPolicy.Validate()
	assert.NoError(t, err)
}

func TestValidate_InvalidIntervalReturnsError(t *testing.T) {
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "invalid",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        1,
			TTLCleanupInterval: "1s",
		},
	}

	err := expirationPolicy.Validate()
	assert.ErrorContains(t, err, "interval must be a valid duration")
}

func TestValidate_InvalidMaxCapacityReturnsError(t *testing.T) {
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "1s",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        0,
			TTLCleanupInterval: "1s",
		},
	}

	err := expirationPolicy.Validate()
	assert.ErrorContains(t, err, "cache_configuration::max_capacity must be greater than zero")
}

func TestValidate_InvalidTTLCleanupIntervalReturnsError(t *testing.T) {
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "1s",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        1,
			TTLCleanupInterval: "invalid",
		},
	}

	err := expirationPolicy.Validate()
	assert.ErrorContains(t, err, "cache_configuration::ttl_cleanup_interval: invalid format")
}

func TestValidate_TTLCleanupIntervalLessThanOneSecondReturnsError(t *testing.T) {
	expirationPolicy := ExpirationSettings{
		Enabled:  true,
		Interval: "1s",
		CacheConfiguration: CacheConfiguration{
			MaxCapacity:        1,
			TTLCleanupInterval: "500ms",
		},
	}

	err := expirationPolicy.Validate()
	assert.ErrorContains(t, err, "cache_configuration::ttl_cleanup_interval must be at least 1 second")
}
