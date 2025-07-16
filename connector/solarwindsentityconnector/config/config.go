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

package config

import (
	"errors"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
)

type Config struct {
	Schema     Schema           `mapstructure:"schema"`
	Expiration ExpirationPolicy `mapstructure:"expiration_policy" yaml:"expiration_policy"`

	SourcePrefix      string `mapstructure:"source_prefix" yaml:"source_prefix"`
	DestinationPrefix string `mapstructure:"destination_prefix" yaml:"destination_prefix"`
}

func NewDefaultConfig() component.Config {
	return &Config{
		Expiration: ExpirationPolicy{
			Enabled:  true,
			Interval: defaultInterval.String(),
			CacheConfiguration: &CacheConfiguration{
				TTLCleanupInterval: defaultTTLCleanupInterval.String(),
				MaxCapacity:        defaultMaxCapacity,
			},
		},
	}
}

func (c *Config) Validate(logger *zap.Logger) error {
	return errors.Join(
		c.Schema.Validate(logger),
		c.Expiration.Validate(),
	)
}
