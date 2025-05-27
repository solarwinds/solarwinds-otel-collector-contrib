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
	"fmt"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type EventBuilder struct {
	entities      map[string]config.Entity
	relationships []config.RelationshipEvent
	sourcePrefix  string
	destPrefix    string
	eventLogs     *plog.LogRecordSlice
	logger        *zap.Logger
}

func NewEventBuilder(entities map[string]config.Entity, relationships []config.RelationshipEvent, sourcePrefix string, destPrefix string, events *plog.Logs, logger *zap.Logger) *EventBuilder {
	return &EventBuilder{
		entities:      entities,
		relationships: relationships,
		sourcePrefix:  sourcePrefix,
		destPrefix:    destPrefix,
		eventLogs:     createEventLog(events),
		logger:        logger,
	}
}

// createResultEventLog prepares a clean LogRecordSlice, where log records representing events should be appended.
// Creates a resource log in input plog.Logs with single scope log decorated with attributes necessary for proper SWO ingestion.
func createEventLog(logs *plog.Logs) *plog.LogRecordSlice {
	resourceLog := logs.ResourceLogs().AppendEmpty()
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().Attributes().PutBool(entityEventAsLog, true)
	lrs := scopeLog.LogRecords()

	return &lrs
}

func (e *EventBuilder) AppendEntityUpdateEvent(entity config.Entity, resourceAttrs pcommon.Map) {

	entityLog, err := e.createEntityEvent(resourceAttrs, entity)
	if err != nil {
		e.logger.Debug("failed to create update event", zap.Error(err))
		return
	}

	entityLog.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	entityLog.Attributes().PutStr(entityEventType, entityUpdateEventType)
	eventLog := e.eventLogs.AppendEmpty()
	entityLog.CopyTo(eventLog)
}

func (e *EventBuilder) createEntityEvent(resourceAttrs pcommon.Map, entity config.Entity) (plog.LogRecord, error) {
	lr := plog.NewLogRecord()
	attrs := lr.Attributes()
	attrs.PutStr(entityType, entity.Type)

	if err := setIdAttributes(attrs, entity.IDs, resourceAttrs, entityIds); err != nil {
		return plog.LogRecord{}, fmt.Errorf("failed to set ID attributes: %w", err)
	}

	setAttributes(attrs, entity.Attributes, resourceAttrs, entityAttributes)

	lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	return lr, nil
}

func (e *EventBuilder) AppendRelationshipUpdateEvent(relationship config.RelationshipEvent, resourceAttrs pcommon.Map) {
	relationshipLog, err := e.createRelationshipEvent(relationship, resourceAttrs)
	if err != nil {
		e.logger.Debug("failed to create relationship event", zap.Error(err))
		return
	}

	relationshipLog.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	relationshipLog.Attributes().PutStr(entityEventType, relationshipUpdateEventType)
	eventLog := e.eventLogs.AppendEmpty()
	relationshipLog.CopyTo(eventLog)

}

func (e *EventBuilder) createRelationshipEvent(relationship config.RelationshipEvent, resourceAttrs pcommon.Map) (plog.LogRecord, error) {
	source, ok := e.entities[relationship.Source]
	if !ok {
		return plog.NewLogRecord(), fmt.Errorf("bad source entity")
	}

	dest, ok := e.entities[relationship.Destination]
	if !ok {
		return plog.NewLogRecord(), fmt.Errorf("bad destination entity")
	}

	lr := plog.NewLogRecord()
	attrs := lr.Attributes()

	if source.Type == dest.Type {
		// same type relationships
		hasPrefixSrc, err := setIdAttributesWithPrefix(attrs, source.IDs, resourceAttrs, relationshipSrcEntityIds, e.sourcePrefix)
		if err != nil || !hasPrefixSrc {
			return plog.NewLogRecord(), fmt.Errorf("missing prefixed ID attribute for source entity")
		}

		hasPrefixDst, err := setIdAttributesWithPrefix(attrs, dest.IDs, resourceAttrs, relationshipDestEntityIds, e.destPrefix)
		if err != nil || !hasPrefixDst {
			return plog.NewLogRecord(), fmt.Errorf("missing prefixed ID attribute for destination entity")
		}

	} else {

		if err := setIdAttributes(attrs, source.IDs, resourceAttrs, relationshipSrcEntityIds); err != nil {
			return plog.NewLogRecord(), fmt.Errorf("missing ID attribute for source entity")
		}

		if err := setIdAttributes(attrs, dest.IDs, resourceAttrs, relationshipDestEntityIds); err != nil {
			return plog.NewLogRecord(), fmt.Errorf("missing ID attribute for destination entity")
		}
	}

	setAttributes(attrs, relationship.Attributes, resourceAttrs, relationshipAttributes)
	attrs.PutStr(relationshipType, relationship.Type)
	attrs.PutStr(srcEntityType, source.Type)
	attrs.PutStr(destEntityType, dest.Type)

	return lr, nil
}
