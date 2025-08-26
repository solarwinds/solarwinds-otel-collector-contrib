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

package solarwindsprocessor

import (
	"context"
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/solarwindsprocessor/internal"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type mockHost struct {
	component.Host
}

// ExtensionProvider mock
type mockExtensionProvider struct{}

func (m mockExtensionProvider) Init(*zap.Logger, string, component.Host) (*solarwindsextension.SolarwindsExtension, error) {
	return nil, nil
}

func (m mockExtensionProvider) SetAttributes(resourceAttributes *map[string]string) error {
	attrs := *resourceAttributes
	if attrs == nil {
		attrs = make(map[string]string)
	}

	attrs["sw.otelcol.collector.name"] = "test-collector-name"
	attrs["sw.otelcol.collector.entity_creation"] = "on"

	return nil
}

var _ ExtensionProvider = (*mockExtensionProvider)(nil)

// ExtensionProvider mock that fails on Init
type mockExtensionProviderInitError struct{}

func (m mockExtensionProviderInitError) Init(*zap.Logger, string, component.Host) (*solarwindsextension.SolarwindsExtension, error) {
	return nil, fmt.Errorf("mock extension init error")
}

func (m mockExtensionProviderInitError) SetAttributes(resourceAttributes *map[string]string) error {
	return nil
}

var _ ExtensionProvider = (*mockExtensionProviderInitError)(nil)

func newTestProcessor(t *testing.T, cfg *Config, ha *internal.HostAttributesDecorator) *solarwindsprocessor {
	return &solarwindsprocessor{
		logger:            zaptest.NewLogger(t),
		cfg:               cfg,
		extensionProvider: mockExtensionProvider{},
		hostAttributes:    ha,
	}
}

func newTestProcessorWithHostAttributes(t *testing.T, cfg *Config, hostAttrs *internal.HostAttributesDecorator) *solarwindsprocessor {
	return &solarwindsprocessor{
		logger:            zaptest.NewLogger(t),
		cfg:               cfg,
		hostAttributes:    hostAttrs,
		extensionProvider: mockExtensionProvider{},
	}
}

func TestResourceAttributesPrecedenceOverHostAttributes(t *testing.T) {
	cfg := &Config{
		ExtensionName: "test",
		ResourceAttributes: map[string]string{
			"host.id":   "resource-host-id",
			"host.name": "resource-host-name",
			"os.type":   "windows",
		},
		HostAttributesDecoration: internal.HostDecoration{
			Enabled:          true,
			OnPremOverrideID: "client-id-123",
		},
	}

	hostAttrs := &internal.HostAttributesDecorator{
		ContainerID:       "container-abc",
		IsRunInContainerd: true,
		OnPremOverrideId:  "client-id-123",
	}

	proc := newTestProcessorWithHostAttributes(t, cfg, hostAttrs)

	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	attrs := rm.Resource().Attributes()
	attrs.PutStr("host.id", "original-host-id")
	attrs.PutStr("host.name", "original-host-name")
	attrs.PutStr("os.type", "windows")

	out, err := proc.processMetrics(context.Background(), metrics)
	assert.NoError(t, err)

	outAttrs := out.ResourceMetrics().At(0).Resource().Attributes()
	val, _ := outAttrs.Get("host.id")
	assert.Equal(t, "resource-host-id", val.Str(), "resource attribute 'host.id' did not take precedence")

	val, _ = outAttrs.Get("host.name")
	assert.Equal(t, "resource-host-name", val.Str(), "resource attribute 'host.name' did not take precedence")

	val, _ = outAttrs.Get("os.type")
	assert.Equal(t, "windows", val.Str(), "resource attribute 'os.type' did not take precedence")
}

func TestProcessorDoesNotFailWhenHostDecorationDisabled(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: internal.HostDecoration{
			Enabled:          false,
			OnPremOverrideID: "",
		},
	}

	p := newTestProcessor(t, cfg, nil)
	metrics := pmetric.NewMetrics()
	metrics.ResourceMetrics().AppendEmpty()
	_, err := p.processMetrics(context.Background(), metrics)
	assert.NoError(t, err, "processor failed when host decoration disabled")
}

func TestProcessorDoesNotFailOnStartWhenHostDecorationDisabled(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: internal.HostDecoration{
			Enabled:          false,
			OnPremOverrideID: "",
		},
	}

	p := newTestProcessor(t, cfg, nil)
	err := p.start(context.Background(), &mockHost{})
	assert.NoError(t, err, "processor start failed when host decoration disabled")
}

func TestProcessorStartWithExtensionProviderError(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: internal.HostDecoration{
			Enabled: false,
		},
	}

	p := &solarwindsprocessor{
		logger:            zaptest.NewLogger(t),
		cfg:               cfg,
		extensionProvider: mockExtensionProviderInitError{},
		hostAttributes:    nil,
	}

	err := p.start(context.Background(), &mockHost{})
	assert.Error(t, err, "processor start should fail when extension provider init fails")
	assert.Contains(t, err.Error(), "mock extension init error", "expected error message not found")
}

func TestHostDecorationInAllSignalTypes(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: internal.HostDecoration{
			Enabled:          true,
			OnPremOverrideID: "client-id-xyz",
		},
	}

	hostAttrs := &internal.HostAttributesDecorator{
		ContainerID:       "container-xyz",
		IsRunInContainerd: true,
		OnPremOverrideId:  "client-id-xyz",
	}

	proc := newTestProcessorWithHostAttributes(t, cfg, hostAttrs)

	// Helper to set initial attributes
	setInitialAttributes := func(attrs pcommon.Map) {
		attrs.PutStr("host.id", "original-host-id")
		attrs.PutStr("host.name", "original-host-name")
		attrs.PutStr("os.type", "linux")
	}

	// Helper to assert attributes
	assertAttributes := func(t *testing.T, attrs pcommon.Map, signalType string) {
		val, _ := attrs.Get("host.id")
		assert.Equal(t, "client-id-xyz", val.Str(), "host.id not decorated in %s", signalType)

		val, _ = attrs.Get("host.name")
		assert.Equal(t, "original-host-name", val.Str(), "host.name should not be overridden in %s when not in AWS cloud", signalType)

		val, _ = attrs.Get("os.type")
		assert.Equal(t, "Linux", val.Str(), "os.type not normalized in %s", signalType)
	}

	testCases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "metrics",
			testFunc: func(t *testing.T) {
				metrics := pmetric.NewMetrics()
				rm := metrics.ResourceMetrics().AppendEmpty()
				setInitialAttributes(rm.Resource().Attributes())

				out, err := proc.processMetrics(context.Background(), metrics)
				assert.NoError(t, err)
				assertAttributes(t, out.ResourceMetrics().At(0).Resource().Attributes(), "metrics")
			},
		},
		{
			name: "logs",
			testFunc: func(t *testing.T) {
				logs := plog.NewLogs()
				rl := logs.ResourceLogs().AppendEmpty()
				setInitialAttributes(rl.Resource().Attributes())

				out, err := proc.processLogs(context.Background(), logs)
				assert.NoError(t, err)
				assertAttributes(t, out.ResourceLogs().At(0).Resource().Attributes(), "logs")
			},
		},
		{
			name: "traces",
			testFunc: func(t *testing.T) {
				traces := ptrace.NewTraces()
				rs := traces.ResourceSpans().AppendEmpty()
				setInitialAttributes(rs.Resource().Attributes())

				out, err := proc.processTraces(context.Background(), traces)
				assert.NoError(t, err)
				assertAttributes(t, out.ResourceSpans().At(0).Resource().Attributes(), "traces")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestProcessorWithNilHostAttributes(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{"custom.attr": "value"},
		HostAttributesDecoration: internal.HostDecoration{
			Enabled: false,
		},
	}

	proc := newTestProcessor(t, cfg, nil)
	// hostAttributes should be nil when decoration is disabled

	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	attrs := rm.Resource().Attributes()
	attrs.PutStr("original.attr", "original-value")

	out, err := proc.processMetrics(context.Background(), metrics)
	assert.NoError(t, err)

	outAttrs := out.ResourceMetrics().At(0).Resource().Attributes()
	val, _ := outAttrs.Get("custom.attr")
	assert.Equal(t, "value", val.Str(), "custom resource attribute should be added")

	val, _ = outAttrs.Get("original.attr")
	assert.Equal(t, "original-value", val.Str(), "original attribute should be preserved")
}
