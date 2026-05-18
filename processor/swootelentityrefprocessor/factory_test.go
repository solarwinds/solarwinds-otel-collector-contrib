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

package swootelentityrefprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestFactory_CreateLogsProcessor_InvalidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{Action: "unsupported"}
	_, err := factory.CreateLogs(
		context.Background(),
		processortest.NewNopSettings(factory.Type()),
		cfg,
		consumertest.NewNop(),
	)
	require.Error(t, err)
}

func TestFactory_CreateMetricsProcessor_InvalidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{Action: "unsupported"}
	_, err := factory.CreateMetrics(
		context.Background(),
		processortest.NewNopSettings(factory.Type()),
		cfg,
		consumertest.NewNop(),
	)
	require.Error(t, err)
}

func TestFactory_CreateTracesProcessor_InvalidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{Action: "unsupported"}
	_, err := factory.CreateTraces(
		context.Background(),
		processortest.NewNopSettings(factory.Type()),
		cfg,
		consumertest.NewNop(),
	)
	require.Error(t, err)
}
