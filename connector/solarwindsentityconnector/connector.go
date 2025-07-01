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
	"errors"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/config"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal/storage"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type solarwindsentity struct {
	logger *zap.Logger

	logsConsumer  consumer.Logs
	eventDetector *internal.EventDetector

	expirationPolicy config.ExpirationPolicy
	storageManager   *storage.Manager
}

var _ connector.Metrics = (*solarwindsentity)(nil)
var _ connector.Logs = (*solarwindsentity)(nil)

func (s *solarwindsentity) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (s *solarwindsentity) Start(context.Context, component.Host) error {
	if s.expirationPolicy.Enabled {
		expirationCfg, err := s.expirationPolicy.Parse()
		if err != nil {
			return errors.Join(err, fmt.Errorf("expiration policy is invalid"))
		}

		ec := internal.NewEventConsumer(s.logsConsumer)
		sm, err := storage.NewStorageManager(expirationCfg, s.logger, ec)
		if err != nil {
			s.logger.Error("failed to create storage manager", zap.Error(err))
			return err
		}

		s.storageManager = sm
		err = s.storageManager.Start()

		if err != nil {
			s.logger.Error("failed to start storage manager", zap.Error(err))
			return err
		}
		s.logger.Info("expiration policy is enabled and started, expiration logs will be generated")
	} else {
		s.logger.Info("expiration policy is disabled, no expiration logs will be generated")
	}

	return nil
}

func (s *solarwindsentity) Shutdown(context.Context) error {
	if s.storageManager != nil {
		err := s.storageManager.Shutdown()
		if err != nil {
			s.logger.Error("failed to shutdown storage manager", zap.Error(err))
			return err
		}
		s.logger.Info("storage manager shutdown successfully")
	}

	s.logger.Info("solarwindsentity connector shutdown successfully")
	return nil
}

func (s *solarwindsentity) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	eventLogs := plog.NewLogs()
	logRecords := internal.CreateEventLog(&eventLogs)

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		resourceMetric := metrics.ResourceMetrics().At(i)
		resourceAttrs := resourceMetric.Resource().Attributes()
		for j := 0; j < resourceMetric.ScopeMetrics().Len(); j++ {
			scopeMetric := resourceMetric.ScopeMetrics().At(j)

			for k := 0; k < scopeMetric.Metrics().Len(); k++ {
				metric := scopeMetric.Metrics().At(k)

				tc := ottlmetric.NewTransformContext(metric, scopeMetric.Metrics(), scopeMetric.Scope(), resourceMetric.Resource(), scopeMetric, resourceMetric)
				events, err := s.eventDetector.DetectMetric(ctx, resourceAttrs, tc)

				if err != nil {
					s.logger.Error("Failed to process metric condition", zap.Error(err))
					return err
				}
				for _, event := range events {
					err := s.handleEvent(event, logRecords)
					if err != nil {
						s.logger.Error("failed to handle event", zap.Error(err))
						return err
					}
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
	logRecords := internal.CreateEventLog(&eventLogs)

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		resourceLog := logs.ResourceLogs().At(i)
		resourceAttrs := resourceLog.Resource().Attributes()

		for j := 0; j < resourceLog.ScopeLogs().Len(); j++ {
			scopeLog := resourceLog.ScopeLogs().At(j)

			for k := 0; k < scopeLog.LogRecords().Len(); k++ {
				logRecord := scopeLog.LogRecords().At(k)

				tc := ottllog.NewTransformContext(logRecord, scopeLog.Scope(), resourceLog.Resource(), scopeLog, resourceLog)
				events, err := s.eventDetector.DetectLog(ctx, resourceAttrs, tc)

				if err != nil {
					s.logger.Error("Failed to process logs condition", zap.Error(err))
					return err
				}

				for _, event := range events {
					err := s.handleEvent(event, logRecords)
					if err != nil {
						s.logger.Error("failed to handle event", zap.Error(err))
						return err
					}
				}
			}
		}
	}

	if eventLogs.LogRecordCount() == 0 {
		return nil
	}

	return s.logsConsumer.ConsumeLogs(ctx, eventLogs)
}

func (s *solarwindsentity) handleEvent(event internal.Event, eventLogs *plog.LogRecordSlice) error {
	action, err := internal.GetActionType(event)
	if err != nil {
		return err
	}
	switch action {
	case internal.EventUpdateAction:
		event.Update(eventLogs)
		if s.storageManager != nil {
			err := s.storageManager.Update(event)
			if err != nil {
				s.logger.Error("Failed to update storage with event", zap.Error(err))
				return err
			}
		}
	case internal.EventDeleteAction:
		event.Delete(eventLogs)
		if s.storageManager != nil {
			err := s.storageManager.Delete(event)
			if err != nil {
				s.logger.Error("Failed to delete event from storage", zap.Error(err))
				return err
			}
		}
	}
	return nil
}
