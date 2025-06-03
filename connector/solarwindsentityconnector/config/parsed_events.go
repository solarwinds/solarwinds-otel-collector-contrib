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
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
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

type ParsedRelationshipEvent[C any] struct {
	Definition *RelationshipEvent
	Conditions []*ottl.Condition[C]
}

type ParsedEntityEvent[C any] struct {
	Definition *EntityEvent
	Conditions []*ottl.Condition[C]
}

// CreateParsedEvents initializes and returns a ParsedEvents structure containing parsed entity and relationship events.
// It exist to patse ottl conditions for entity and relationship events at the time of creation, allowing for efficient evaluation later.
func CreateParsedEvents(s Schema, settings component.TelemetrySettings) ParsedEvents {
	metricParser, metErr := ottlmetric.NewParser(nil, settings)
	logParser, logErr := ottllog.NewParser(nil, settings)
	if metErr != nil || logErr != nil {
		panic("failed to create parser for events")
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
			addEntityEvent(&metricGroup, event)
		case ottllog.ContextName:
			addEntityEvent(&logGroup, event)
		}
	}

	for _, event := range s.Events.Relationships {
		if len(event.Conditions) == 0 {
			event.Conditions = []string{"true"}
		}
		switch event.Context {
		case ottlmetric.ContextName:
			addRelationshipEvent(&metricGroup, event)
		case ottllog.ContextName:
			addRelationshipEvent(&logGroup, event)
		}
	}

	return ParsedEvents{
		MetricEvents: metricGroup,
		LogEvents:    logGroup,
	}
}

func addEntityEvent[C any](group *EventsGroup[C], event EntityEvent) {
	stmts, err := group.Parser.ParseConditions(event.Conditions)
	if err != nil {
		panic("failed to parse conditions for entity event")
	}
	group.Entities = append(group.Entities, ParsedEntityEvent[C]{
		Definition: &event,
		Conditions: stmts,
	})
}

func addRelationshipEvent[C any](group *EventsGroup[C], event RelationshipEvent) {
	stmts, err := group.Parser.ParseConditions(event.Conditions)
	if err != nil {
		panic("failed to parse conditions for relationship event")
	}
	group.Relationships = append(group.Relationships, ParsedRelationshipEvent[C]{
		Definition: &event,
		Conditions: stmts,
	})
}
