package internal

import (
	"context"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
)

type EventConsumer interface {
	SendExpiredEvents(ctx context.Context, relationships []Event)
}

type eventConsumer struct {
	logsConsumer consumer.Logs
}

var _ EventConsumer = (*eventConsumer)(nil)

func NewConsumer(logsConsumer consumer.Logs) EventConsumer {
	return &eventConsumer{
		logsConsumer: logsConsumer,
	}
}

func (c *eventConsumer) SendExpiredEvents(ctx context.Context, events []Event) {
	logs := plog.NewLogs()
	logRecords := CreateEventLog(&logs)

	for _, e := range events {
		e.Delete(logRecords)
	}

	err := c.logsConsumer.ConsumeLogs(ctx, logs)
	// TODO: This has to be reworked to use error channel in the refactoring task,
	// since the consumer is run in the go routine.
	if err != nil {
		panic("failed to consume logs: " + err.Error())
	}
}
