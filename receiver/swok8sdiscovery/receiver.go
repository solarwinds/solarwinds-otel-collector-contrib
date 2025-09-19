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

package swok8sdiscovery

import (
	"context"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
)

type swok8sdiscoveryReceiver struct {
	setting  receiver.Settings
	config   *Config
	consumer consumer.Logs
	client   dynamic.Interface // retained if needed elsewhere
	kclient  k8s.Interface
	cancel   context.CancelFunc

	// Optional callback invoked after every discovery cycle (test instrumentation / extensibility)
	cycleCallback func()
}

func newReceiver(params receiver.Settings, config *Config, consumer consumer.Logs) (receiver.Logs, error) {

	return &swok8sdiscoveryReceiver{
		setting:  params,
		config:   config,
		consumer: consumer,
	}, nil
}

func (r *swok8sdiscoveryReceiver) Start(ctx context.Context, host component.Host) error {
	// typed client
	kclient, err := r.config.getClient()
	if err != nil {
		return err
	}
	r.kclient = kclient

	r.setting.Logger.Info("Starting swok8sdiscovery receiver")

	loopCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	go r.discoveryLoop(loopCtx)
	return nil
}

func (r *swok8sdiscoveryReceiver) discoveryLoop(ctx context.Context) {
	ticker := newTicker(ctx, r.config.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.performDiscoveryCycle(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// performDiscoveryCycle performs one full discovery iteration.
func (r *swok8sdiscoveryReceiver) performDiscoveryCycle(ctx context.Context) {
	pods, err := r.config.listPods(ctx, r.kclient)
	if err != nil {
		r.setting.Logger.Error("Failed to list pods", zap.Error(err))
		return
	}
	services, err := r.config.listServices(ctx, r.kclient)
	if err != nil {
		r.setting.Logger.Error("Failed to list services", zap.Error(err))
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		r.discoverDatabasesByImages(ctx, pods, services)
	}()

	go func() {
		defer wg.Done()
		r.discoverDatabasesByDomains(ctx, pods, services)
	}()

	wg.Wait()
	if r.cycleCallback != nil {
		r.cycleCallback()
	}
}

func (r *swok8sdiscoveryReceiver) Shutdown(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
