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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestEvaluateConditionsTrueCondition(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()

	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)

	ctx := context.Background()
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	ok, err := evaluateConditions(ctx, settings, []string{"true"}, &parser, tc)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestEvaluateConditionsFalseCondition(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)

	ctx := context.Background()
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	ok, err := evaluateConditions(ctx, settings, []string{"false"}, &parser, tc)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestEvaluateConditionsEmptyCondition(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)

	ctx := context.Background()
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	ok, err := evaluateConditions(ctx, settings, []string{}, &parser, tc)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestEvaluateConditionsInvalidCondition(t *testing.T) {
	settings := componenttest.NewNopTelemetrySettings()
	parser, err := ottllog.NewParser(nil, settings)
	require.NoError(t, err)

	ctx := context.Background()
	logs := plog.NewLogs()
	rLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := rLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	resource := rLogs.Resource()
	scope := scopeLogs.Scope()
	tc := ottllog.NewTransformContext(logRecord, scope, resource, scopeLogs, rLogs)

	ok, err := evaluateConditions(ctx, settings, []string{"invalid syntax"}, &parser, tc)
	require.Error(t, err)
	require.False(t, ok)
}
