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

package solarwindsexporter

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

// dataCenterToURLMapping maps a data center ID to
// to its corresponding OTLP endpoint URL.
var dataCenterToURLMapping = map[string]string{
	"na-01": "otel.collector.na-01.cloud.solarwinds.com:443",
	"na-02": "otel.collector.na-02.cloud.solarwinds.com:443",
	"eu-01": "otel.collector.eu-01.cloud.solarwinds.com:443",
}

// lookupDataCenterURL returns the OTLP endpoint URL
// for a `dc` data center ID. Matching is case-insensitive.
// It fails with an error if `dc` doesn't identify a data center.
func lookupDataCenterURL(dc string) (string, error) {
	dcLowercase := strings.ToLower(dc)

	url, ok := dataCenterToURLMapping[dcLowercase]
	if !ok {
		return "", fmt.Errorf("unknown data center ID: %s", dc)
	}

	return url, nil
}

// Config represents a Solarwinds Exporter configuration.
type Config struct {
	// DataCenter ID (e.g. na-01).
	DataCenter string `mapstructure:"data_center"`
	// EndpointURLOverride sets OTLP endpoint directly.
	// Warning: Intended for testing use only, use `DataCenter` instead.
	EndpointURLOverride string `mapstructure:"endpoint_url_override"`
	// IngestionToken is your secret generated SWO ingestion token.
	IngestionToken configopaque.String `mapstructure:"token"`
	// BackoffSettings configures retry behavior of the exporter.
	// See [configretry.BackOffConfig] documentation.
	BackoffSettings configretry.BackOffConfig `mapstructure:"retry_on_failure"`
	// QueueSettings defines configuration for queueing batches in the OTLP Exporter.
	// See [exporterhelper.QueueConfig] documentation.
	QueueSettings exporterhelper.QueueConfig `mapstructure:"sending_queue"`
	// Timeout configures timeout in the underlying OTLP exporter.
	Timeout exporterhelper.TimeoutConfig `mapstructure:"timeout,squash"`
}

// NewDefaultConfig creates a new default configuration.
//
// Warning: it doesn't define mandatory `Token` and `DataCenter`
// fields that need to be explicitly provided.
func NewDefaultConfig() component.Config {
	// Using a higher default than OTLP Exporter does (5s)
	// based on previous experience with unnecessary timeouts.
	defaultTimeout := exporterhelper.TimeoutConfig{
		Timeout: 10 * time.Second,
	}

	return &Config{
		Timeout:         defaultTimeout,
		BackoffSettings: configretry.NewDefaultBackOffConfig(),
		QueueSettings:   exporterhelper.NewDefaultQueueConfig(),
	}
}

// Validate checks the configuration for its validity.
func (cfg *Config) Validate() error {
	if cfg.DataCenter == "" && cfg.EndpointURLOverride == "" {
		return errors.New("invalid configuration: data center must be provided")
	}

	if _, err := lookupDataCenterURL(cfg.DataCenter); err != nil {
		return fmt.Errorf("invalid data center ID: %w", err)
	}

	if cfg.IngestionToken == "" {
		return errors.New("invalid configuration: token must be set")
	}

	return nil
}

// OTLPConfig generates a full OTLP Exporter configuration from the configuration.
func (cfg *Config) OTLPConfig() (*otlpexporter.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Use overridden URL if provided.
	endpointURL := cfg.EndpointURLOverride
	if endpointURL == "" {
		// Error doesn't need to be checked, it's been validated above.
		endpointURL, _ = lookupDataCenterURL(cfg.DataCenter)
	}

	// Headers - set bearer auth.
	bearer := fmt.Sprintf("Bearer %s", string(cfg.IngestionToken))
	headers := map[string]configopaque.String{
		"Authorization": configopaque.String(bearer),
	}

	// gRPC client configuration.
	clientCfg := configgrpc.ClientConfig{
		TLSSetting:   configtls.NewDefaultClientConfig(),
		Keepalive:    configgrpc.NewDefaultKeepaliveClientConfig(),
		BalancerName: configgrpc.BalancerName(),
		Headers:      headers,
		Endpoint:     endpointURL,
	}

	otlpConfig := &otlpexporter.Config{
		QueueConfig:   cfg.QueueSettings,
		RetryConfig:   cfg.BackoffSettings,
		TimeoutConfig: cfg.Timeout,
		ClientConfig:  clientCfg,
	}

	if err := otlpConfig.Validate(); err != nil {
		return nil, err
	}

	return otlpConfig, nil
}
