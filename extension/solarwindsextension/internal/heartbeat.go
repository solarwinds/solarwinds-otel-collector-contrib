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
	"errors"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

const (
	defaultHeartbeatInterval = 30 * time.Second
	CollectorNameAttribute   = "sw.otelcol.collector.name"
)

type MetricsExporter interface {
	start(context.Context, component.Host) error
	shutdown(context.Context) error
	push(context.Context, pmetric.Metrics) error
}

type Heartbeat struct {
	logger *zap.Logger

	cancel           context.CancelFunc
	startShutdownMtx sync.Mutex

	metric        *UptimeMetric
	exporter      MetricsExporter
	collectorName string

	beatInterval time.Duration
	resource     map[string]string
}

var ErrAlreadyRunning = errors.New("heartbeat already running")

func NewHeartbeat(ctx context.Context, set extension.Settings, cfg *Config) (*Heartbeat, error) {
	set.Logger.Debug("Creating Heartbeat")

	exp, err := newExporter(ctx, set, cfg)
	if err != nil {
		return nil, err
	}

	return newHeartbeatWithExporter(set, cfg, exp), nil
}

func newHeartbeatWithExporter(
	set extension.Settings,
	cfg *Config,
	exporter MetricsExporter,
) *Heartbeat {
	return &Heartbeat{
		logger:        set.Logger,
		metric:        newUptimeMetric(set.Logger),
		collectorName: cfg.CollectorName,
		exporter:      exporter,
		beatInterval:  defaultHeartbeatInterval,
		resource:      cfg.Resource,
	}
}

func (h *Heartbeat) Start(ctx context.Context, host component.Host) error {
	h.startShutdownMtx.Lock()
	defer h.startShutdownMtx.Unlock()

	h.logger.Debug("Starting Heartbeat routine")
	if h.cancel != nil {
		return ErrAlreadyRunning
	}

	err := h.exporter.start(ctx, host)
	if err != nil {
		return err
	}

	var loopCtx context.Context
	loopCtx, h.cancel = context.WithCancel(context.Background())
	go h.loop(loopCtx)
	return nil
}

func (h *Heartbeat) Shutdown(ctx context.Context) error {
	h.startShutdownMtx.Lock()
	defer h.startShutdownMtx.Unlock()

	h.logger.Debug("Stopping Heartbeat routine")
	if h.cancel == nil {
		// already stopped
		return nil
	}
	h.cancel()
	h.cancel = nil
	return h.exporter.shutdown(ctx)
}

func (h *Heartbeat) loop(ctx context.Context) {
	tick := time.NewTicker(h.beatInterval)
	defer tick.Stop()

	// Start beat
	if err := h.generate(ctx); err != nil {
		h.logger.Error("Generating heartbeat failed", zap.Error(err))
	}

	for {
		select {
		case <-tick.C:
			if err := h.generate(ctx); err != nil {
				h.logger.Error("Generating heartbeat failed", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}

}

func (h *Heartbeat) generate(ctx context.Context) error {
	h.logger.Debug("Generating heartbeat")
	md := pmetric.NewMetrics()

	if err := h.metric.add(ctx, md); err != nil {
		return err
	}

	for i, rms := 0, md.ResourceMetrics(); i < rms.Len(); i++ {
		rm := rms.At(i)
		if err := h.decorateResourceAttributes(rm.Resource()); err != nil {
			return err
		}
	}

	return h.exporter.push(ctx, md)
}

func (h *Heartbeat) decorateResourceAttributes(resource pcommon.Resource) error {
	if h.resource != nil {
		for key, value := range h.resource {
			resource.Attributes().PutStr(key, value)
		}
	}

	if h.collectorName != "" {
		resource.Attributes().PutStr(CollectorNameAttribute, h.collectorName)
	}
	return nil
}
