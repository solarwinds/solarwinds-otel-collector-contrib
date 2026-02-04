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

package config

import (
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"
	"go.opentelemetry.io/collector/component"
)

type ParsedEvents struct {
	MetricEvents EventsGroup[*ottlmetric.TransformContext]
	LogEvents    EventsGroup[*ottllog.TransformContext]
}

type EventsGroup[C any] struct {
	Entities      []ParsedEvent[EntityEvent, C]
	Relationships []ParsedEvent[RelationshipEvent, C]
}

type ParsedEvent[T any, C any] struct {
	Definition   *T
	ConditionSeq ottl.ConditionSequence[C]
}

type ProcessingContext struct {
	metricParser ottl.Parser[*ottlmetric.TransformContext]
	logParser    ottl.Parser[*ottllog.TransformContext]
	settings     component.TelemetrySettings
}

// createParsedEvents initializes and returns a ParsedEvents structure containing parsed entity and relationship events.
// It exists to parse ottl conditions for entity and relationship events at the time of creation, allowing for efficient evaluation later.
func createParsedEvents(s Schema, settings component.TelemetrySettings) (ParsedEvents, error) {
	ctx := ProcessingContext{settings: settings}

	var err error
	ctx.metricParser, err = ottlmetric.NewParser(ottlfuncs.StandardConverters[*ottlmetric.TransformContext](), settings)
	if err != nil {
		return ParsedEvents{}, fmt.Errorf("failed to create parser for metric events: %w", err)
	}
	ctx.logParser, err = ottllog.NewParser(ottlfuncs.StandardConverters[*ottllog.TransformContext](), settings)
	if err != nil {
		return ParsedEvents{}, fmt.Errorf("failed to create parser for log events: %w", err)
	}

	var result ParsedEvents

	for _, event := range s.Events.Entities {
		err = processEvent(
			event,
			event.Context,
			event.Conditions,
			&result.MetricEvents.Entities,
			&result.LogEvents.Entities,
			ctx)
		if err != nil {
			return ParsedEvents{}, err
		}
	}
	for _, event := range s.Events.Relationships {
		err = processEvent(
			event,
			event.Context,
			event.Conditions,
			&result.MetricEvents.Relationships,
			&result.LogEvents.Relationships,
			ctx)
		if err != nil {
			return ParsedEvents{}, err
		}
	}

	return result, nil
}

func processEvent[T any](
	event T,
	context string,
	conditions []string,
	metricEvents *[]ParsedEvent[T, *ottlmetric.TransformContext],
	logEvents *[]ParsedEvent[T, *ottllog.TransformContext],
	ctx ProcessingContext) error {
	if len(conditions) == 0 {
		conditions = []string{"true"}
	}

	switch context {
	case ottlmetric.ContextName:
		return addEvent(metricEvents, event, conditions, ctx.metricParser, ctx.settings)
	case ottllog.ContextName:
		return addEvent(logEvents, event, conditions, ctx.logParser, ctx.settings)
	default:
		return fmt.Errorf("unsupported context: %s", context)
	}
}

func addEvent[T any, C any](
	parsedEvents *[]ParsedEvent[T, C],
	event T,
	conditions []string,
	parser ottl.Parser[C],
	settings component.TelemetrySettings) error {
	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return fmt.Errorf("failed to parse conditions for event: %w", err)
	}
	seq := ottl.NewConditionSequence(stmts, settings)
	*parsedEvents = append(*parsedEvents, ParsedEvent[T, C]{
		Definition:   &event,
		ConditionSeq: seq,
	})
	return nil
}
