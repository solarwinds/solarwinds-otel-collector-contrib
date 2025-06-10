package storage

import (
	"context"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
	"time"
)

type Manager struct {
	cache        *ttlCache
	expiredCh    chan relationship
	logsConsumer consumer.Consumer

	logger *zap.Logger
}

func NewStorageManager(cfg *config.ExpirationSettings, logger *zap.Logger, logsConsumer consumer.Consumer) *Manager {
	if cfg == nil {
		logger.Error("Expiration settings configuration is nil")
		return nil
	}

	expiredCh := make(chan relationship)
	cache := NewTTLCache(cfg, logger, expiredCh)

	return &Manager{
		cache:        cache,
		expiredCh:    expiredCh,
		logsConsumer: logsConsumer,
		logger:       logger,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	go m.receiveExpired(ctx)
	go m.cache.Run(ctx)

	return nil
}

func (m *Manager) Update(r config.RelationshipEvent, src, dest pcommon.Map) {
	m.cache.Update(r, src, dest)
}

func (m *Manager) receiveExpired(ctx context.Context) {
	var batch []relationship
	var timer *time.Timer
	var timerC <-chan time.Time

	for {
		m.logger.Info("in for loop of eviction manager")
		select {
		case <-ctx.Done():
			m.logger.Info("Context done, stopping eviction manager")
			// Send the remaining batch of evicted relationships
			if len(batch) > 0 {
				m.send(batch, ctx)
			}
			close(m.expiredCh)
			return

		case rel := <-m.expiredCh:
			m.logger.Info("Received evicted relationship", zap.String("type", rel.Type))
			if batch == nil {
				batch = make([]relationship, 0)
				timer = time.NewTimer(1 * time.Second)
				timerC = timer.C
			}
			batch = append(batch, rel)

		case <-timerC:
			// Timer expired, send the batch of evicted relationships
			m.logger.Info("Timer expired, sending batch of evicted relationships", zap.Int("count", len(batch)))
			if len(batch) > 0 {
				m.send(batch, ctx)
				batch = nil // Reset the batch after sending
				timer.Stop()
				timerC = nil
			}
		}
	}
}

func (m *Manager) send(batch []relationship, ctx context.Context) {
	result := make([]consumer.Relationship, 0, len(batch))
	for _, rel := range batch {
		result = append(result, consumer.Relationship{
			Relationship: config.Relationship{
				Type:        rel.Type,
				Source:      rel.Source,
				Destination: rel.Destination,
			},
			SourceEntityIDs:      rel.SourceEntityIDs,
			DestinationEntityIDs: rel.DestinationEntityIDs,
		})
	}
	m.logsConsumer.SendExpiredRelationships(ctx, result)
}
