package consumer

import (
	"context"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	otelConsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
)

type Consumer interface {
	SendExpiredEvents(ctx context.Context, relationships []internal.Subject)
}

const (
	entityEventAsLog = "otel.entity.event_as_log"
)

type consumer struct {
	logsConsumer otelConsumer.Logs
	entities     map[string]config.Entity
}

var _ Consumer = (*consumer)(nil)

func NewConsumer(logsConsumer otelConsumer.Logs, entities map[string]config.Entity) Consumer {
	return &consumer{
		logsConsumer: logsConsumer,
		entities:     entities,
	}
}

func (c *consumer) SendExpiredEvents(ctx context.Context, events []internal.Subject) {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr(entityEventAsLog, "true")
	//sl := rl.ScopeLogs().AppendEmpty()
	//timestamp := pcommon.NewTimestampFromTime(time.Now())

	//	for _, e := range events {
	//		e.Expire()
	//	}
}
