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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type mockExporter struct {
	pushed []pmetric.Metrics
}

func newMockExporter() *mockExporter {
	return &mockExporter{
		pushed: []pmetric.Metrics{},
	}
}

func (m *mockExporter) start(_ context.Context, _ component.Host) error {
	return nil
}

func (m *mockExporter) shutdown(_ context.Context) error {
	return nil
}

func (m *mockExporter) push(_ context.Context, metrics pmetric.Metrics) error {
	m.pushed = append(m.pushed, metrics)

	return nil
}

// TestHeartbeatEmittingMetrics runs the `Hearthbeat`
// in isolation with a mocked exporter and inspects
// emitted metrics.
func TestHeartbeatEmittingMetrics(t *testing.T) {
	const (
		testDuration  = 1000 * time.Millisecond
		beatInterval  = 100 * time.Millisecond
		expectedCount = int(testDuration / beatInterval)
	)

	mockExp := newMockExporter()
	hb := newHeartbeatWithExporter(
		extensiontest.NewNopSettings(),
		&Config{},
		mockExp,
	)
	// Adjust the heartbeat interval to shave off some time.
	hb.beatInterval = beatInterval

	// Start the heartbeat.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := hb.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)

	// Inspect exported metrics in a loop.
	assert.Eventuallyf(
		t,
		func() bool {
			return len(mockExp.pushed) == expectedCount
		},
		testDuration+50*time.Millisecond, // Allow some leeway.
		10*time.Millisecond,
		"expected %d metrics, got %d",
		expectedCount,
		len(mockExp.pushed),
	)

	err = hb.Shutdown(ctx)
	require.NoError(t, err)
}
