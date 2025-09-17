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

package mqttreceiver

import (
	"context"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type mqttObjectsReceiver struct {
	settings          receiver.Settings
	ctx               context.Context
	consumer          consumer.Metrics
	config            *Config
	subscribedBrokers []*brokerSubscription
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

func newReceiver(params receiver.Settings, config *Config, ctx context.Context, consumer consumer.Metrics) (receiver.Metrics, error) {
	params.Logger.Info("Creating new MQTT receiver")

	ctx, cancel := context.WithCancel(ctx)

	return &mqttObjectsReceiver{
		settings: params,
		ctx:      ctx,
		consumer: consumer,
		config:   config,
		cancel:   cancel,
	}, nil
}

func (kr *mqttObjectsReceiver) Start(_ context.Context, _ component.Host) error {
	kr.settings.Logger.Info("Starting MQTT receiver")

	kr.subscribedBrokers = make([]*brokerSubscription, 0, len(kr.config.Brokers))

	for _, broker := range kr.config.Brokers {
		subBroker := newBrokerSubscription(broker, kr.settings, kr.settings.Logger, kr.consumer, kr.ctx)
		if err := subBroker.Start(); err != nil {
			kr.settings.Logger.Error("Failed to start broker subscription",
				zap.String("broker", broker.Name),
				zap.Error(err))
			continue
		}
		kr.subscribedBrokers = append(kr.subscribedBrokers, subBroker)
	}

	return nil
}

func (kr *mqttObjectsReceiver) Shutdown(ctx context.Context) error {
	kr.settings.Logger.Info("Shutting down MQTT receiver")

	if kr.cancel != nil {
		kr.cancel()
	}

	// Stop all subscribed brokers
	for _, subBroker := range kr.subscribedBrokers {
		if subBroker != nil {
			subBroker.Stop()
		}
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		kr.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		kr.settings.Logger.Info("All MQTT connections closed gracefully")
	case <-ctx.Done():
		kr.settings.Logger.Warn("Shutdown timeout reached, forcing close")
	}

	return nil
}
