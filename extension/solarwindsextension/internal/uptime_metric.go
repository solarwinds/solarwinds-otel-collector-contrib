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
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-releases/extension/solarwindsextension/internal/metadata"
	"github.com/solarwinds/solarwinds-otel-collector-releases/pkg/version"
)

func newUptimeMetric(logger *zap.Logger) *UptimeMetric {
	logger.Debug("Creating UptimeMetric")
	return &UptimeMetric{logger: logger, uptime: newUptimeCounter()}
}

type UptimeMetric struct {
	logger *zap.Logger
	uptime *uptimeCounter
}

func (um *UptimeMetric) add(_ context.Context, md pmetric.Metrics) error {
	um.logger.Debug("Adding uptime metric")
	res := md.ResourceMetrics().AppendEmpty()
	scopeMetrics := res.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName(metadata.ScopeName)
	scopeMetrics.Scope().SetVersion(version.Version)
	m := scopeMetrics.Metrics().AppendEmpty()

	m.SetName("sw.otelcol.uptime")
	dataPoint := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dataPoint.SetDoubleValue(um.uptime.Get())
	return nil
}
