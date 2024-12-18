// Copyright 2024 SolarWinds Worldwide, LLC. All rights reserved.
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
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

// Config represents a Solarwinds Extension configuration.
type Config struct {
	// DataCenter ID (e.g. na-01).
	DataCenter string `mapstructure:"data_center"`
	// IngestionToken is your secret generated SolarWinds Observability SaaS ingestion token.
	IngestionToken configopaque.String `mapstructure:"token"`
	// CollectorName name of the collector passed in the heartbeat metric
	CollectorName string `mapstructure:"collector_name"`
	// Insecure disables TLS in the exporters.

	// ⚠️ Warning: For testing purpose only.
	// EndpointURLOverride sets OTLP endpoint directly, it overrides the DataCenter configuration.
	EndpointURLOverride string `mapstructure:"endpoint_url_override"`
	// ⚠️ Warning: For testing purpose only.
	// Insecure disables the TLS security. It can be used only together with EndpointURLOverride.
	Insecure bool `mapstructure:"insecure"`
}

var (
	ErrMissingDataCenter    = errors.New("invalid configuration: 'data_center' must be set")
	ErrMissingToken         = errors.New("invalid configuration: 'token' must be set")
	ErrMissingCollectorName = errors.New("invalid configuration: 'collector_name' must be set")
	ErrInsecureInProd       = errors.New("invalid configuration: 'insecure' is not allowed in production mode")
)

// NewDefaultConfig creates a new default configuration.
//
// Warning: it doesn't define mandatory `Token` and `DataCenter`
// fields that need to be explicitly provided.
func NewDefaultConfig() component.Config {
	return &Config{}
}

// Validate checks the configuration for its validity.
func (cfg *Config) Validate() error {
	if cfg.DataCenter == "" && cfg.EndpointURLOverride == "" {
		return ErrMissingDataCenter
	}

	if cfg.Insecure && cfg.EndpointURLOverride == "" {
		return ErrInsecureInProd
	}

	if _, err := cfg.EndpointUrl(); err != nil {
		return fmt.Errorf("invalid 'data_center' value: %w", err)
	}

	if cfg.IngestionToken == "" {
		return ErrMissingToken
	}
	if cfg.CollectorName == "" {
		return ErrMissingCollectorName
	}

	return nil
}

// OTLPConfig generates a full OTLP Exporter configuration from the configuration.
func (cfg *Config) OTLPConfig() (*otlpexporter.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	endpointURL, err := cfg.EndpointUrl()
	if err != nil {
		return nil, err
	}

	// Headers - set bearer auth.
	bearer := configopaque.String(fmt.Sprintf("Bearer %s", string(cfg.IngestionToken)))
	headers := map[string]configopaque.String{
		"Authorization": bearer,
	}

	// gRPC client configuration.
	otlpConfig := &otlpexporter.Config{
		ClientConfig: configgrpc.ClientConfig{
			TLSSetting:   configtls.NewDefaultClientConfig(),
			Keepalive:    configgrpc.NewDefaultKeepaliveClientConfig(),
			BalancerName: configgrpc.BalancerName(),
			Headers:      headers,
			Endpoint:     endpointURL,
		},
	}

	// Disable TLS for testing.
	if cfg.Insecure {
		otlpConfig.ClientConfig.TLSSetting.Insecure = true
	}

	if err = otlpConfig.Validate(); err != nil {
		return nil, err
	}

	return otlpConfig, nil
}

func (cfg *Config) EndpointUrl() (string, error) {
	// Use overridden URL if provided.
	if cfg.EndpointURLOverride != "" {
		return cfg.EndpointURLOverride, nil
	}
	return lookupDataCenterURL(cfg.DataCenter)
}

// dataCenterToURLMapping maps a data center ID to
// to its corresponding OTLP endpoint URL.
var dataCenterToURLMapping = map[string]string{
	"na-01": "otel.collector.na-01.cloud.solarwinds.com:4317",
	"na-02": "otel.collector.na-02.cloud.solarwinds.com:4317",
	"eu-01": "otel.collector.eu-01.cloud.solarwinds.com:4317",
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
