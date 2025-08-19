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
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

var (
	endpointMustBeSetError        = errors.New("invalid configuration: 'endpoint' must be set")
	notValidAuthorizationFoundErr = errors.New("invalid configuration: not valid 'Authorization' found in 'headers', use 'headers: {\"Authorization\": \"Bearer ${YOUR_TOKEN}\"}'")
	collectorNameMustBeSetErr     = errors.New("invalid configuration: 'collector_name' must be set")
)

// Config represents a Solarwinds Extension configuration.
type Config struct {
	// CollectorName name of the collector passed in the heartbeat metric
	CollectorName           string            `mapstructure:"collector_name"`
	Resource                map[string]string `mapstructure:"resource"`
	WithoutEntity           bool              `mapstructure:"without_entity"`
	configgrpc.ClientConfig `mapstructure:"grpc"`
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
	if cfg.Endpoint == "" {
		return endpointMustBeSetError
	}

	if val, found := cfg.Headers["Authorization"]; !found || val == "" {
		return notValidAuthorizationFoundErr
	}

	if cfg.CollectorName == "" {
		return collectorNameMustBeSetErr
	}

	return nil
}

// OTLPConfig generates a full OTLP Exporter configuration from the configuration.
func (cfg *Config) OTLPConfig() (*otlpexporter.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// gRPC client configuration.
	otlpConfig := &otlpexporter.Config{
		ClientConfig: configgrpc.ClientConfig{
			TLS:          configtls.NewDefaultClientConfig(),
			Keepalive:    configoptional.Some(configgrpc.NewDefaultKeepaliveClientConfig()),
			BalancerName: configgrpc.BalancerName(),
			Headers:      cfg.Headers,
			Endpoint:     cfg.Endpoint,
		},
	}

	if err := otlpConfig.Validate(); err != nil {
		return nil, err
	}

	return otlpConfig, nil
}
