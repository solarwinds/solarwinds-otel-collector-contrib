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
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension"
)

func newExporter(ctx context.Context, set extension.Settings, cfg *Config) (exporter.Metrics, error) {
	set.Logger.Debug("Creating OTLP Exporter")

	factory := otlpexporter.NewFactory()
	eCfg, err := cfg.ApplyGrpcConfig(factory.CreateDefaultConfig().(*otlpexporter.Config))
	if err != nil {
		return nil, err
	}
	expSet := toExporterSettings(set)
	return factory.CreateMetrics(ctx, expSet, eCfg)
}

func toExporterSettings(set extension.Settings) exporter.Settings {
	return exporter.Settings{
		ID:                component.NewIDWithName(component.MustNewType("otlp"), fmt.Sprintf("%s-%s", set.ID.Type(), set.ID.Name())),
		TelemetrySettings: set.TelemetrySettings,
		BuildInfo:         set.BuildInfo,
	}
}
