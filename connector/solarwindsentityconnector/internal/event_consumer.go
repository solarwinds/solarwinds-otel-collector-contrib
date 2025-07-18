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

func NewEventConsumer(logsConsumer consumer.Logs) EventConsumer {
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
