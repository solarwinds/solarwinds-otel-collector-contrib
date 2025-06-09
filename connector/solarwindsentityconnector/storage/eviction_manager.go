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

func NewEvictionManager() *EvictionManager {
	return &EvictionManager{
		evicted: make(chan relationship), // Buffered channel to handle evictions
	}
}

func (em *EvictionManager) Start(ctx context.Context) {
	go func() {
		var batch []relationship
		var timer *time.Timer
		var timerC <-chan time.Time

		for {
			select {
			case <-ctx.Done():
				// Send the remaining batch of evicted relationships
				if len(batch) > 0 {
					em.send(batch, ctx)
				}
				return

			case rel := <-em.evicted:
				if batch == nil {
					batch = make([]relationship, 0)
					timer = time.NewTimer(1 * time.Second)
					timerC = timer.C
				}
				batch = append(batch, rel)

			case <-timerC:
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
	err := em.logsConsumer.ConsumeLogs(ctx, plog.NewLogs())
	if err != nil {
		em.logger.Error("failed to generate eviction logs", zap.Error(err))
	}
}
