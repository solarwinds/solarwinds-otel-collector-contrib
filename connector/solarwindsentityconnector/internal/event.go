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
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

func AppendEntityUpdateEvent(lrs *plog.LogRecordSlice, entity config.Entity, resourceAttrs pcommon.Map) {
	event, err := CreateEntityEvent(resourceAttrs, entity)
	if err != nil {
		zap.L().Debug("failed to create update event", zap.Error(err))
		return
	}
	event.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	event.Attributes().PutStr(entityEventType, entityUpdateEventType)
	eventLog := lrs.AppendEmpty()
	event.CopyTo(eventLog)
}

func AppendRelationshipUpdateEvent(lrs *plog.LogRecordSlice, relationship config.Relationship, resourceAttrs pcommon.Map, entities map[string]config.Entity) {
	src, ok := entities[relationship.Source]
	if !ok {
		zap.L().Debug("source entity not found", zap.String("entity", relationship.Source))
		return
	}

	dest, ok := entities[relationship.Destination]
	if !ok {
		zap.L().Debug("destination entity not found", zap.String("entity", relationship.Destination))
		return
	}

	event, err := CreateRelationshipEvent(resourceAttrs, relationship, src, dest)
	if err != nil {
		zap.L().Debug("failed to create relationship update event", zap.Error(err))
		return
	}

	event.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	event.Attributes().PutStr(entityEventType, relationshipUpdateEventType)

	eventLog := lrs.AppendEmpty()
	event.CopyTo(eventLog)
}
