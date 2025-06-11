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
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal/storage"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type EventBuilder struct {
	entitiesDefinitions map[string]config.Entity
	sourcePrefix        string
	destPrefix          string
	logger              *zap.Logger

	storageManager *storage.Manager
}

func NewEventBuilder(entities map[string]config.Entity, sourcePrefix string, destPrefix string, logger *zap.Logger) *EventBuilder {
	return &EventBuilder{
		entitiesDefinitions: entities,
		sourcePrefix:        sourcePrefix,
		destPrefix:          destPrefix,
		logger:              logger,
	}
}

func (e *EventBuilder) AppendUpdateEvent(eventLogs *plog.LogRecordSlice, subject Subject) {
	lr := eventLogs.AppendEmpty()
	lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	subject.Update(&lr)

}

func (e *EventBuilder) AppendEntityUpdateEvent(eventLogs *plog.LogRecordSlice, entity config.Entity, resourceAttrs pcommon.Map) {
	entityLog, err := e.createEntityEvent(resourceAttrs, entity)
	if err != nil {
		e.logger.Debug("failed to create update event", zap.Error(err))
		return
	}

	entityLog.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	entityLog.Attributes().PutStr(entityEventType, entityUpdateEventType)
	eventLog := eventLogs.AppendEmpty()
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

func (e *EventBuilder) AppendRelationshipUpdateEvent(eventLogs *plog.LogRecordSlice, relationship config.RelationshipEvent, resourceAttrs pcommon.Map, sm *storage.Manager) {
	relationshipLog, err := e.createRelationshipEvent(relationship, resourceAttrs)
	if err != nil {
		e.logger.Debug("failed to create relationship event", zap.Error(err))
		return
	}

	relationshipLog.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	relationshipLog.Attributes().PutStr(entityEventType, relationshipUpdateEventType)
	eventLog := eventLogs.AppendEmpty()
	relationshipLog.CopyTo(eventLog)

	srcIds, _ := relationshipLog.Attributes().Get(relationshipSrcEntityIds)
	destIds, _ := relationshipLog.Attributes().Get(relationshipDestEntityIds)
	sm.Update(relationship, srcIds.Map(), destIds.Map())
}

func (e *EventBuilder) createRelationshipEvent(relationship config.RelationshipEvent, resourceAttrs pcommon.Map) (plog.LogRecord, error) {
	source, ok := e.entitiesDefinitions[relationship.Source]
	if !ok {
		return plog.NewLogRecord(), fmt.Errorf("bad source entity")
	}

	dest, ok := e.entitiesDefinitions[relationship.Destination]
	if !ok {
		return plog.NewLogRecord(), fmt.Errorf("bad destination entity")
	}

	lr := plog.NewLogRecord()
	attrs := lr.Attributes()

	if source.Type == dest.Type {
		if err := e.setAttributesForSameTypeRelationships(attrs, source, dest, resourceAttrs); err != nil {
			return plog.NewLogRecord(), err
		}

	} else {
		if err := e.setAttributesForDifferentTypeRelationships(attrs, source, dest, resourceAttrs); err != nil {
			return plog.NewLogRecord(), err
		}
	}

	setAttributes(attrs, relationship.Attributes, resourceAttrs, relationshipAttributes)
	attrs.PutStr(relationshipType, relationship.Type)
	attrs.PutStr(srcEntityType, source.Type)
	attrs.PutStr(destEntityType, dest.Type)

	return lr, nil
}

func (e *EventBuilder) setAttributesForSameTypeRelationships(attrs pcommon.Map, source config.Entity, dest config.Entity, resourceAttrs pcommon.Map) error {
	if e.sourcePrefix == "" || e.destPrefix == "" {
		return fmt.Errorf("prefixes are mandatory for same type relationships")
	}

	hasPrefixSrc, err := setIdAttributesForRelationships(attrs, source.IDs, resourceAttrs, relationshipSrcEntityIds, e.sourcePrefix)
	if err != nil || !hasPrefixSrc {
		return fmt.Errorf("missing prefixed ID attribute for source entity")
	}

	hasPrefixDst, err := setIdAttributesForRelationships(attrs, dest.IDs, resourceAttrs, relationshipDestEntityIds, e.destPrefix)
	if err != nil || !hasPrefixDst {
		return fmt.Errorf("missing prefixed ID attribute for destination entity")
	}
	return nil
}

func (e *EventBuilder) setAttributesForDifferentTypeRelationships(attrs pcommon.Map, source config.Entity, dest config.Entity, resourceAttrs pcommon.Map) error {
	// For different type relationships, prefixes are optional.
	_, err := setIdAttributesForRelationships(attrs, source.IDs, resourceAttrs, relationshipSrcEntityIds, e.sourcePrefix)
	if err != nil {
		return fmt.Errorf("missing ID attribute for source entity")
	}

	_, err = setIdAttributesForRelationships(attrs, dest.IDs, resourceAttrs, relationshipDestEntityIds, e.destPrefix)
	if err != nil {
		return fmt.Errorf("missing ID attribute for destination entity")
	}
	return nil
}
