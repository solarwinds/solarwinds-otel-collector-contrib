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
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

var (
	endpointMustBeSetErr      = errors.New("invalid configuration: 'endpoint' must be set")
	authorizationMustBeSetErr = errors.New("invalid configuration: 'Authorization' header must be set, use 'headers: {\"Authorization\": \"Bearer ${YOUR_TOKEN}\"}'")
	authorizationNotValidErr  = errors.New("invalid configuration: the 'Authorization' header is invalid, use 'headers: {\"Authorization\": \"Bearer ${YOUR_TOKEN}\"}'")
	collectorNameMustBeSetErr = errors.New("invalid configuration: 'collector_name' must be set")
)

type Config struct {
	CollectorName string                  `mapstructure:"collector_name"`
	Resource      map[string]string       `mapstructure:"resource"`
	WithoutEntity bool                    `mapstructure:"without_entity"`
	Grpc          configgrpc.ClientConfig `mapstructure:"grpc"`
}

func NewDefaultConfig() component.Config {
	return &Config{Grpc: configgrpc.NewDefaultClientConfig()}
}

func (cfg *Config) Validate() error {
	var errs error

	if cfg.Grpc.Endpoint == "" {
		errs = errors.Join(errs, endpointMustBeSetErr)
	}

	if val, found := cfg.Grpc.Headers["Authorization"]; !found {
		errs = errors.Join(errs, authorizationMustBeSetErr)
	} else if !strings.HasPrefix(string(val), "Bearer ") {
		errs = errors.Join(errs, authorizationNotValidErr)
	}

	if cfg.CollectorName == "" {
		errs = errors.Join(errs, collectorNameMustBeSetErr)
	}

	return errs
}

func (cfg *Config) ApplyGrpcConfig(otlpConfig *otlpexporter.Config) (*otlpexporter.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	otlpConfig.ClientConfig = cfg.Grpc

	if err := otlpConfig.Validate(); err != nil {
		return nil, err
	}

	return otlpConfig, nil
}
