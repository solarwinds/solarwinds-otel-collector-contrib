package solarwindsprocessor

import (
	"context"
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
		HostAttributesDecoration: HostDecoration{
			Enabled:  false,
			ClientId: "",
		},
	}

	p := newTestProcessor(t, cfg)
	metrics := pmetric.NewMetrics()
	metrics.ResourceMetrics().AppendEmpty()
	_, err := p.processMetrics(context.Background(), metrics)
	assert.NoError(t, err, "processor failed when host decoration disabled")
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
	assert.Error(t, err, "expected error when host decoration enabled without client id")
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

	assert.NoError(t, err, "processor start failed")
	assert.Equal(t, "client-id-xyz", p.hostAttributes.ClientId, "hostAttributes not set correctly")
}

func TestHostDecorationInAllSignalTypes(t *testing.T) {
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
