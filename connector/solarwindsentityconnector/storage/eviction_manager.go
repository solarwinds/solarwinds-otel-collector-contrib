package storage

import (
	"context"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"time"
)

type EvictionManager struct {
	evicted      chan relationship
	logger       *zap.Logger
	logsConsumer consumer.Logs
}

func NewEvictionManager(logsConsumer consumer.Logs, logger *zap.Logger) *EvictionManager {
	return &EvictionManager{
		evicted:      make(chan relationship),
		logger:       logger,
		logsConsumer: logsConsumer,
	}
}

func (em *EvictionManager) Add(rel relationship) {
	em.logger.Info("Adding relationship to eviction manager")
	em.evicted <- rel
}

func (em *EvictionManager) Start(ctx context.Context) {
	go func() {
		var batch []relationship
		var timer *time.Timer
		var timerC <-chan time.Time

		for {
			em.logger.Info("in for loop of eviction manager")
			select {
			case <-ctx.Done():
				em.logger.Info("Context done, stopping eviction manager")
				// Send the remaining batch of evicted relationships
				if len(batch) > 0 {
					em.send(batch, ctx)
				}
				close(em.evicted)
				return

			case rel := <-em.evicted:
				em.logger.Info("Received evicted relationship", zap.String("type", rel.RelationshipType))
				if batch == nil {
					batch = make([]relationship, 0)
					timer = time.NewTimer(1 * time.Second)
					timerC = timer.C
				}
				batch = append(batch, rel)

			case <-timerC:
				// Timer expired, send the batch of evicted relationships
				em.logger.Info("Timer expired, sending batch of evicted relationships", zap.Int("count", len(batch)))
				if len(batch) > 0 {
					em.send(batch, ctx)
					batch = nil // Reset the batch after sending
					timer.Stop()
					timerC = nil
				}
			}
		}
	}()
}

func (em *EvictionManager) send(batch []relationship, ctx context.Context) {
	em.logger.Info("Sending eviction logs", zap.Int("count", len(batch)))
	err := em.logsConsumer.ConsumeLogs(ctx, plog.NewLogs())
	if err != nil {
		em.logger.Error("failed to generate eviction logs", zap.Error(err))
	}
}
