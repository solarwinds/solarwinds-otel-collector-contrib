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
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestConditionTrue(t *testing.T) {
	// Initialize test data
	ctx := context.Background()
	resourceAttrs := pcommon.NewMap()
	settings := componenttest.NewNopTelemetrySettings()
	logs := plog.NewLogs()

	parserLog, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parserLog.ParseConditions([]string{"true"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entityEvent := config.ParsedEntityEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Type: "test-entity"},
		ConditionSeq: seq,
	}
	relationshipEvent := config.ParsedRelationshipEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: "test-rel"},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Entities:      []config.ParsedEntityEvent[ottllog.TransformContext]{entityEvent},
		Relationships: []config.ParsedRelationshipEvent[ottllog.TransformContext]{relationshipEvent},
		Parser:        &parserLog,
	}

	entitiesDefinitions := map[string]config.Entity{
		"test-entity": {Type: "test-entity"},
	}
	eventBuilder := NewEventBuilder(entitiesDefinitions, "source", "destination", &logs, componenttest.NewNopTelemetrySettings().Logger)

	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	// Tested function
	err = ProcessEvents(ctx, eventBuilder, eventsGroup, resourceAttrs, tc)
	require.NoError(t, err)

}

func TestConditionFalse(t *testing.T) {
	// Initialize test data
	ctx := context.Background()
	resourceAttrs := pcommon.NewMap()
	settings := componenttest.NewNopTelemetrySettings()
	logs := plog.NewLogs()

	parserLog, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)
	stmts, err := parserLog.ParseConditions([]string{"false"})
	require.NoError(t, err)
	seq := ottl.NewConditionSequence(stmts, settings)

	entityEvent := config.ParsedEntityEvent[ottllog.TransformContext]{
		Definition:   &config.EntityEvent{Type: "test-entity"},
		ConditionSeq: seq,
	}
	relationshipEvent := config.ParsedRelationshipEvent[ottllog.TransformContext]{
		Definition:   &config.RelationshipEvent{Type: "test-rel"},
		ConditionSeq: seq,
	}
	eventsGroup := config.EventsGroup[ottllog.TransformContext]{
		Entities:      []config.ParsedEntityEvent[ottllog.TransformContext]{entityEvent},
		Relationships: []config.ParsedRelationshipEvent[ottllog.TransformContext]{relationshipEvent},
		Parser:        &parserLog,
	}

	entitiesDefinitions := map[string]config.Entity{
		"test-entity": {Type: "test-entity"},
	}
	eventBuilder := NewEventBuilder(entitiesDefinitions, "source", "destination", &logs, componenttest.NewNopTelemetrySettings().Logger)

	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	// Tested function
	err = ProcessEvents(ctx, eventBuilder, eventsGroup, resourceAttrs, tc)
	require.NoError(t, err)

}
