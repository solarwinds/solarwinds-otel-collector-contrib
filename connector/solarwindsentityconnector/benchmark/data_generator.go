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
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// Helper function to generate metrics based on the provided configuration.
// It creates metrics for each entity and relationship defined in the configuration, and repeats them as specified by the count.
// Also adds a specified amount of "invalid" metrics to the final output.
// The count parameter specifies the total number of metrics to generate, including both valid and invalid metrics.
// The invalidRatio parameter defines the proportion of invalid metrics.
// The multiple parameter determines whether to return a multiple metric instead of single metrics.
func generateTestMetrics(b *testing.B, cfg *solarwindsentityconnector.Config, count int, invalidRatio float32, multiple bool) []pmetric.Metrics {
	b.Helper()

	var finalMetrics []pmetric.Metrics
	invalidCount := int(float32(count) * invalidRatio)
	validCount := count - invalidCount

	if validCount > 0 {
		finalMetrics = append(finalMetrics, buildValidMetrics(b, validCount, cfg, multiple)...)
	}

	if invalidCount > 0 {
		finalMetrics = append(finalMetrics, buildInvalidMetric(b, invalidCount)...)
	}

	return finalMetrics
}

// Build valid metrics based on the provided count and configuration.
func buildValidMetrics(b *testing.B, count int, cfg *solarwindsentityconnector.Config, multiple bool) []pmetric.Metrics {
	b.Helper()

	var outputMetrics []pmetric.Metrics
	configuredMetrics := buildConfiguredMetrics(b, cfg)
	if multiple {
		outputMetrics = append(outputMetrics, buildMultipleMetric(b, count, configuredMetrics))
	} else {
		outputMetrics = append(outputMetrics, buildSingleMetrics(b, count, configuredMetrics)...)
	}

	return outputMetrics
}

// Build an invalid metrics with a specific attribute to indicate its invalidity
func buildInvalidMetric(b *testing.B, count int) []pmetric.Metrics {
	b.Helper()

	var invalidMetric []pmetric.Metrics
	for range count {
		metric := buildEmptyMetric(b)
		attrs := metric.ResourceMetrics().At(0).Resource().Attributes()
		attrs.PutStr("invalid", "test-value")
		invalidMetric = append(invalidMetric, metric)
	}

	return invalidMetric
}

// Build metrics for each entity and relationship from the configuration.
func buildConfiguredMetrics(b *testing.B, cfg *solarwindsentityconnector.Config) []pmetric.Metrics {
	b.Helper()

	var configuredMetrics []pmetric.Metrics
	for _, e := range cfg.Schema.Events.Entities {
		if e.Context != ottlmetric.ContextName {
			continue
		}
		metric := buildEntityMetric(b, cfg, e.Type)
		configuredMetrics = append(configuredMetrics, metric)
	}

	for _, r := range cfg.Schema.Events.Relationships {
		if r.Context != ottlmetric.ContextName {
			continue
		}
		metric := buildRelationshipMetric(b, cfg, r.Type, r.Source, r.Destination)
		configuredMetrics = append(configuredMetrics, metric)
	}

	return configuredMetrics
}

// Build a array of single metrics that contains all unique metrics repeated as provided by the count.
func buildSingleMetrics(b *testing.B, count int, configuredMetrics []pmetric.Metrics) []pmetric.Metrics {
	b.Helper()

	var metric []pmetric.Metrics
	for i := 0; i < count; i++ {
		metric = append(metric, configuredMetrics[i%len(configuredMetrics)]) // Loop through the metrics cyclically
	}

	return metric
}

// Build a multiple metric that contains all unique metrics repeated as provided by the count.
func buildMultipleMetric(b *testing.B, count int, metrics []pmetric.Metrics) pmetric.Metrics {
	b.Helper()

	multiple := pmetric.NewMetrics()
	for i := 0; i < count; i++ {
		metric := metrics[i%len(metrics)]       // Loop through the metrics cyclically
		srcRM := metric.ResourceMetrics().At(0) // There is only one resource metric in each metric
		dstRM := multiple.ResourceMetrics().AppendEmpty()
		srcRM.CopyTo(dstRM)
	}

	return multiple
}

// Build an empty metric with a test scope metric.
func buildEmptyMetric(b *testing.B) pmetric.Metrics {
	b.Helper()

	metric := pmetric.NewMetrics()
	rm := metric.ResourceMetrics().AppendEmpty()

	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("benchmark")

	m := sm.Metrics().AppendEmpty()
	m.SetName("dummy_metric")
	m.SetEmptyGauge()

	return metric
}

// Build metric for a specific entity.
func buildEntityMetric(b *testing.B, cfg *solarwindsentityconnector.Config, entityType string) pmetric.Metrics {
	b.Helper()

	metric := buildEmptyMetric(b)
	attrs := metric.ResourceMetrics().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		if ent.Type != entityType {
			continue
		}
		for _, id := range ent.IDs {
			attrs.PutStr(id, "test-value")
		}
		break
	}

	return metric
}

// Build metric for a specific relationship type between source and destination entities.
func buildRelationshipMetric(b *testing.B, cfg *solarwindsentityconnector.Config, relType, sourceType, destType string) pmetric.Metrics {
	b.Helper()

	metric := buildEmptyMetric(b)
	attrs := metric.ResourceMetrics().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		if ent.Type == sourceType {
			for _, id := range ent.IDs {
				attrs.PutStr("src."+id, "test-value")
			}
		}

		if ent.Type == destType {
			for _, id := range ent.IDs {
				attrs.PutStr("dst."+id, "test-value")
			}
		}
	}

	return metric
}

// Helper function to generate logs based on the provided configuration.
// It creates logs for each entity and relationship defined in the configuration, and repeats them as specified by the count.
// Also adds a specified amount of "invalid" logs to the final output.
// The count parameter specifies the total number of logs to generate, including both valid and invalid logs.
// The invalidRatio parameter defines the proportion of invalid logs.
// The multiple parameter determines whether to return a multiple log instead of single logs.
func generateTestLogs(b *testing.B, cfg *solarwindsentityconnector.Config, count int, invalidRatio float32, multiple bool) []plog.Logs {
	b.Helper()

	var finalLogs []plog.Logs
	invalidCount := int(float32(count) * invalidRatio)
	validCount := count - invalidCount

	if validCount > 0 {
		finalLogs = append(finalLogs, buildValidLogs(b, validCount, cfg, multiple)...)
	}

	if invalidCount > 0 {
		finalLogs = append(finalLogs, buildInvalidLog(b, invalidCount)...)
	}

	return finalLogs
}

// Build valid logs based on the provided count and configuration.
func buildValidLogs(b *testing.B, count int, cfg *solarwindsentityconnector.Config, multiple bool) []plog.Logs {
	b.Helper()

	var outputLogs []plog.Logs
	configuredLogs := buildConfiguredLogs(b, cfg)
	if multiple {
		outputLogs = append(outputLogs, buildMultipleLog(b, count, configuredLogs))
	} else {
		outputLogs = append(outputLogs, buildSingleLogs(b, count, configuredLogs)...)
	}

	return outputLogs
}

// Build an invalid logs with a specific attribute to indicate its invalidity
func buildInvalidLog(b *testing.B, count int) []plog.Logs {
	b.Helper()
	var invalidLogs []plog.Logs
	for range count {
		log := buildEmptyLog(b)
		attrs := log.ResourceLogs().At(0).Resource().Attributes()
		attrs.PutStr("invalid", "test-value")
		invalidLogs = append(invalidLogs, log)
	}
	return invalidLogs
}

// Build logs for each entity and relationship from the configuration.
func buildConfiguredLogs(b *testing.B, cfg *solarwindsentityconnector.Config) []plog.Logs {
	b.Helper()

	var configuredLogs []plog.Logs
	for _, e := range cfg.Schema.Events.Entities {
		if e.Context != ottllog.ContextName {
			continue
		}
		log := buildEntityLog(b, cfg, e.Type)
		configuredLogs = append(configuredLogs, log)
	}

	for _, r := range cfg.Schema.Events.Relationships {
		if r.Context != ottllog.ContextName {
			continue
		}
		log := buildRelationshipLog(b, cfg, r.Type, r.Source, r.Destination)
		configuredLogs = append(configuredLogs, log)
	}

	return configuredLogs
}

// Build a array of single logs that contains all unique logs repeated as provided by the count.
func buildSingleLogs(b *testing.B, count int, configuredLogs []plog.Logs) []plog.Logs {
	b.Helper()

	var logs []plog.Logs
	for i := 0; i < count; i++ {
		logs = append(logs, configuredLogs[i%len(configuredLogs)]) // Loop through the logs cyclically
	}
	return logs
}

// Build a multiple log that contains all unique logs repeated as provided by the count.
func buildMultipleLog(b *testing.B, amount int, logs []plog.Logs) plog.Logs {
	b.Helper()

	multipleLog := plog.NewLogs()

	for i := 0; i < amount; i++ {
		log := logs[i%len(logs)]          // Loop through the logs cyclically
		srcRL := log.ResourceLogs().At(0) // There is only one resource log in each log
		dstRL := multipleLog.ResourceLogs().AppendEmpty()
		srcRL.CopyTo(dstRL)
	}

	return multipleLog
}

// Build an empty log with a test scope log.
func buildEmptyLog(b *testing.B) plog.Logs {
	b.Helper()

	log := plog.NewLogs()
	rl := log.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return log
}

// Build log for a specific entity.
func buildEntityLog(b *testing.B, cfg *solarwindsentityconnector.Config, entityType string) plog.Logs {
	b.Helper()

	log := buildEmptyLog(b)
	attrs := log.ResourceLogs().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		if ent.Type != entityType {
			continue
		}
		for _, id := range ent.IDs {
			attrs.PutStr(id, "test-value")
		}
		break
	}
	return log
}

// Build log for a specific relationship type between source and destination entities.
func buildRelationshipLog(b *testing.B, cfg *solarwindsentityconnector.Config, relType, sourceType, destType string) plog.Logs {
	b.Helper()

	log := buildEmptyLog(b)
	attrs := log.ResourceLogs().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		if ent.Type == sourceType {
			for _, id := range ent.IDs {
				attrs.PutStr("src."+id, "test-value")
			}
		}

		if ent.Type == destType {
			for _, id := range ent.IDs {
				attrs.PutStr("dst."+id, "test-value")
			}
		}
	}
	return log
}
