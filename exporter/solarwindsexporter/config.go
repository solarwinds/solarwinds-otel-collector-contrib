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

package solarwindsexporter

import (
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

// Config represents a Solarwinds Exporter configuration.
type Config struct {
	// Extension identifies a Solarwinds Extension to
	// use for obtaining connection credentials in this exporter.
	Extension string `mapstructure:"extension"`
	// BackoffSettings configures retry behavior of the exporter.
	// See [configretry.BackOffConfig] documentation.
	BackoffSettings configretry.BackOffConfig `mapstructure:"retry_on_failure"`
	// QueueSettings defines configuration for queueing batches in the OTLP Exporter.
	// See [exporterhelper.QueueBatchConfig] documentation.
	QueueSettings exporterhelper.QueueBatchConfig `mapstructure:"sending_queue"`
	// Timeout configures timeout in the underlying OTLP exporter.
	Timeout exporterhelper.TimeoutConfig `mapstructure:"timeout,squash"`

	// Fields below populated from Solarwinds Extension at runtime:
	// ingestionToken stores the token used to export telemetry.
	ingestionToken configopaque.String `mapstructure:"-"`
	// endpointURL stores the URL for exporting telemetry.
	endpointURL string `mapstructure:"-"`
	// collectorName is added as an attribute to telemetry.
	collectorName string `mapstructure:"-"`
	withoutEntity bool   `mapstructure:"-"`
}

// extensionAsComponent tries to parse `extension` value of the form 'type/name'
// or 'type' from the configuration to [component.ID]. If the `extension value is empty,
// it returns `nil` with a `nil` error.
//
// It uses [component.ID.UnmarshalText] and behaves accordingly.
func (cfg *Config) extensionAsComponent() (*component.ID, error) {
	if cfg.Extension == "" {
		return nil, nil
	}

	parsedID := &component.ID{}
	err := parsedID.UnmarshalText([]byte(cfg.Extension))

	return parsedID, err
}

// NewDefaultConfig creates a new default configuration.
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
	if len(cfg.Extension) != 0 {
		_, err := cfg.extensionAsComponent()
		if err != nil {
			return fmt.Errorf(
				"invalid configuration: %q is not a correct value for 'extension'",
				cfg.Extension,
			)
		}
	}

	return nil
}

// OTLPConfig generates a full OTLP Exporter configuration from the configuration.
func (cfg *Config) OTLPConfig() (*otlpexporter.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Headers - set bearer auth.
	bearer := fmt.Sprintf("Bearer %s", string(cfg.ingestionToken))
	headers := map[string]configopaque.String{
		"Authorization": configopaque.String(bearer),
	}

	// gRPC client configuration.
	clientCfg := configgrpc.ClientConfig{
		TLS:          configtls.NewDefaultClientConfig(),
		Keepalive:    configoptional.Some(configgrpc.NewDefaultKeepaliveClientConfig()),
		BalancerName: configgrpc.BalancerName(),
		Headers:      headers,
		Endpoint:     cfg.endpointURL,
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
