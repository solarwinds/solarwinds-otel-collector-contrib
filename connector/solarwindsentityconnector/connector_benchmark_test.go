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

package solarwindsentityconnector

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/connector/connectortest"
	"go.opentelemetry.io/collector/consumer/consumertest"
)

func BenchmarkMetrics(b *testing.B) {
	testCases := []struct {
		name         string
		totalMetrics int
		rubbishRatio float32
		composite    bool
	}{
		// No composite metrics, rubbish ratio 0.0
		{name: "1000_metrics_0.0_false", totalMetrics: 1000, rubbishRatio: 0.0, composite: false},
		{name: "10000_metrics_0.0_false", totalMetrics: 10000, rubbishRatio: 0.0, composite: false},
		{name: "100000_metrics_0.0_false", totalMetrics: 100000, rubbishRatio: 0.0, composite: false},
		{name: "1000000_metrics_0.0_false", totalMetrics: 1000000, rubbishRatio: 0.0, composite: false},
		// No composite metrics, rubbish ratio 0.5
		{name: "1000_metrics_0.5_false", totalMetrics: 1000, rubbishRatio: 0.5, composite: false},
		{name: "10000_metrics_0.5_false", totalMetrics: 10000, rubbishRatio: 0.5, composite: false},
		{name: "100000_metrics_0.5_false", totalMetrics: 100000, rubbishRatio: 0.5, composite: false},
		{name: "1000000_metrics_0.5_false", totalMetrics: 1000000, rubbishRatio: 0.5, composite: false},
		// No composite metrics, rubbish ratio 1.0
		{name: "1000_metrics_1.0_false", totalMetrics: 1000, rubbishRatio: 1.0, composite: false},
		{name: "10000_metrics_1.0_false", totalMetrics: 10000, rubbishRatio: 1.0, composite: false},
		{name: "100000_metrics_1.0_false", totalMetrics: 100000, rubbishRatio: 1.0, composite: false},
		{name: "10000000_metrics_1.0_false", totalMetrics: 10000000, rubbishRatio: 1.0, composite: false},
		// Composite metrics, rubbish ratio 0.0
		{name: "1000_metrics_0.0_true", totalMetrics: 1000, rubbishRatio: 0.0, composite: true},
		{name: "10000_metrics_0.0_true", totalMetrics: 10000, rubbishRatio: 0.0, composite: true},
		{name: "100000_metrics_0.0_true", totalMetrics: 100000, rubbishRatio: 0.0, composite: true},
		{name: "1000000_metrics_0.0_true", totalMetrics: 1000000, rubbishRatio: 0.0, composite: true},
		// Composite metrics, rubbish ratio 0.5
		{name: "1000_metrics_0.5_true", totalMetrics: 1000, rubbishRatio: 0.5, composite: true},
		{name: "10000_metrics_0.5_true", totalMetrics: 10000, rubbishRatio: 0.5, composite: true},
		{name: "100000_metrics_0.5_true", totalMetrics: 100000, rubbishRatio: 0.5, composite: true},
		{name: "1000000_metrics_0.5_true", totalMetrics: 1000000, rubbishRatio: 0.5, composite: true},
		// Composite metrics, rubbish ratio 1.0
		{name: "1000_metrics_1.0_true", totalMetrics: 1000, rubbishRatio: 1.0, composite: true},
		{name: "10000_metrics_1.0_true", totalMetrics: 10000, rubbishRatio: 1.0, composite: true},
		{name: "100000_metrics_1.0_true", totalMetrics: 100000, rubbishRatio: 1.0, composite: true},
		{name: "10000000_metrics_1.0_true", totalMetrics: 10000000, rubbishRatio: 1.0, composite: true},
	}

	for _, tc := range testCases {
		b.Run("ConsumeMetrics", func(b *testing.B) {
			cfg, err := loadConfigFromFileBenchmark(b, filepath.Join("testdata", "benchmark", "config.yaml"))
			require.NoError(b, err)
			factory := NewFactory()
			sink := &consumertest.LogsSink{}

			conn, err := factory.CreateMetricsToLogs(context.Background(),
				connectortest.NewNopSettings(metadata.Type), cfg, sink)
			require.NoError(b, err)

			require.NoError(b, conn.Start(context.Background(), componenttest.NewNopHost()))
			defer func() {
				assert.NoError(b, conn.Shutdown(context.Background()))
			}()

			metrics := genMetricsFromConfig(b, cfg, tc.totalMetrics, tc.rubbishRatio, tc.composite)

			b.ReportAllocs()
			b.ResetTimer()
			for _, metric := range metrics {
				assert.NoError(b, conn.ConsumeMetrics(context.Background(), metric))
			}
			b.StopTimer()

			getLogsCount := len(sink.AllLogs())
			expectedLogsCount := tc.totalMetrics - int(float32(tc.totalMetrics)*tc.rubbishRatio)
			if tc.composite { // If composite metrics are used, we expect only one log from the connector
				expectedLogsCount = 1
			}

			assert.Equal(b, getLogsCount, expectedLogsCount, "Expected to receive %d logs, but got %d logs",
				expectedLogsCount, getLogsCount)
		})
	}
}

func BenchmarkLogs(b *testing.B) {
	testCases := []struct {
		name         string
		totalLogs    int
		rubbishRatio float32
		composite    bool
	}{
		// No composite logs, rubbish ratio 0.0
		{name: "1000_logs_0.0_false", totalLogs: 1000, rubbishRatio: 0.0, composite: false},
		{name: "10000_logs_0.0_false", totalLogs: 10000, rubbishRatio: 0.0, composite: false},
		{name: "100000_logs_0.0_false", totalLogs: 100000, rubbishRatio: 0.0, composite: false},
		{name: "1000000_logs_0.0_false", totalLogs: 1000000, rubbishRatio: 0.0, composite: false},
		// No composite logs, rubbish ratio 0.5
		{name: "1000_logs_0.5_false", totalLogs: 1000, rubbishRatio: 0.5, composite: false},
		{name: "10000_logs_0.5_false", totalLogs: 10000, rubbishRatio: 0.5, composite: false},
		{name: "100000_logs_0.5_false", totalLogs: 100000, rubbishRatio: 0.5, composite: false},
		{name: "1000000_logs_0.5_false", totalLogs: 1000000, rubbishRatio: 0.5, composite: false},
		// No composite logs, rubbish ratio 1.0
		{name: "1000_logs_1.0_false", totalLogs: 1000, rubbishRatio: 1.0, composite: false},
		{name: "10000_logs_1.0_false", totalLogs: 10000, rubbishRatio: 1.0, composite: false},
		{name: "100000_logs_1.0_false", totalLogs: 100000, rubbishRatio: 1.0, composite: false},
		{name: "10000000_logs_1.0_false", totalLogs: 10000000, rubbishRatio: 1.0, composite: false},
		// Composite logs, rubbish ratio 0.0
		{name: "1000_logs_0.0_true", totalLogs: 1000, rubbishRatio: 0.0, composite: true},
		{name: "10000_logs_0.0_true", totalLogs: 10000, rubbishRatio: 0.0, composite: true},
		{name: "100000_logs_0.0_true", totalLogs: 100000, rubbishRatio: 0.0, composite: true},
		{name: "1000000_logs_0.0_true", totalLogs: 1000000, rubbishRatio: 0.0, composite: true},
		// Composite logs, rubbish ratio 0.5
		{name: "1000_logs_0.5_true", totalLogs: 1000, rubbishRatio: 0.5, composite: true},
		{name: "10000_logs_0.5_true", totalLogs: 10000, rubbishRatio: 0.5, composite: true},
		{name: "100000_logs_0.5_true", totalLogs: 100000, rubbishRatio: 0.5, composite: true},
		{name: "1000000_logs_0.5_true", totalLogs: 1000000, rubbishRatio: 0.5, composite: true},
		// Composite logs, rubbish ratio 1.0
		{name: "1000_logs_1.0_true", totalLogs: 1000, rubbishRatio: 1.0, composite: true},
		{name: "10000_logs_1.0_true", totalLogs: 10000, rubbishRatio: 1.0, composite: true},
		{name: "100000_logs_1.0_true", totalLogs: 100000, rubbishRatio: 1.0, composite: true},
		{name: "10000000_logs_1.0_true", totalLogs: 10000000, rubbishRatio: 1.0, composite: true},
	}

	for _, tc := range testCases {
		b.Run("ConsumeLogs", func(b *testing.B) {
			cfg, err := loadConfigFromFileBenchmark(b, filepath.Join("testdata", "benchmark", "config.yaml"))
			require.NoError(b, err)
			factory := NewFactory()
			sink := &consumertest.LogsSink{}

			conn, err := factory.CreateLogsToLogs(context.Background(),
				connectortest.NewNopSettings(metadata.Type), cfg, sink)
			require.NoError(b, err)

			require.NoError(b, conn.Start(context.Background(), componenttest.NewNopHost()))
			defer func() {
				assert.NoError(b, conn.Shutdown(context.Background()))
			}()

			logs := genLogsFromConfig(b, cfg, tc.totalLogs, tc.rubbishRatio, tc.composite)

			b.ReportAllocs()
			b.ResetTimer()
			for _, log := range logs {
				// golden.WriteLogsToFile(fmt.Sprintf("benchmark_%s_%d.log", tc.name, i), log)
				assert.NoError(b, conn.ConsumeLogs(context.Background(), log))
			}
			b.StopTimer()

			getLogsCount := len(sink.AllLogs())
			expectedLogsCount := tc.totalLogs - int(float32(tc.totalLogs)*tc.rubbishRatio)
			if tc.composite { // If composite logs are used, we expect only one log from the connector
				expectedLogsCount = 1
			}

			assert.Equal(b, getLogsCount, expectedLogsCount, "Expected to receive %d logs, but got %d logs",
				expectedLogsCount, getLogsCount)
		})
	}

}
