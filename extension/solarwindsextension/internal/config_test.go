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

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"
)

func TestConfigUnmarshalFull(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "full")

	cfg := &Config{}
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.Equal(t, &Config{
		CollectorName: "test-name",
		Resource:      map[string]string{"att": "custom_attribute_value"},
		Grpc: configgrpc.ClientConfig{
			Endpoint: "test-url:0",
			TLS:      configtls.ClientConfig{Insecure: true},
			Headers:  map[string]configopaque.String{"Authorization": "Bearer test-token"},
		},
		WithoutEntity: true,
	}, cfg)
}

func TestConfigValidateOK(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "minimal")

	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.NoError(t, cfg.(*Config).Validate())
}

func TestConfigValidateMissingAuthorizationHeader(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "missing_authorization_header")

	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.ErrorContains(t, cfg.(*Config).Validate(), "'Authorization' header must be set")
}

func TestConfigValidateInvalidAuthorizationHeader(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "invalid_authorization_header")

	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.ErrorContains(t, cfg.(*Config).Validate(), "the 'Authorization' header is invalid")
}

func TestConfigValidateMissingEndpoint(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "missing_endpoint")

	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.ErrorContains(t, cfg.(*Config).Validate(), "'endpoint' must be set")
}

func TestConfigValidateMissingCollectorName(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "missing_collector_name")

	cfg := NewDefaultConfig()
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	assert.ErrorContains(t, cfg.(*Config).Validate(), "'collector_name' must be set")
}

func TestConfigOTLPOk(t *testing.T) {
	cfgFile := testutil.LoadConfigTestdata(t, "full")

	cfg := &Config{}
	require.NoError(t, cfgFile.Unmarshal(&cfg))

	otlpCfg, err := cfg.OTLPConfig()
	require.NoError(t, err)

	assert.Equal(t, "test-url:0", otlpCfg.ClientConfig.Endpoint)
	assert.Equal(
		t,
		map[string]configopaque.String{"Authorization": "Bearer test-token"},
		otlpCfg.ClientConfig.Headers,
	)
}
