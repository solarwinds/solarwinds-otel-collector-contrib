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

package benchmark

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector"
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
		invalidRatio float32
		multiple     bool
	}{
		// No multiple metric, invalid ratio 0.0
		{name: "100_metrics_0.0_false", totalMetrics: 100, invalidRatio: 0.0, multiple: false},
		{name: "10K_metrics_0.0_false", totalMetrics: 10000, invalidRatio: 0.0, multiple: false},
		{name: "1M_metrics_0.0_false", totalMetrics: 1000000, invalidRatio: 0.0, multiple: false},
		// No multiple metric, invalid ratio 0.5
		{name: "100_metrics_0.5_false", totalMetrics: 100, invalidRatio: 0.5, multiple: false},
		{name: "10K_metrics_0.5_false", totalMetrics: 10000, invalidRatio: 0.5, multiple: false},
		{name: "1M_metrics_0.5_false", totalMetrics: 1000000, invalidRatio: 0.5, multiple: false},
		// No multiple metric, invalid ratio 1.0
		{name: "100_metrics_1.0_false", totalMetrics: 100, invalidRatio: 1.0, multiple: false},
		{name: "10K_metrics_1.0_false", totalMetrics: 10000, invalidRatio: 1.0, multiple: false},
		{name: "1M_metrics_1.0_false", totalMetrics: 1000000, invalidRatio: 1.0, multiple: false},
		// Multiple metric, invalid ratio 0.0
		{name: "100_metrics_0.0_true", totalMetrics: 100, invalidRatio: 0.0, multiple: true},
		{name: "10K_metrics_0.0_true", totalMetrics: 10000, invalidRatio: 0.0, multiple: true},
		{name: "1M_metrics_0.0_true", totalMetrics: 1000000, invalidRatio: 0.0, multiple: true},
		// Multiple metric, invalid ratio 0.5
		{name: "100_metrics_0.5_true", totalMetrics: 100, invalidRatio: 0.5, multiple: true},
		{name: "10K_metrics_0.5_true", totalMetrics: 10000, invalidRatio: 0.5, multiple: true},
		{name: "1M_metrics_0.5_true", totalMetrics: 1000000, invalidRatio: 0.5, multiple: true},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {

			cfg, err := solarwindsentityconnector.LoadConfigFromFile(b, filepath.Join("..", "testdata", "benchmark", "config.yaml"))
			require.NoError(b, err)
			factory := solarwindsentityconnector.NewFactory()
			sink := &consumertest.LogsSink{}

			conn, err := factory.CreateMetricsToLogs(context.Background(),
				connectortest.NewNopSettings(metadata.Type), cfg, sink)
			require.NoError(b, err)

			require.NoError(b, conn.Start(context.Background(), componenttest.NewNopHost()))
			defer func() {
				assert.NoError(b, conn.Shutdown(context.Background()))
			}()

			metrics := generateTestMetrics(b, cfg, tc.totalMetrics, tc.invalidRatio, tc.multiple)

			b.ReportAllocs()
			for b.Loop() {

				for _, metric := range metrics {
					assert.NoError(b, conn.ConsumeMetrics(context.Background(), metric))
				}

				b.StopTimer()

				getLogsCount := len(sink.AllLogs())
				expectedLogsCount := tc.totalMetrics - int(float32(tc.totalMetrics)*tc.invalidRatio)
				if tc.multiple { // If multiple metric is used, we expect only one log from the connector or zero if invalidRatio is 1.0
					expectedLogsCount = int(1 - int(tc.invalidRatio))
				}

				assert.Equal(b, expectedLogsCount, getLogsCount)
				sink.Reset()

				b.StartTimer()
			}
		})
	}
}

func BenchmarkLogs(b *testing.B) {
	testCases := []struct {
		name         string
		totalLogs    int
		invalidRatio float32
		multiple     bool
	}{
		// No multiple log, invalid ratio 0.0
		{name: "100_logs_0.0_false", totalLogs: 100, invalidRatio: 0.0, multiple: false},
		{name: "10K_logs_0.0_false", totalLogs: 10000, invalidRatio: 0.0, multiple: false},
		{name: "1M_logs_0.0_false", totalLogs: 1000000, invalidRatio: 0.0, multiple: false},
		// No multiple log, invalid ratio 0.5
		{name: "100_logs_0.5_false", totalLogs: 100, invalidRatio: 0.5, multiple: false},
		{name: "10K_logs_0.5_false", totalLogs: 10000, invalidRatio: 0.5, multiple: false},
		{name: "1M_logs_0.5_false", totalLogs: 1000000, invalidRatio: 0.5, multiple: false},
		// No multiple log, invalid ratio 1.0
		{name: "100_logs_1.0_false", totalLogs: 100, invalidRatio: 1.0, multiple: false},
		{name: "10K_logs_1.0_false", totalLogs: 10000, invalidRatio: 1.0, multiple: false},
		{name: "1M_logs_1.0_false", totalLogs: 1000000, invalidRatio: 1.0, multiple: false},
		// Multiple log, invalid ratio 0.0
		{name: "100_logs_0.0_true", totalLogs: 100, invalidRatio: 0.0, multiple: true},
		{name: "10K_logs_0.0_true", totalLogs: 10000, invalidRatio: 0.0, multiple: true},
		{name: "1M_logs_0.0_true", totalLogs: 1000000, invalidRatio: 0.0, multiple: true},
		// Multiple log, invalid ratio 0.5
		{name: "100_logs_0.5_true", totalLogs: 100, invalidRatio: 0.5, multiple: true},
		{name: "10K_logs_0.5_true", totalLogs: 10000, invalidRatio: 0.5, multiple: true},
		{name: "1M_logs_0.5_true", totalLogs: 1000000, invalidRatio: 0.5, multiple: true},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			cfg, err := solarwindsentityconnector.LoadConfigFromFile(b, filepath.Join("..", "testdata", "benchmark", "config.yaml"))
			require.NoError(b, err)
			factory := solarwindsentityconnector.NewFactory()
			sink := &consumertest.LogsSink{}

			conn, err := factory.CreateLogsToLogs(context.Background(),
				connectortest.NewNopSettings(metadata.Type), cfg, sink)
			require.NoError(b, err)

			require.NoError(b, conn.Start(context.Background(), componenttest.NewNopHost()))
			defer func() {
				assert.NoError(b, conn.Shutdown(context.Background()))
			}()

			logs := generateTestLogs(b, cfg, tc.totalLogs, tc.invalidRatio, tc.multiple)

			b.ReportAllocs()
			for b.Loop() {

				for _, log := range logs {
					assert.NoError(b, conn.ConsumeLogs(context.Background(), log))
				}

				b.StopTimer()

				getLogsCount := len(sink.AllLogs())
				expectedLogsCount := tc.totalLogs - int(float32(tc.totalLogs)*tc.invalidRatio)
				if tc.multiple { // If multiple log is used, we expect only one log from the connector or zero if invalidRatio is 1.0
					expectedLogsCount = int(1 - int(tc.invalidRatio))
				}

				assert.Equal(b, expectedLogsCount, getLogsCount)
				sink.Reset()

				b.StartTimer()
			}
		})

	}

}
