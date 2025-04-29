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

package swok8sworkloadtypeprocessor // import "github.com/solarwinds/solarwinds-otel-collector-releases/processor/swok8sworkloadtypeprocessor"

import (
	"context"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-releases/internal/k8sconfig"
	"github.com/solarwinds/solarwinds-otel-collector-releases/processor/swok8sworkloadtypeprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/processor/xprocessor"
)

func NewFactory() processor.Factory {
	return xprocessor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		xprocessor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		WatchSyncPeriod:     time.Minute * 5,
		mappedExpectedTypes: make(map[string]groupVersionResourceKind),
	}
}

func createMetricsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextMetricsConsumer consumer.Metrics,
) (processor.Metrics, error) {
	p := &swok8sworkloadtypeProcessor{
		logger:   params.Logger,
		config:   cfg.(*Config),
		settings: params,
	}

	return processorhelper.NewMetrics(
		ctx,
		params,
		cfg,
		nextMetricsConsumer,
		p.processMetrics,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
		processorhelper.WithStart(p.Start),
		processorhelper.WithShutdown(p.Shutdown))
}
