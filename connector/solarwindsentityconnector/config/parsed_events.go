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

type EventsGroup[C any] struct {
	Entities      []ParsedEntityEvent[C]
	Relationships []ParsedRelationshipEvent[C]
	Parser        *ottl.Parser[C]
}

type ParsedEvent[T any, C any] struct {
	Definition   *T
	ConditionSeq ottl.ConditionSequence[C]
}

type ParsedRelationshipEvent[C any] = ParsedEvent[RelationshipEvent, C]
type ParsedEntityEvent[C any] = ParsedEvent[EntityEvent, C]

// createParsedEvents initializes and returns a ParsedEvents structure containing parsed entity and relationship events.
// It exist to patse ottl conditions for entity and relationship events at the time of creation, allowing for efficient evaluation later.
func createParsedEvents(s Schema, settings component.TelemetrySettings) (ParsedEvents, error) {
	metricParser, err := ottlmetric.NewParser(ottlfuncs.StandardConverters[ottlmetric.TransformContext](), settings)
	if err != nil {
		return ParsedEvents{}, fmt.Errorf("failed to create parser for metric events %w", err)
	}
	logParser, err := ottllog.NewParser(ottlfuncs.StandardConverters[ottllog.TransformContext](), settings)
	if err != nil {
		return ParsedEvents{}, fmt.Errorf("failed to create parser for log events %w", err)
	}

	metricGroup := EventsGroup[ottlmetric.TransformContext]{
		Entities:      []ParsedEntityEvent[ottlmetric.TransformContext]{},
		Relationships: []ParsedRelationshipEvent[ottlmetric.TransformContext]{},
		Parser:        &metricParser,
	}

	logGroup := EventsGroup[ottllog.TransformContext]{
		Entities:      []ParsedEntityEvent[ottllog.TransformContext]{},
		Relationships: []ParsedRelationshipEvent[ottllog.TransformContext]{},
		Parser:        &logParser,
	}

	for _, event := range s.Events.Entities {
		if len(event.Conditions) == 0 {
			event.Conditions = []string{"true"}
		}
		switch event.Context {
		case ottlmetric.ContextName:
			err = addEntityEvent(&metricGroup, event, settings)
			if err != nil {
				return ParsedEvents{}, err
			}
		case ottllog.ContextName:
			err = addEntityEvent(&logGroup, event, settings)
			if err != nil {
				return ParsedEvents{}, err
			}
		}
	}

	for _, event := range s.Events.Relationships {
		if len(event.Conditions) == 0 {
			event.Conditions = []string{"true"}
		}
		switch event.Context {
		case ottlmetric.ContextName:
			err = addRelationshipEvent(&metricGroup, event, settings)
			if err != nil {
				return ParsedEvents{}, err
			}
		case ottllog.ContextName:
			err = addRelationshipEvent(&logGroup, event, settings)
			if err != nil {
				return ParsedEvents{}, err
			}
		}
	}

	return ParsedEvents{
		MetricEvents: metricGroup,
		LogEvents:    logGroup,
	}, nil
}

func addEntityEvent[C any](group *EventsGroup[C], event EntityEvent, settings component.TelemetrySettings) error {
	stmts, err := group.Parser.ParseConditions(event.Conditions)
	seq := ottl.NewConditionSequence(stmts, settings)
	if err != nil {
		return fmt.Errorf("failed to parse conditions for entity event: %w", err)
	}
	group.Entities = append(group.Entities, ParsedEntityEvent[C]{
		Definition:   &event,
		ConditionSeq: seq,
	})
	return nil
}

func addRelationshipEvent[C any](group *EventsGroup[C], event RelationshipEvent, settings component.TelemetrySettings) error {
	stmts, err := group.Parser.ParseConditions(event.Conditions)
	seq := ottl.NewConditionSequence(stmts, settings)
	if err != nil {
		return fmt.Errorf("failed to parse conditions for relationship event: %w", err)
	}
	group.Relationships = append(group.Relationships, ParsedRelationshipEvent[C]{
		Definition:   &event,
		ConditionSeq: seq,
	})
	return nil
}
