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

// ParsedEvents holds pre-parsed event configurations for both metric and log contexts.
// This structure allows for efficient event evaluation by parsing OTTL conditions once at initialization.
type ParsedEvents struct {
	MetricEvents []ParsedEventInterface[ottlmetric.TransformContext]
	LogEvents    []ParsedEventInterface[ottllog.TransformContext]
}

// createParsedEvents initializes and returns a ParsedEvents structure containing parsed entity and relationship events.
// It exists to parse ottl conditions for entity and relationship events at the time of creation, allowing for efficient evaluation later.
func createParsedEvents(s Schema, settings component.TelemetrySettings) (ParsedEvents, error) {
	// Create parsers for different contexts
	metricParser, err := ottlmetric.NewParser(ottlfuncs.StandardConverters[ottlmetric.TransformContext](), settings)
	if err != nil {
		return ParsedEvents{}, fmt.Errorf("failed to create parser for metric events: %w", err)
	}
	logParser, err := ottllog.NewParser(ottlfuncs.StandardConverters[ottllog.TransformContext](), settings)
	if err != nil {
		return ParsedEvents{}, fmt.Errorf("failed to create parser for log events: %w", err)
	}

	var result ParsedEvents

	// Process entity events
	for _, event := range s.Events.Entities {
		err = processEntityEvent(event, &result.MetricEvents, &result.LogEvents, metricParser, logParser, settings)
		if err != nil {
			return ParsedEvents{}, fmt.Errorf("failed to process entity event: %w", err)
		}
	}

	// Process relationship events
	for _, event := range s.Events.Relationships {
		err = processRelationshipEvent(event, &result.MetricEvents, &result.LogEvents, metricParser, logParser, settings)
		if err != nil {
			return ParsedEvents{}, fmt.Errorf("failed to process relationship event: %w", err)
		}
	}

	return result, nil
}

// processEntityEvent handles entity event processing with type safety
func processEntityEvent(
	event EntityEvent,
	metricEvents *[]ParsedEventInterface[ottlmetric.TransformContext],
	logEvents *[]ParsedEventInterface[ottllog.TransformContext],
	metricParser ottl.Parser[ottlmetric.TransformContext],
	logParser ottl.Parser[ottllog.TransformContext],
	settings component.TelemetrySettings) error {

	conditions := event.Conditions
	if len(conditions) == 0 {
		conditions = []string{"true"}
	}

	switch event.Context {
	case ottlmetric.ContextName:
		return addEntityEvent(metricEvents, event, conditions, metricParser, settings)
	case ottllog.ContextName:
		return addEntityEvent(logEvents, event, conditions, logParser, settings)
	default:
		return fmt.Errorf("unsupported context: %s", event.Context)
	}
}

// processRelationshipEvent handles relationship event processing with type safety
func processRelationshipEvent(
	event RelationshipEvent,
	metricEvents *[]ParsedEventInterface[ottlmetric.TransformContext],
	logEvents *[]ParsedEventInterface[ottllog.TransformContext],
	metricParser ottl.Parser[ottlmetric.TransformContext],
	logParser ottl.Parser[ottllog.TransformContext],
	settings component.TelemetrySettings) error {

	conditions := event.Conditions
	if len(conditions) == 0 {
		conditions = []string{"true"}
	}

	switch event.Context {
	case ottlmetric.ContextName:
		return addRelationshipEvent(metricEvents, event, conditions, metricParser, settings)
	case ottllog.ContextName:
		return addRelationshipEvent(logEvents, event, conditions, logParser, settings)
	default:
		return fmt.Errorf("unsupported context: %s", event.Context)
	}
}

// addEntityEvent adds a parsed entity event to the events slice
func addEntityEvent[C any](
	events *[]ParsedEventInterface[C],
	event EntityEvent,
	conditions []string,
	parser ottl.Parser[C],
	settings component.TelemetrySettings) error {

	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return fmt.Errorf("failed to parse conditions for entity event: %w", err)
	}

	*events = append(*events, EntityParsedEvent[C]{
		Definition:   &event,
		ConditionSeq: ottl.NewConditionSequence(stmts, settings),
	})
	return nil
}

// addRelationshipEvent adds a parsed relationship event to the events slice
func addRelationshipEvent[C any](
	events *[]ParsedEventInterface[C],
	event RelationshipEvent,
	conditions []string,
	parser ottl.Parser[C],
	settings component.TelemetrySettings) error {

	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return fmt.Errorf("failed to parse conditions for relationship event: %w", err)
	}

	*events = append(*events, RelationshipParsedEvent[C]{
		Definition:   &event,
		ConditionSeq: ottl.NewConditionSequence(stmts, settings),
	})
	return nil
}
