package internal

import (
	"context"
	otelConsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
)

type Consumer interface {
	SendExpiredEvents(ctx context.Context, relationships []Subject)
}

type consumer struct {
	logsConsumer otelConsumer.Logs
}

var _ Consumer = (*consumer)(nil)

func NewConsumer(logsConsumer otelConsumer.Logs) Consumer {
	return &consumer{
		logsConsumer: logsConsumer,
	}
}

func (c *consumer) SendExpiredEvents(ctx context.Context, events []Subject) {
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)

	for _, e := range events {
		e.Delete(logRecords)
	}

	err := c.logsConsumer.ConsumeLogs(ctx, logs)
	if err != nil {
		panic("failed to consume logs: " + err.Error())
	}
}
