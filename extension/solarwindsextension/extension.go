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

package solarwindsextension

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension/internal"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/componentcommunication"
)

// CollectorNameAttribute is a resource attribute
// representing the configured name of the collector.
const CollectorNameAttribute = internal.CollectorNameAttribute
const EntityCreation = internal.EntityCreation
const EntityCreationValue = internal.EntityCreationValue

type SolarwindsExtension struct {
	logger    *zap.Logger
	config    *internal.Config
	heartbeat *internal.Heartbeat

	componentCommunicationClient componentcommunication.Client
}

func newExtension(
	ctx context.Context,
	set extension.Settings,
	cfg *internal.Config,
	client componentcommunication.Client,
) (*SolarwindsExtension, error) {
	set.Logger.Info("Creating Solarwinds Extension - Custom Build 4")
	set.Logger.Info("Config", zap.Any("config", cfg))

	e := &SolarwindsExtension{
		logger:                       set.Logger,
		config:                       cfg,
		componentCommunicationClient: client,
	}
	var err error
	e.heartbeat, err = internal.NewHeartbeat(ctx, set, cfg)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *SolarwindsExtension) GetCommonConfig() CommonConfig { return newCommonConfig(e.config) }

func (e *SolarwindsExtension) Start(ctx context.Context, host component.Host) error {
	e.logger.Info("Starting Solarwinds Extension")

	controllingCtx, controllingCancel := context.WithCancel(ctx)
	ccMessages, ccErrors := e.componentCommunicationClient.Start(controllingCtx, host)
	go e.componentCommunicationLoop(ccMessages, ccErrors, controllingCtx, controllingCancel)

	return e.heartbeat.Start(controllingCtx, host)
}

func (e *SolarwindsExtension) componentCommunicationLoop(
	ccMessages <-chan *componentcommunication.Message,
	ccErrors <-chan error,
	controllingCtx context.Context,
	controllingCancel context.CancelFunc,
) {
	e.logger.Info("Starting component communication loop in extension")

	for {
		select {
		case msg, ok := <-ccMessages:
			if !ok {
				e.logger.Info("Component communication channel closed")
				ccMessages = nil
				continue
			}
			e.logger.Info(
				"Received message from component communication",
				zap.String("extension", msg.ExtensionID),
				zap.String("capability", msg.CustomMessage.Capability),
				zap.String("type", msg.CustomMessage.Type),
				zap.String("message", string(msg.CustomMessage.Data)),
			)
			continue
		case err, ok := <-ccErrors:
			if !ok {
				e.logger.Info("Component communication error channel closed")
				ccErrors = nil
				continue
			}
			e.logger.Error("Error received from component communication", zap.Error(err))
			controllingCancel()
		case <-controllingCtx.Done():
			e.logger.Info("Controlling context done")
			return
		}
	}
}

func (e *SolarwindsExtension) Shutdown(ctx context.Context) error {
	e.logger.Info("Shutting down Solarwinds Extension")
	// Everything must be shut down, regardless of the failure.
	e.componentCommunicationClient.Stop()
	return e.heartbeat.Shutdown(ctx)

}
