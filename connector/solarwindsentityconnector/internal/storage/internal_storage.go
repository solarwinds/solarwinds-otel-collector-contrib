package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.uber.org/zap"
	"hash/fnv"
	"time"
)

const (
	// ristretto cache provides cost management, but we need to set the same cost for all items.
	// Additionally, when itemCost is set to 1 the MaxCost is equal to the maximum number of items in the cache.
	itemCost = 1
	// When dealing with TTL for entities we do not need to be as precise as with relationships.
	// Thus cleaning up the cache every 10 times the interval.
	entityTTLCleanupFactor = 10
	// entityTTLFactor is a factor that controls how long the entity will be kept in the cache,
	// after the relationship is updated.
	entityTTLFactor = 5
	// bufferItems is a configuration parameter that controls how many key-value operations
	// are buffered before being processed by the internal ristretto eviction logic.
	// The value is recommended by ristretto documentation but not set to any default value.
	bufferItems = 64
)

type storedRelationship struct {
	sourceHash       string
	destHash         string
	relationshipType string
}

type internalStorage struct {
	entities      *ristretto.Cache[string, internal.RelationshipEntity]
	relationships *ristretto.Cache[string, storedRelationship]
	ttl           time.Duration
	interval      time.Duration
	logger        *zap.Logger

	// TODO: Introduce mutex to protect concurrent access to the cache when parallelization is used
	// in the upper layers (NH-112603).
	// mu sync.Mutex
}

func newInternalStorage(cfg *config.ExpirationSettings, logger *zap.Logger, em chan<- internal.Event) (*internalStorage, error) {
	var err error

	// maxCost sets the maximum number of items, when itemCost is set to 1
	maxCost := cfg.MaxCapacity
	// numCounters are internal setting for ristretto cache that helps to manage eviction mechanism.
	// It is recommended to set it to 10 times the maximum capacity for most of the use cases.
	numCounters := cfg.MaxCapacity * 10
	// ttlCleanup sets the interval for sequence scan of the evicted items. When item is evicted,
	// it is not immediately removed from the cache.
	ttlCleanup := int64(cfg.TTLCleanupInterval.Seconds())

	entityCache, err := ristretto.NewCache(&ristretto.Config[string, internal.RelationshipEntity]{
		NumCounters:            numCounters,
		MaxCost:                maxCost,
		TtlTickerDurationInSec: ttlCleanup * entityTTLCleanupFactor,
		BufferItems:            bufferItems,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create entity cache: %w", err)
	}

	relationshipCache, err := ristretto.NewCache(&ristretto.Config[string, storedRelationship]{
		NumCounters:            numCounters,
		MaxCost:                maxCost,
		TtlTickerDurationInSec: ttlCleanup,
		BufferItems:            bufferItems,
		OnEvict: func(item *ristretto.Item[storedRelationship]) {
			onRelationshipEvict(item, entityCache, logger, em)
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create relaitonship cache: %w", err)
	}

	return &internalStorage{
		entities:      entityCache,
		relationships: relationshipCache,
		ttl:           cfg.Interval,
		logger:        logger,
	}, nil
}

// onRelationshipEvict is a callback function that is called when a relationship item is evicted from the cache.
// It retrieves the source and destination entities from the entity cache and sends a relationship event.
func onRelationshipEvict(
	item *ristretto.Item[storedRelationship],
	entityCache *ristretto.Cache[string, internal.RelationshipEntity],
	logger *zap.Logger,
	em chan<- internal.Event) {

	logger.Debug("relationship item evicted", zap.String("relationshipType", item.Value.relationshipType))
	source, sourceExists := entityCache.Get(item.Value.sourceHash)
	if !sourceExists {
		logger.Warn("source entity not found in cache", zap.String("hash", item.Value.sourceHash))
		return
	}

	dest, destExists := entityCache.Get(item.Value.destHash)
	if !destExists {
		logger.Warn("destination entity not found in cache", zap.String("hash", item.Value.sourceHash))
		return
	}

	em <- &internal.Relationship{
		Type: item.Value.relationshipType,
		Source: internal.RelationshipEntity{
			Type: source.Type,
			IDs:  source.IDs,
		},
		Destination: internal.RelationshipEntity{
			Type: dest.Type,
			IDs:  dest.IDs,
		},
	}
}

func (c *internalStorage) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("internalStorage stopped")
			return
		}
	}
}

// Reset TTL for existing entries, or creates a new entries with default TTL, for given relationship
// as well as source and destination entities.
func (c *internalStorage) update(relationship *internal.Relationship) error {
	c.logger.Debug("updating relationship in internal storage", zap.String("relationshipType", relationship.Type))

	sourceHash, err := buildKey(relationship.Source)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to hash key for source entity: %s", relationship.Source.Type))
	}
	destHash, err := buildKey(relationship.Destination)
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to hash key for destination entity: %s", relationship.Destination.Type))
	}

	sourceUpdated := c.entities.SetWithTTL(sourceHash, relationship.Source, itemCost, c.ttl*entityTTLFactor)
	if !sourceUpdated {
		return fmt.Errorf("failed to update source entity: %s", relationship.Source.Type)
	}

	destUpdated := c.entities.SetWithTTL(destHash, relationship.Destination, itemCost, c.ttl*entityTTLFactor)
	if !destUpdated {
		return fmt.Errorf("failed to update relationship entity: %s", relationship.Destination.Type)
	}

	relationshipKey := fmt.Sprintf("%s:%s:%s", relationship.Type, sourceHash, destHash)
	relationshipValue := storedRelationship{
		sourceHash:       sourceHash,
		destHash:         destHash,
		relationshipType: relationship.Type,
	}

	relationshipUpdated := c.relationships.SetWithTTL(relationshipKey, relationshipValue, itemCost, c.ttl)
	if !relationshipUpdated {
		return fmt.Errorf("failed to update relationship: %s", relationship.Type)
	}
	return nil
}

// buildKey constructs a unique key for the entity referenced in the relationship.
// The key is composition of entity type and its ID attributes.
func buildKey(entity internal.RelationshipEntity) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(struct {
		Type string
		IDs  map[string]any
	}{
		entity.Type,
		entity.IDs.AsRaw(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to encode entity: %w", err)
	}

	h := fnv.New64a()
	_, err = h.Write(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to write entity bytes to hash: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}
