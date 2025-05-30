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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

// processCondition evaluates the conditions for entities and relationships.
// If the conditions are met, it appends the corresponding entity or relationship update event to the event builder.
func ProcessEvents[C any](
	ctx context.Context,
	eventBuilder *EventBuilder,
	contextEntityEvents []config.EntityEvent,
	contextRelationshipEvents []config.RelationshipEvent,
	resourceAttrs pcommon.Map,
	logger *zap.Logger,
	entityDefinitions map[string]config.Entity,
	telemetrySettings component.TelemetrySettings,
	parser *ottl.Parser[C],
	tc C) error {

	for _, entityEvent := range contextEntityEvents {
		condition := entityEvent.Conditions
		ok, err := evaluateConditions(ctx, telemetrySettings, condition, parser, tc)
		if err != nil {
			logger.Error("Failed to evaluate entity condition", zap.Error(err))
		}
		if ok {
			entity := entityDefinitions[entityEvent.Type]
			eventBuilder.AppendEntityUpdateEvent(entity, resourceAttrs)
		}
	}

	for _, relationshipEvent := range contextRelationshipEvents {
		condition := relationshipEvent.Conditions
		ok, err := evaluateConditions(ctx, telemetrySettings, condition, parser, tc)
		if err != nil {
			logger.Error("Failed to evaluate relationship condition", zap.Error(err))
			return err
		}

		if ok {
			eventBuilder.AppendRelationshipUpdateEvent(relationshipEvent, resourceAttrs)
		}
	}
	return nil
}

// evaluateConditions evaluates the provided conditions using the given parser and telemetry settings.
// It returns true if the conditions are met, otherwise false.
func evaluateConditions[C any](
	ctx context.Context,
	telemetrySettings component.TelemetrySettings,
	condition []string,
	parser *ottl.Parser[C],
	tc C,
) (bool, error) {
	seq, err := newConditionSeq(
		*parser,
		telemetrySettings,
		condition,
	)

	if err != nil {
		return false, err
	}

	ok, err := seq.Eval(ctx, tc)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// newConditionSeq creates a new condition sequence from the provided conditions.
// If no conditions are provided, it defaults to a sequence that always evaluates to true.
func newConditionSeq[C any](
	parser ottl.Parser[C],
	settings component.TelemetrySettings,
	conditions []string,
) (*ottl.ConditionSequence[C], error) {
	if len(conditions) == 0 {
		conditions = []string{"true"}
	}

	stmts, err := parser.ParseConditions(conditions)
	if err != nil {
		return nil, err
	}

	seq := ottl.NewConditionSequence(stmts, settings)
	return &seq, nil
}
