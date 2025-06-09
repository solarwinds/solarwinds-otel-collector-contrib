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
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/storage"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type solarwindsentity struct {
	logger *zap.Logger

	logsConsumer        consumer.Logs
	entitiesDefinitions map[string]config.Entity
	sourcePrefix        string
	destinationPrefix   string
	events              config.ParsedEvents

	expirationPolicy config.ExpirationPolicy
	cache            *storage.Cache

	component.ShutdownFunc
}

var _ connector.Metrics = (*solarwindsentity)(nil)
var _ connector.Logs = (*solarwindsentity)(nil)

func (s *solarwindsentity) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (s *solarwindsentity) Start(ctx context.Context, host component.Host) error {
	s.logger.Info("Starting solarwindsentity connector")
	if s.expirationPolicy.Enabled {
		expirationCfg := s.expirationPolicy.Parse()
		if expirationCfg == nil {
			s.logger.Error("expiration policy is invalid")
		}

		em := storage.NewEvictionManager(s.logsConsumer, s.logger)
		em.Start(ctx)
		cache := storage.NewCache(expirationCfg, s.logger, em)
		s.cache = cache
		go cache.Run(ctx)
	} else {
		s.logger.Debug("expiration policy is disabled, no expiration logs will be generated")
	}

	return nil
}

func (s *solarwindsentity) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	eventLogs := plog.NewLogs()
	eventBuilder := internal.NewEventBuilder(s.entitiesDefinitions, s.sourcePrefix, s.destinationPrefix, &eventLogs, s.logger, s.cache)

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		resourceMetric := metrics.ResourceMetrics().At(i)
		resourceAttrs := resourceMetric.Resource().Attributes()
		for j := 0; j < resourceMetric.ScopeMetrics().Len(); j++ {
			scopeMetric := resourceMetric.ScopeMetrics().At(j)

			for k := 0; k < scopeMetric.Metrics().Len(); k++ {
				metric := scopeMetric.Metrics().At(k)

				tc := ottlmetric.NewTransformContext(metric, scopeMetric.Metrics(), scopeMetric.Scope(), resourceMetric.Resource(), scopeMetric, resourceMetric)
				err := internal.ProcessEvents(ctx, eventBuilder, s.events.MetricEvents, resourceAttrs, tc)
				if err != nil {
					s.logger.Error("Failed to process metric condition", zap.Error(err))
					return err
				}
			}
		}
	}

	if eventLogs.LogRecordCount() == 0 {
		return nil
	}

	return s.logsConsumer.ConsumeLogs(ctx, eventLogs)
}

func (s *solarwindsentity) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	eventLogs := plog.NewLogs()
	eventBuilder := internal.NewEventBuilder(s.entitiesDefinitions, s.sourcePrefix, s.destinationPrefix, &eventLogs, s.logger, s.cache)

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		resourceLog := logs.ResourceLogs().At(i)
		resourceAttrs := resourceLog.Resource().Attributes()

		for j := 0; j < resourceLog.ScopeLogs().Len(); j++ {
			scopeLog := resourceLog.ScopeLogs().At(j)

			for k := 0; k < scopeLog.LogRecords().Len(); k++ {
				logRecord := scopeLog.LogRecords().At(k)

				tc := ottllog.NewTransformContext(logRecord, scopeLog.Scope(), resourceLog.Resource(), scopeLog, resourceLog)
				err := internal.ProcessEvents(ctx, eventBuilder, s.events.LogEvents, resourceAttrs, tc)
				if err != nil {
					s.logger.Error("Failed to process logs condition", zap.Error(err))
					return err
				}
			}
		}
	}

	if eventLogs.LogRecordCount() == 0 {
		return nil
	}

	return s.logsConsumer.ConsumeLogs(ctx, eventLogs)
}
