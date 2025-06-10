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
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/storage"
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// ProcessEvents evaluates the conditions for entities and relationships events.
// If the conditions are met, it appends the corresponding entity or relationship update event to the event builder.
// Multiple condition items are evaluated using OR logic.
func ProcessEvents[C any](
	ctx context.Context,
	result *plog.LogRecordSlice,
	eventBuilder *EventBuilder,
	events config.EventsGroup[C],
	resourceAttrs pcommon.Map,
	sm *storage.Manager,
	tc C) error {

	for _, entityEvent := range events.Entities {
		ok, err := entityEvent.ConditionSeq.Eval(ctx, tc)
		if err != nil {
			return err
		}

		if ok {
			entity := eventBuilder.entitiesDefinitions[entityEvent.Definition.Type]
			eventBuilder.AppendEntityUpdateEvent(result, entity, resourceAttrs)
		}
	}

	for _, relationshipEvent := range events.Relationships {
		ok, err := relationshipEvent.ConditionSeq.Eval(ctx, tc)
		if err != nil {
			return err
		}

		if ok {
			eventBuilder.AppendRelationshipUpdateEvent(result, *relationshipEvent.Definition, resourceAttrs, sm)
		}
	}
	return nil
}
