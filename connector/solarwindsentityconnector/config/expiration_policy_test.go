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
