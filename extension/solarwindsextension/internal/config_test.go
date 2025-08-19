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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/testutil"
)

// TestConfigUnmarshalFull tries to parse a configuration file
// with all values provided and verifies the configuration.
func TestConfigUnmarshalFull(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "full")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	attributeMap := make(map[string]string)
	attributeMap["att1"] = "custom_attribute_value_1"
	attributeMap["att2"] = "custom_attribute_value_2"

	// Verify the values.
	assert.Equal(t, &Config{
		CollectorName: "test-collector",
		Resource:      attributeMap,
		WithoutEntity: true,
	}, cfg)
}

// TestConfigValidateOK verifies that a configuration
// file containing only the mandatory values successfully
// validates.
func TestConfigValidateOK(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "minimal")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	// Try to validate it.
	assert.NoError(t, cfg.(*Config).Validate())
}

// TestConfigValidateMissingToken verifies that
// the validation of a configuration file with
// the token missing fails as expected.
func TestConfigValidateMissingToken(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "missing_token")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.ErrorContains(
		t,
		cfg.(*Config).Validate(),
		"'token' must be set",
	)
}

// TestConfigValidateMissingDataCenter verifies that
// the validation of a configuration file with
// the data center ID missing fails as expected.
func TestConfigValidateMissingDataCenter(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "missing_dc")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.ErrorContains(
		t,
		cfg.(*Config).Validate(),
		"'data_center' must be set",
	)
}

// TestConfigValidateMissingDataCenter verifies that
// the validation of a configuration file with
// the collector name missing fails as expected.
func TestConfigValidateMissingCollectorName(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "missing_collector_name")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.ErrorContains(
		t,
		cfg.(*Config).Validate(),
		"'collector_name' must be set",
	)
}

// TestConfigOTLPWithOverride converts configuration to
// OTLP gRPC Exporter configuration and verifies that overridden
// endpoint and token propagate correctly.
func TestConfigOTLPWithOverride(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "url_override")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	// Convert it to the OTLP Exporter configuration.
	otlpCfg, err := cfg.(*Config).OTLPConfig()
	require.NoError(t, err)

	// Verify that both the token and overridden URL were propagated
	// to the OTLP configuration.
	assert.Equal(t, "127.0.0.1:1234", otlpCfg.ClientConfig.Endpoint)
	assert.Equal(
		t,
		map[string]configopaque.String{"Authorization": "Bearer YOUR-INGESTION-TOKEN"},
		otlpCfg.ClientConfig.Headers,
	)
}

func TestConfigUnmarshalWithGrpc(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "grpc")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	// Verify the values.
	assert.Equal(
		t,
		&Config{
			CollectorName: "test-collector",
			Resource:      nil,
			WithoutEntity: true,
			ClientConfig: configgrpc.ClientConfig{
				Endpoint: "url",
				TLS:      configtls.ClientConfig{Insecure: false},
				Headers:  map[string]configopaque.String{"Authorization": "token"},
			},
		},
		cfg,
	)
}

func TestConfigOTLPDataCenterOverridenByGrpc(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "grpc_overrides_datacenter")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	// Convert it to the OTLP Exporter configuration.
	otlpCfg, err := cfg.(*Config).OTLPConfig()
	require.NoError(t, err)

	// Verify that the gRPC configuration was propagated correctly.
	assert.Equal(t, "grpc.url:443", otlpCfg.ClientConfig.Endpoint)
}

func TestConfig_UrlOverridendByGrpc(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "grpc_overrides_urloverride")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	// Convert it to the OTLP Exporter configuration.
	otlpCfg, err := cfg.(*Config).OTLPConfig()
	require.NoError(t, err)

	// Verify that the gRPC configuration was propagated correctly.
	assert.Equal(t, "grpc.url:443", otlpCfg.ClientConfig.Endpoint)
}

func TestConfig_TokenOverridendByGrpc(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "grpc_overrides_token")

	// Parse configuration.
	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	// Convert it to the OTLP Exporter configuration.
	otlpCfg, err := cfg.(*Config).OTLPConfig()
	require.NoError(t, err)

	// Verify that the gRPC configuration was propagated correctly.
	assert.Equal(
		t,
		map[string]configopaque.String{"Authorization": "Bearer grpc_token"},
		otlpCfg.ClientConfig.Headers,
	)
}
