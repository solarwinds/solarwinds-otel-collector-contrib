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

package solarwindsextension

import (
	"go.opentelemetry.io/collector/config/configopaque"

	"github.com/solarwinds/solarwinds-otel-collector/extension/solarwindsextension/internal"
)

type EndpointConfig interface {
	Url() (string, error)
	Token() configopaque.String
}

type endpointConfig struct{ cfg *internal.Config }

var _ EndpointConfig = (*endpointConfig)(nil)

func newEndpointConfig(cfg *internal.Config) *endpointConfig {
	return &endpointConfig{cfg: cfg}
}

func (c *endpointConfig) Url() (string, error) {
	return c.cfg.EndpointUrl()
}

func (c *endpointConfig) Token() configopaque.String {
	return c.cfg.IngestionToken
}
