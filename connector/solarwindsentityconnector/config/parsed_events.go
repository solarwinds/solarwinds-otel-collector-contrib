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
	MetricEvents EventsGroup[ottlmetric.TransformContext]
	LogEvents    EventsGroup[ottllog.TransformContext]
}

// ParsedEventInterface represents a unified interface for parsed events
type ParsedEventInterface[C any] interface {
	IsEntityEvent() bool
	GetEntityEvent() *EntityEvent
	GetRelationshipEvent() *RelationshipEvent
	GetConditionSeq() ottl.ConditionSequence[C]
}

// EntityParsedEvent wraps an EntityEvent with its conditions
type EntityParsedEvent[C any] struct {
	Definition   *EntityEvent
	ConditionSeq ottl.ConditionSequence[C]
}

func (e EntityParsedEvent[C]) IsEntityEvent() bool                        { return true }
func (e EntityParsedEvent[C]) GetEntityEvent() *EntityEvent               { return e.Definition }
func (e EntityParsedEvent[C]) GetRelationshipEvent() *RelationshipEvent   { return nil }
func (e EntityParsedEvent[C]) GetConditionSeq() ottl.ConditionSequence[C] { return e.ConditionSeq }

// RelationshipParsedEvent wraps a RelationshipEvent with its conditions
type RelationshipParsedEvent[C any] struct {
	Definition   *RelationshipEvent
	ConditionSeq ottl.ConditionSequence[C]
}

func (r RelationshipParsedEvent[C]) IsEntityEvent() bool                      { return false }
func (r RelationshipParsedEvent[C]) GetEntityEvent() *EntityEvent             { return nil }
func (r RelationshipParsedEvent[C]) GetRelationshipEvent() *RelationshipEvent { return r.Definition }
func (r RelationshipParsedEvent[C]) GetConditionSeq() ottl.ConditionSequence[C] {
	return r.ConditionSeq
}

type EventsGroup[C any] struct {
	Events []ParsedEventInterface[C]
}

type ParsedEvent[T any, C any] struct {
	Definition   *T
	ConditionSeq ottl.ConditionSequence[C]
}

type ProcessingContext struct {
	metricParser ottl.Parser[ottlmetric.TransformContext]
	logParser    ottl.Parser[ottllog.TransformContext]
	settings     component.TelemetrySettings
}

// createParsedEvents initializes and returns a ParsedEvents structure containing parsed entity and relationship events.
// It exists to parse ottl conditions for entity and relationship events at the time of creation, allowing for efficient evaluation later.
func createParsedEvents(s Schema, settings component.TelemetrySettings) (ParsedEvents, error) {
	ctx := ProcessingContext{settings: settings}

	var err error
	ctx.metricParser, err = ottlmetric.NewParser(ottlfuncs.StandardConverters[ottlmetric.TransformContext](), settings)
	if err != nil {
		return ParsedEvents{}, fmt.Errorf("failed to create parser for metric events: %w", err)
	}
	ctx.logParser, err = ottllog.NewParser(ottlfuncs.StandardConverters[ottllog.TransformContext](), settings)
	if err != nil {
		return ParsedEvents{}, fmt.Errorf("failed to create parser for log events: %w", err)
	}

	var result ParsedEvents

	// Process entity events
	for _, event := range s.Events.Entities {
		err = processEntityEvent(event, event.Context, event.Conditions, &result.MetricEvents, &result.LogEvents, ctx)
		if err != nil {
			return ParsedEvents{}, err
		}
	}

	// Process relationship events
	for _, event := range s.Events.Relationships {
		err = processRelationshipEvent(event, event.Context, event.Conditions, &result.MetricEvents, &result.LogEvents, ctx)
		if err != nil {
			return ParsedEvents{}, err
		}
	}

	return result, nil
}

func processEntityEvent(
	event EntityEvent,
	context string,
	conditions []string,
	metricEvents *EventsGroup[ottlmetric.TransformContext],
	logEvents *EventsGroup[ottllog.TransformContext],
	ctx ProcessingContext) error {
	if len(conditions) == 0 {
		conditions = []string{"true"}
	}

	switch context {
	case ottlmetric.ContextName:
		return addEntityEvent(metricEvents, event, conditions, ctx.metricParser, ctx.settings)
	case ottllog.ContextName:
		return addEntityEvent(logEvents, event, conditions, ctx.logParser, ctx.settings)
	default:
		return fmt.Errorf("unsupported context: %s", context)
	}
}

func processRelationshipEvent(
	event RelationshipEvent,
	context string,
	conditions []string,
	metricEvents *EventsGroup[ottlmetric.TransformContext],
	logEvents *EventsGroup[ottllog.TransformContext],
	ctx ProcessingContext) error {
	if len(conditions) == 0 {
		conditions = []string{"true"}
	}

	switch context {
	case ottlmetric.ContextName:
		return addRelationshipEvent(metricEvents, event, conditions, ctx.metricParser, ctx.settings)
	case ottllog.ContextName:
		return addRelationshipEvent(logEvents, event, conditions, ctx.logParser, ctx.settings)
	default:
		return fmt.Errorf("unsupported context: %s", context)
	}
}

func addEntityEvent[C any](
	eventsGroup *EventsGroup[C],
	event EntityEvent,
	conditions []string,
	parser ottl.Parser[C],
	settings component.TelemetrySettings) error {
	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return fmt.Errorf("failed to parse conditions for entity event: %w", err)
	}
	seq := ottl.NewConditionSequence(stmts, settings)
	eventsGroup.Events = append(eventsGroup.Events, EntityParsedEvent[C]{
		Definition:   &event,
		ConditionSeq: seq,
	})
	return nil
}

func addRelationshipEvent[C any](
	eventsGroup *EventsGroup[C],
	event RelationshipEvent,
	conditions []string,
	parser ottl.Parser[C],
	settings component.TelemetrySettings) error {
	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return fmt.Errorf("failed to parse conditions for relationship event: %w", err)
	}
	seq := ottl.NewConditionSequence(stmts, settings)
	eventsGroup.Events = append(eventsGroup.Events, RelationshipParsedEvent[C]{
		Definition:   &event,
		ConditionSeq: seq,
	})
	return nil
}
