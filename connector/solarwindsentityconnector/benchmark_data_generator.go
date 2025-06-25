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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"gopkg.in/yaml.v3"
)

// Helper function to generate metrics based on the provided configuration.
// It creates metrics for each entity and relationship defined in the configuration,
// and adds a specified amount of "rubbish" metrics to the final output.
// The `count` parameter specifies the total number of metrics to generate, including both valid and rubbish metrics.
// The `rubbishRatio` parameter defines the proportion of invalid metrics.
// The `composite` parameter determines whether to return a single composite metric instead of multiple individual metrics.
func genMetricsFromConfig(b *testing.B, cfg *Config, count int, rubbishRatio float32, composite bool) []pmetric.Metrics {
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

	rubbishCount := int(float32(count) * rubbishRatio)
	normalCount := count - rubbishCount

	finalMetrics := make([]pmetric.Metrics, 0, count)
	if composite {
		finalMetrics = append(finalMetrics, buildCompositeMetric(b, count, uniqueMetrics))
	} else {
		for i := 0; i < normalCount; i++ {
			finalMetrics = append(finalMetrics, uniqueMetrics[i%len(uniqueMetrics)])
		}
	}

	for i := 0; i < rubbishCount; i++ {
		finalMetrics = append(finalMetrics, buildRubbishMetric(b, i))
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

func buildEntityMetric(b *testing.B, cfg *Config, entityType string) pmetric.Metrics {
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

func buildRelationshipMetric(b *testing.B, cfg *Config, relType, sourceType, destType string) pmetric.Metrics {
	b.Helper()

	metric := buildEmptyMetric(b)
	attrs := metric.ResourceMetrics().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		switch ent.Type {
		case sourceType:
			for _, id := range ent.IDs {
				attrs.PutStr(id, fmt.Sprintf("%s.%s.%s", id, ent.Type, relType))
			}
		case destType:
			for _, id := range ent.IDs {
				attrs.PutStr(id, fmt.Sprintf("%s.%s.%s", id, ent.Type, relType))
			}
		}
	}

	return metric
}

func buildRubbishMetric(b *testing.B, i int) pmetric.Metrics {
	b.Helper()

	metric := buildEmptyMetric(b)
	attrs := metric.ResourceMetrics().At(0).Resource().Attributes()
	attrs.PutStr("NotValid", fmt.Sprintf("%d", i))

	return metric
}

func buildCompositeMetric(b *testing.B, count int, metrics []pmetric.Metrics) pmetric.Metrics {
	b.Helper()

	composite := pmetric.NewMetrics()
	for i := 0; i < count; i++ {
		metric := metrics[i%len(metrics)]
		for j := 0; j < metric.ResourceMetrics().Len(); j++ {
			src := metric.ResourceMetrics().At(j)
			dst := composite.ResourceMetrics().AppendEmpty()
			src.MoveTo(dst)
		}
	}

	return composite
}

// Helper function to generate logs based on the provided configuration.
// It creates logs for each entity and relationship defined in the configuration,
// and adds a specified amount of "rubbish" logs to the final output.
// The `count` parameter specifies the total number of logs to generate, including both valid and rubbish logs.
// The `rubbishRatio` parameter defines the proportion of invalid logs.
// The `composite` parameter determines whether to return a single composite log instead of multiple individual logs.
func genLogsFromConfig(b *testing.B, cfg *Config, count int, rubbishRatio float32, composite bool) []plog.Logs {
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

	rubbishCount := int(float32(count) * rubbishRatio)
	normalCount := count - rubbishCount

	finalLogs := make([]plog.Logs, 0, count)

	// If composite logs are requested, we will build a single composite log
	// that contains all unique logs repeated as provided by the count.
	if composite {
		finalLogs = append(finalLogs, buildCompositeLog(b, count, uniqueLogs))
	} else {
		for i := 0; i < normalCount; i++ {
			finalLogs = append(finalLogs, uniqueLogs[i%len(uniqueLogs)])
		}
	}

	for i := 0; i < rubbishCount; i++ {
		finalLogs = append(finalLogs, buildRubbishLog(b, i))
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

func buildEntityLog(b *testing.B, cfg *Config, entityType string) plog.Logs {
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

func buildRelationshipLog(b *testing.B, cfg *Config, relType, sourceType, destType string) plog.Logs {
	b.Helper()

	log := buildEmptyLog(b)
	attrs := log.ResourceLogs().At(0).Resource().Attributes()

	for _, ent := range cfg.Schema.Entities {
		switch ent.Type {
		case sourceType:
			for _, id := range ent.IDs {
				attrs.PutStr(id, fmt.Sprintf("%s.%s.%s", id, ent.Type, relType))
			}
		case destType:
			for _, id := range ent.IDs {
				attrs.PutStr(id, fmt.Sprintf("%s.%s.%s", id, ent.Type, relType))
			}
		}
	}
	return log
}

func buildRubbishLog(b *testing.B, i int) plog.Logs {
	b.Helper()

	log := buildEmptyLog(b)
	attrs := log.ResourceLogs().At(0).Resource().Attributes()
	attrs.PutStr("NotValid", fmt.Sprintf("%d", i))
	return log
}

func buildCompositeLog(b *testing.B, amount int, logs []plog.Logs) plog.Logs {
	b.Helper()

	compositeLog := plog.NewLogs()

	for i := 0; i < amount; i++ {
		log := logs[i%len(logs)]
		for j := 0; j < log.ResourceLogs().Len(); j++ {
			srcRL := log.ResourceLogs().At(j)
			dstRL := compositeLog.ResourceLogs().AppendEmpty()
			srcRL.MoveTo(dstRL)
		}
	}

	return compositeLog
}

func loadConfigFromFileBenchmark(b *testing.B, path string) (*Config, error) {
	b.Helper()

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
