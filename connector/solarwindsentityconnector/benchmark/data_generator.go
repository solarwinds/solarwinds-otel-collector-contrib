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
	"fmt"
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
// It creates metrics for each entity and relationship defined in the configuration,
// and adds a specified amount of "invalid" metrics to the final output.
// The `count` parameter specifies the total number of metrics to generate, including both valid and invalid metrics.
// The `invalidRatio` parameter defines the proportion of invalid metrics.
// The `multiple` parameter determines whether to return a multiple metric instead of single metrics.
func genMetricsFromConfig(b *testing.B, cfg *solarwindsentityconnector.Config, count int, invalidRatio float32, multiple bool) []pmetric.Metrics {
	b.Helper()

	var uniqueMetrics []pmetric.Metrics
	for _, e := range cfg.Schema.Events.Entities {
		if e.Context != ottlmetric.ContextName {
			continue
		}
		metric := buildEntityMetric(b, cfg, e.Type)
		uniqueMetrics = append(uniqueMetrics, metric)
	}

	for _, r := range cfg.Schema.Events.Relationships {
		if r.Context != ottlmetric.ContextName {
			continue
		}
		metric := buildRelationshipMetric(b, cfg, r.Type, r.Source, r.Destination)
		uniqueMetrics = append(uniqueMetrics, metric)
	}

	invalidCount := int(float32(count) * invalidRatio)
	validCount := count - invalidCount

	finalMetrics := make([]pmetric.Metrics, 0, count)

	// If multiple flag is true, will build a multiple metric
	// that contains all unique metrics repeated as provided by the count.
	if multiple {
		finalMetrics = append(finalMetrics, buildMultipleMetric(b, validCount, uniqueMetrics))
	} else {
		for i := 0; i < validCount; i++ {
			finalMetrics = append(finalMetrics, uniqueMetrics[i%len(uniqueMetrics)])
		}
	}

	for i := 0; i < invalidCount; i++ {
		finalMetrics = append(finalMetrics, buildInvalidMetric(b, i))
	}

	return finalMetrics
}

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

func buildEntityMetric(b *testing.B, cfg *solarwindsentityconnector.Config, entityType string) pmetric.Metrics {
	b.Helper()

	metric := buildEmptyMetric(b)
	attrs := metric.ResourceMetrics().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		if ent.Type != entityType {
			continue
		}
		for _, id := range ent.IDs {
			attrs.PutStr(id, fmt.Sprintf("%s.%s", id, ent.Type))
		}
		break
	}

	return metric
}

func buildRelationshipMetric(b *testing.B, cfg *solarwindsentityconnector.Config, relType, sourceType, destType string) pmetric.Metrics {
	b.Helper()

	metric := buildEmptyMetric(b)
	attrs := metric.ResourceMetrics().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		if ent.Type == sourceType {
			for _, id := range ent.IDs {
				attrs.PutStr(fmt.Sprintf("src.%s", id), fmt.Sprintf("%s.%s.%s", id, ent.Type, relType))
			}
		}

		if ent.Type == destType {
			for _, id := range ent.IDs {
				attrs.PutStr(fmt.Sprintf("dst.%s", id), fmt.Sprintf("%s.%s.%s", id, ent.Type, relType))
			}
		}
	}

	return metric
}

func buildInvalidMetric(b *testing.B, i int) pmetric.Metrics {
	b.Helper()

	metric := buildEmptyMetric(b)
	attrs := metric.ResourceMetrics().At(0).Resource().Attributes()
	attrs.PutStr("InValid", fmt.Sprintf("%d", i))

	return metric
}

func buildMultipleMetric(b *testing.B, count int, metrics []pmetric.Metrics) pmetric.Metrics {
	b.Helper()

	multiple := pmetric.NewMetrics()
	for i := 0; i < count; i++ {
		metric := metrics[i%len(metrics)]
		for j := 0; j < metric.ResourceMetrics().Len(); j++ {
			src := metric.ResourceMetrics().At(j)
			dst := multiple.ResourceMetrics().AppendEmpty()
			src.CopyTo(dst)
		}
	}

	return multiple
}

// Helper function to generate logs based on the provided configuration.
// It creates logs for each entity and relationship defined in the configuration,
// and adds a specified amount of "invalid" logs to the final output.
// The `count` parameter specifies the total number of logs to generate, including both valid and invalid logs.
// The `invalidRatio` parameter defines the proportion of invalid logs.
// The `multiple` parameter determines whether to return a multiple log instead of single logs.
func genLogsFromConfig(b *testing.B, cfg *solarwindsentityconnector.Config, count int, invalidRatio float32, multiple bool) []plog.Logs {
	b.Helper()
	var uniqueLogs []plog.Logs

	for _, e := range cfg.Schema.Events.Entities {
		if e.Context != ottllog.ContextName {
			continue
		}
		log := buildEntityLog(b, cfg, e.Type)
		uniqueLogs = append(uniqueLogs, log)
	}

	for _, r := range cfg.Schema.Events.Relationships {
		if r.Context != ottllog.ContextName {
			continue
		}
		log := buildRelationshipLog(b, cfg, r.Type, r.Source, r.Destination)
		uniqueLogs = append(uniqueLogs, log)
	}

	invalidCount := int(float32(count) * invalidRatio)
	validCount := count - invalidCount

	finalLogs := make([]plog.Logs, 0, count)

	// If multiple flag is true, will build a multiple log
	// that contains all unique logs repeated as provided by the count.
	if multiple {
		finalLogs = append(finalLogs, buildMultipleLog(b, validCount, uniqueLogs))
	} else {
		for i := 0; i < validCount; i++ {
			finalLogs = append(finalLogs, uniqueLogs[i%len(uniqueLogs)])
		}
	}

	for i := 0; i < invalidCount; i++ {
		finalLogs = append(finalLogs, buildInvalidLog(b, i))
	}

	return finalLogs
}

func buildEmptyLog(b *testing.B) plog.Logs {
	b.Helper()

	log := plog.NewLogs()
	rl := log.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return log
}

func buildEntityLog(b *testing.B, cfg *solarwindsentityconnector.Config, entityType string) plog.Logs {
	b.Helper()

	log := buildEmptyLog(b)
	attrs := log.ResourceLogs().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		if ent.Type != entityType {
			continue
		}
		for _, id := range ent.IDs {
			attrs.PutStr(id, fmt.Sprintf("%s.%s", id, ent.Type))
		}
		break
	}
	return log
}

func buildRelationshipLog(b *testing.B, cfg *solarwindsentityconnector.Config, relType, sourceType, destType string) plog.Logs {
	b.Helper()

	log := buildEmptyLog(b)
	attrs := log.ResourceLogs().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		if ent.Type == sourceType {
			for _, id := range ent.IDs {
				attrs.PutStr(fmt.Sprintf("src.%s", id), fmt.Sprintf("%s.%s.%s", id, ent.Type, relType))
			}
		}

		if ent.Type == destType {
			for _, id := range ent.IDs {
				attrs.PutStr(fmt.Sprintf("dst.%s", id), fmt.Sprintf("%s.%s.%s", id, ent.Type, relType))
			}
		}
	}
	return log
}

func buildInvalidLog(b *testing.B, i int) plog.Logs {
	b.Helper()

	log := buildEmptyLog(b)
	attrs := log.ResourceLogs().At(0).Resource().Attributes()
	attrs.PutStr("InValid", fmt.Sprintf("%d", i))
	return log
}

func buildMultipleLog(b *testing.B, amount int, logs []plog.Logs) plog.Logs {
	b.Helper()

	multipleLog := plog.NewLogs()

	for i := 0; i < amount; i++ {
		log := logs[i%len(logs)]
		for j := 0; j < log.ResourceLogs().Len(); j++ {
			srcRL := log.ResourceLogs().At(j)
			dstRL := multipleLog.ResourceLogs().AppendEmpty()
			srcRL.CopyTo(dstRL)
		}
	}

	return multipleLog
}
