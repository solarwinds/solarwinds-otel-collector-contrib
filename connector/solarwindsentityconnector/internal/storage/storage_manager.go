package storage

import (
	"context"
	"fmt"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.uber.org/zap"
	"time"
)

type Manager struct {
	cache        *internalStorage
	expiredCh    chan internal.Subject
	logsConsumer internal.Consumer

	logger *zap.Logger
}

func NewStorageManager(cfg *config.ExpirationSettings, logger *zap.Logger, logsConsumer internal.Consumer) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("expiration settings configuration is nil")
	}

	expiredCh := make(chan internal.Subject)
	cache, err := newInternalStorage(cfg, logger, expiredCh)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal storage: %w", err)
	}

	return &Manager{
		cache:        cache,
		expiredCh:    expiredCh,
		logsConsumer: logsConsumer,
		logger:       logger,
	}, nil
}

func (m *Manager) Start(ctx context.Context) error {
	go m.receiveExpired(ctx)
	go m.cache.run(ctx)

	return nil
}

func (m *Manager) Update(s internal.Subject) error {
	if r, ok := s.(*internal.Relationship); ok {
		return m.cache.update(r)
	}
	return nil
}

func (m *Manager) receiveExpired(ctx context.Context) {
	var batch []internal.Subject
	var timer *time.Timer
	var timerC <-chan time.Time

	for {
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
			if batch == nil {
				batch = make([]internal.Subject, 0)
				timer = time.NewTimer(1 * time.Second)
				timerC = timer.C
			}
			batch = append(batch, rel)

		case <-timerC:
			// Timer expired, send the batch of evicted relationships
			m.logger.Debug("timer expired, sending batch of evicted relationships", zap.Int("count", len(batch)))
			if len(batch) > 0 {
				m.send(batch, ctx)
				batch = nil // Reset the batch after sending
				timer.Stop()
				timerC = nil
			}
		}
	}
}

func (m *Manager) send(batch []internal.Subject, ctx context.Context) {
	m.logsConsumer.SendExpiredEvents(ctx, batch)
}
