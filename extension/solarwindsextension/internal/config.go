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
	"fmt"
	"maps"
	"slices"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.uber.org/zap"
)

// Config represents a Solarwinds Extension configuration.
type Config struct {
	// DEPRECATED: Use `endpoint` instead.
	// DataCenter ID (e.g. na-01).
	DataCenter string `mapstructure:"data_center"`
	// DEPRECATED: Use `headers` instead.
	// IngestionToken is your secret generated SolarWinds Observability SaaS ingestion token.
	IngestionToken configopaque.String `mapstructure:"token"`
	// CollectorName name of the collector passed in the heartbeat metric
	CollectorName string `mapstructure:"collector_name"`
	// DEPRECATED: Use `endpoint` instead.
	// ⚠️ Warning: For testing purpose only.
	// EndpointURLOverride sets OTLP endpoint directly, it overrides the DataCenter configuration.
	EndpointURLOverride string            `mapstructure:"endpoint_url_override"`
	Resource            map[string]string `mapstructure:"resource"`
	WithoutEntity       bool              `mapstructure:"without_entity"`
	GRPCConfig          `mapstructure:"grpc"`
}

// Config represents a gRPC configuration for the Solarwinds Extension, which can be fully re-used
// in OTLP exporter.
type GRPCConfig struct {
	configgrpc.ClientConfig `mapstructure:",squash"`
}

// NewDefaultConfig creates a new default configuration.
//
// Warning: it doesn't define mandatory `Token` and `DataCenter`
// fields that need to be explicitly provided.
func NewDefaultConfig() component.Config {
	return &Config{}
}

// Validate checks the configuration for its validity.
func (cfg *Config) Validate() error {
	if cfg.GRPCConfig.ClientConfig.Endpoint == "" {
		if cfg.DataCenter == "" && cfg.EndpointURLOverride == "" {
			msg := "invalid configuration: 'data_center' must be set, due to its DEPRECATION use rather 'endpoint' setting"
			return fmt.Errorf("%s", msg)
		}

		if _, err := cfg.EndpointUrl(); err != nil {
			msg := "invalid configuration: 'data_center' must be set, due to its DEPRECATION use rather 'endpoint' setting"
			return fmt.Errorf("%s: %w", msg, err)
		}
	}

	if val, found := cfg.GRPCConfig.ClientConfig.Headers["Authorization"]; !found || val == "" {
		if cfg.IngestionToken == "" {
			msg := "invalid configuration: 'token' must be set, due to its DEPRECATION use rather 'headers: {\"Authorization\": \"Bearer ${YOUR_TOKEN}\"}'"
			return fmt.Errorf("%s", msg)
		}
	}

	if cfg.CollectorName == "" {
		return fmt.Errorf("%s", "invalid configuration: 'collector_name' must be set")
	}

	return nil
}

func (cfg *Config) ReportDeprecatedFields(logger *zap.Logger) {
	if cfg.DataCenter != "" {
		logger.Warn("The 'data_center' field is deprecated, use 'endpoint' instead.")
	}

	if cfg.EndpointURLOverride != "" {
		logger.Warn("The 'endpoint_url_override' field is deprecated, use 'endpoint' instead.")
	}

	if cfg.IngestionToken != "" {
		logger.Warn("The 'token' field is deprecated, use 'headers: {\"Authorization\": \"Bearer ${YOUR_TOKEN}\" instead.")
	}
}

// OTLPConfig generates a full OTLP Exporter configuration from the configuration.
func (cfg *Config) OTLPConfig() (*otlpexporter.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	headers := cfg.Headers()
	endpointURL, err := cfg.EndpointUrl()
	if err != nil {
		return nil, err
	}

	// gRPC client configuration.
	otlpConfig := &otlpexporter.Config{
		ClientConfig: configgrpc.ClientConfig{
			TLS:          configtls.NewDefaultClientConfig(),
			Keepalive:    configoptional.Some(configgrpc.NewDefaultKeepaliveClientConfig()),
			BalancerName: configgrpc.BalancerName(),
			Headers:      headers,
			Endpoint:     endpointURL,
		},
	}

	if err = otlpConfig.Validate(); err != nil {
		return nil, err
	}

	return otlpConfig, nil
}

func (cfg *Config) Headers() map[string]configopaque.String {
	if _, found := cfg.GRPCConfig.ClientConfig.Headers["Authorization"]; found {
		return cfg.GRPCConfig.ClientConfig.Headers
	}

	bearer := configopaque.String(fmt.Sprintf("Bearer %s", string(cfg.IngestionToken)))
	headers := map[string]configopaque.String{
		"Authorization": bearer,
	}
	return headers
}

func (cfg *Config) EndpointUrl() (string, error) {
	// Use configured OTLP endpoint if provided.
	if cfg.GRPCConfig.ClientConfig.Endpoint != "" {
		return cfg.GRPCConfig.ClientConfig.Endpoint, nil
	}
	// Use overridden URL if provided.
	if cfg.EndpointURLOverride != "" {
		return cfg.EndpointURLOverride, nil
	}
	return lookupDataCenterURL(cfg.DataCenter)
}

// dataCenterToURLMapping maps a data center ID to
// to its corresponding OTLP endpoint URL.
var dataCenterToURLMapping = map[string]string{
	"na-01": "otel.collector.na-01.cloud.solarwinds.com:443",
	"na-02": "otel.collector.na-02.cloud.solarwinds.com:443",
	"eu-01": "otel.collector.eu-01.cloud.solarwinds.com:443",
	"ap-01": "otel.collector.ap-01.cloud.solarwinds.com:443",
}

// lookupDataCenterURL returns the OTLP endpoint URL
// for a `dc` data center ID. Matching is case-insensitive.
// It fails with an error if `dc` doesn't identify a data center.
func lookupDataCenterURL(dc string) (string, error) {
	dcLowercase := strings.ToLower(dc)

	url, ok := dataCenterToURLMapping[dcLowercase]
	if !ok {
		return "", fmt.Errorf("unknown data center ID: %s, valid IDs: %s", dc, slices.Collect(maps.Keys(dataCenterToURLMapping)))
	}

	return url, nil
}
