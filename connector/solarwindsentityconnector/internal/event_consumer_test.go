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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestSendExpiredEventsWithRelationship(t *testing.T) {
	// Create a sink to capture logs
	sink := new(consumertest.LogsSink)

	// Create event consumer with the sink
	eventConsumer := NewEventConsumer(sink)

	// Create test events
	events := []Event{
		&Relationship{
			Type: "dependsOn",
			Source: Entity{
				Type: "service",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("service_id", "service-123")
					return m
				}(),
			},
			Destination: Entity{
				Type: "database",
				IDs: func() pcommon.Map {
					m := pcommon.NewMap()
					m.PutStr("db_id", "db-456")
					return m
				}(),
			},
			Attributes: func() pcommon.Map {
				m := pcommon.NewMap()
				m.PutStr("connection_type", "primary")
				m.PutDouble("latency_ms", 15.5)
				return m
			}(),
		},
	}

	// Send the expired events
	eventConsumer.SendExpiredEvents(context.Background(), events)

	// Verify logs were received
	allLogs := sink.AllLogs()
	require.Equal(t, 1, len(allLogs), "Should have received exactly one batch of logs")

	// Verify Scope attribute
	scopeAttrs := allLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).Scope().Attributes()
	assert.Equal(t, 1, scopeAttrs.Len(), "Scope attributes should have one entry")
	scopeAttr1, _ := scopeAttrs.Get("otel.entity.event_as_log")
	assert.Equal(t, "true", scopeAttr1.AsString())

	// Verify LogRecords attributes
	attributes := allLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes()
	assert.Equal(t, attributes.Len(), 6)
	value1, _ := attributes.Get("otel.entity.event.type")
	assert.Equal(t, "entity_relationship_delete", value1.AsString())

	value2, _ := attributes.Get("otel.entity_relationship.source_entity.id")
	assert.Equal(t, "{\"service_id\":\"service-123\"}", value2.AsString())

	value3, _ := attributes.Get("otel.entity_relationship.destination_entity.id")
	assert.Equal(t, "{\"db_id\":\"db-456\"}", value3.AsString())

	value4, _ := attributes.Get("otel.entity_relationship.type")
	assert.Equal(t, "dependsOn", value4.AsString())

	value5, _ := attributes.Get("otel.entity_relationship.source_entity.type")
	assert.Equal(t, "service", value5.AsString())

	value6, _ := attributes.Get("otel.entity_relationship.destination_entity.type")
	assert.Equal(t, "database", value6.AsString())
}

func TestSendExpiredEventsWithEntity(t *testing.T) {
	// Create a sink to capture logs
	sink := new(consumertest.LogsSink)

	// Create event consumer with the sink
	eventConsumer := NewEventConsumer(sink)

	// Create test events
	events := []Event{
		&Entity{
			Type: "service",
			IDs: func() pcommon.Map {
				m := pcommon.NewMap()
				m.PutStr("service_id", "service-abc-123")
				m.PutStr("instance_id", "i-987xyz")
				return m
			}(),
			Attributes: func() pcommon.Map {
				m := pcommon.NewMap()
				m.PutStr("environment", "production")
				return m
			}(),
			Action: entityDeleteEventType,
		},
	}

	// Send the expired events
	eventConsumer.SendExpiredEvents(context.Background(), events)

	// Verify logs were received
	allLogs := sink.AllLogs()
	require.Equal(t, 1, len(allLogs), "Should have received exactly one batch of logs")

	// Verify Scope attribute
	scopeAttrs := allLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).Scope().Attributes()
	assert.Equal(t, 1, scopeAttrs.Len(), "Scope attributes should have one entry")

	// Verify LogRecords
	logRecords := allLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords()
	assert.Equal(t, 1, logRecords.Len(), "Should have received exactly one logRecords")

	// Verify LogRecords attributes
	attributes := allLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes()
	assert.Equal(t, attributes.Len(), 5)
	value1, _ := attributes.Get("otel.entity.event.type")
	assert.Equal(t, "entity_delete", value1.AsString())

	value2, _ := attributes.Get("otel.entity.delete.type")
	assert.Equal(t, "soft_delete", value2.AsString())

	value3, _ := attributes.Get("otel.entity.type")
	assert.Equal(t, "service", value3.AsString())

	value4, _ := attributes.Get("otel.entity.id")
	assert.Equal(t, "{\"instance_id\":\"i-987xyz\",\"service_id\":\"service-abc-123\"}", value4.AsString())

	value5, _ := attributes.Get("otel.entity.attributes")
	assert.Equal(t, "{\"environment\":\"production\"}", value5.AsString())
}
