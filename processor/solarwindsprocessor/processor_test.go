package solarwindsprocessor

import (
	"context"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/solarwindsprocessor/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type mockHost struct {
	component.Host
}

// Container provider mock
type mockContainerProvider struct{}

func (mock mockContainerProvider) ReadContainerInstanceID() (string, error) {
	return "container-abc", nil
}

func (mock mockContainerProvider) IsRunInContainerd() bool {
	return true
}

type mockExtensionProvider struct{}

func (m mockExtensionProvider) SetExtension(*zap.Logger, string, component.Host) (*solarwindsextension.SolarwindsExtension, error) {
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

func newTestProcessor(t *testing.T, cfg *Config) *solarwindsprocessor {
	return &solarwindsprocessor{
		logger:            zaptest.NewLogger(t),
		cfg:               cfg,
		containerProvider: mockContainerProvider{},
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
		HostAttributesDecoration: HostDecoration{
			Enabled:  true,
			ClientId: "client-id-123",
		},
	}
	proc := newTestProcessor(t, cfg)
	proc.hostAttributes = internal.HostAttributes{
		ContainerID:       "container-abc",
		IsRunInContainerd: true,
		ClientId:          "client-id-123",
	}

	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	attrs := rm.Resource().Attributes()
	attrs.PutStr("host.id", "original-host-id")
	attrs.PutStr("host.name", "original-host-name")
	attrs.PutStr("os.type", "windows")

	out, err := proc.processMetrics(context.Background(), metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outAttrs := out.ResourceMetrics().At(0).Resource().Attributes()
	if val, _ := outAttrs.Get("host.id"); val.Str() != "resource-host-id" {
		t.Errorf("resource attribute 'host.id' did not take precedence, got: %s", val.Str())
	}
	if val, _ := outAttrs.Get("host.name"); val.Str() != "resource-host-name" {
		t.Errorf("resource attribute 'host.name' did not take precedence, got: %s", val.Str())
	}
	if val, _ := outAttrs.Get("os.type"); val.Str() != "windows" {
		t.Errorf("resource attribute 'os.type' did not take precedence, got: %s", val.Str())
	}
}

func TestProcessorDoesNotFailWhenHostDecorationDisabled(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: HostDecoration{
			Enabled:  false,
			ClientId: "",
		},
	}

	p := newTestProcessor(t, cfg)
	metrics := pmetric.NewMetrics()
	metrics.ResourceMetrics().AppendEmpty()
	_, err := p.processMetrics(context.Background(), metrics)
	if err != nil {
		t.Errorf("processor failed when host decoration disabled: %v", err)
	}
}

func TestProcessorFailsWhenHostDecorationEnabledWithoutClientId(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: HostDecoration{
			Enabled:  true,
			ClientId: "",
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Error("expected error when host decoration enabled without client id, got nil")
	}
}

func TestProcessorStartPollsHostAttributesWhenEnabled(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "solarwinds",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: HostDecoration{
			Enabled:  true,
			ClientId: "client-id-xyz",
		},
	}

	p := newTestProcessor(t, cfg)
	err := p.start(context.Background(), &mockHost{})

	if err != nil {
		t.Errorf("processor start failed: %v", err)
	}
	if p.hostAttributes.ClientId != "client-id-xyz" {
		t.Errorf("hostAttributes not set correctly, got: %s", p.hostAttributes.ClientId)
	}
}

func TestHostDecorationInProcessMetrics(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: HostDecoration{
			Enabled:  true,
			ClientId: "client-id-xyz",
		},
	}
	proc := newTestProcessor(t, cfg)
	proc.hostAttributes = internal.HostAttributes{
		ContainerID:       "container-xyz",
		IsRunInContainerd: true,
		ClientId:          "client-id-xyz",
	}

	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	attrs := rm.Resource().Attributes()
	attrs.PutStr("host.id", "original-host-id")
	attrs.PutStr("host.name", "original-host-name")
	attrs.PutStr("os.type", "linux")

	out, err := proc.processMetrics(context.Background(), metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outAttrs := out.ResourceMetrics().At(0).Resource().Attributes()
	if val, _ := outAttrs.Get("host.id"); val.Str() != "client-id-xyz" {
		t.Errorf("host.id not decorated by host attributes, got: %s", val.Str())
	}
	if val, _ := outAttrs.Get("host.name"); val.Str() != "original-host-name" {
		t.Errorf("host.name should not be overriden when not in AWS cloud, got: %s", val.Str())
	}
	if val, _ := outAttrs.Get("os.type"); val.Str() != "Linux" {
		t.Errorf("os.type not normalized, got: %s", val.Str())
	}
}

func TestHostDecorationInProcessLogsAndTraces(t *testing.T) {
	cfg := &Config{
		ExtensionName:      "test",
		ResourceAttributes: map[string]string{},
		HostAttributesDecoration: HostDecoration{
			Enabled:  true,
			ClientId: "client-id-xyz",
		},
	}
	proc := newTestProcessor(t, cfg)
	proc.hostAttributes = internal.HostAttributes{
		ContainerID:       "container-xyz",
		IsRunInContainerd: true,
		ClientId:          "client-id-xyz",
	}

	// Logs
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	attrs := rl.Resource().Attributes()
	attrs.PutStr("host.id", "original-host-id")
	attrs.PutStr("host.name", "original-host-name")
	attrs.PutStr("os.type", "linux")
	outLogs, err := proc.processLogs(context.Background(), logs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outAttrs := outLogs.ResourceLogs().At(0).Resource().Attributes()
	if val, _ := outAttrs.Get("host.id"); val.Str() != "client-id-xyz" {
		t.Errorf("host.id not decorated in logs, got: %s", val.Str())
	}

	// Traces
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	attrs2 := rs.Resource().Attributes()
	attrs2.PutStr("host.id", "original-host-id")
	attrs2.PutStr("host.name", "original-host-name")
	attrs2.PutStr("os.type", "linux")
	outTraces, err := proc.processTraces(context.Background(), traces)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outAttrs2 := outTraces.ResourceSpans().At(0).Resource().Attributes()
	if val, _ := outAttrs2.Get("host.id"); val.Str() != "client-id-xyz" {
		t.Errorf("host.id not decorated in traces, got: %s", val.Str())
	}
}
