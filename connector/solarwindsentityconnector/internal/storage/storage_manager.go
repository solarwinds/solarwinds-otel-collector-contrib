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
	cache         InternalCache
	expiredCh     chan internal.Event
	eventConsumer internal.EventConsumer

	logger *zap.Logger
}

func NewStorageManager(cfg *config.ExpirationSettings, logger *zap.Logger, logsConsumer internal.EventConsumer) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("expiration settings configuration is nil")
	}

	expiredCh := make(chan internal.Event)
	cache, err := newInternalStorage(cfg, logger, expiredCh)
	if err != nil {
		return nil, fmt.Errorf("failed to create internal storage: %w", err)
	}

	return &Manager{
		cache:         cache,
		expiredCh:     expiredCh,
		eventConsumer: logsConsumer,
		logger:        logger,
	}, nil
}

func (m *Manager) Start(ctx context.Context) error {
	go m.receiveExpired(ctx)
	go m.cache.run(ctx)

	return nil
}

// Update last seen of a relationship and its entities in the cache.
func (m *Manager) Update(s internal.Event) error {
	if r, ok := s.(*internal.Relationship); ok {
		return m.cache.update(r)
	}
	return nil
}

// Delete relationship from the cache. Entities stay until expiration.
// Does not trigger onEvict callback, and thus does not send the event to the consumer.
func (m *Manager) Delete(s internal.Event) error {
	if r, ok := s.(*internal.Relationship); ok {
		return m.cache.delete(r)
	}
	return nil
}

// receiveExpired listens for expired relationships and sends them in batches to the event consumer.
func (m *Manager) receiveExpired(ctx context.Context) {
	var batch []internal.Event
	var timer *time.Timer
	var timerC <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Context done, stopping storage manager")
			// when context is closed, we will not send the remaining batch of evicted relationships
			close(m.expiredCh)
			return

		case rel := <-m.expiredCh:
			if batch == nil {
				batch = make([]internal.Event, 0)
				timer = time.NewTimer(1 * time.Second)
				timerC = timer.C
			}
			batch = append(batch, rel)

		case <-timerC:
			m.logger.Debug("timer expired, sending batch of evicted relationships", zap.Int("count", len(batch)))
			if len(batch) > 0 {
				m.eventConsumer.SendExpiredEvents(ctx, batch)
				batch = nil // Reset the batch after sending
				timer.Stop()
				timerC = nil
			}
		}
	}
}
