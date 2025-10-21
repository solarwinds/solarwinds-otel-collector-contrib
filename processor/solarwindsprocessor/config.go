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

package solarwindsprocessor

import (
	"fmt"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/solarwindsprocessor/internal"
)

type Config struct {
	// Limit for maximum size (in MiB) of the serialized signal payload.
	// When maximum size is set greater to zero, limit check on serialized content is
	// performed. When the serialized content exceeds the limit, it is reported to log.
	// When maximum size is set to zero, no limit check is performed.
	MaxSizeMib int `mapstructure:"max_size_mib,omitempty"`
	// Resource attributes to be added to the processed signals.
	ResourceAttributes            map[string]string       `mapstructure:"resource,omitempty"`
	CollectorAttributesDecoration CollectorDecoration     `mapstructure:"collector_attributes_decoration,omitempty"`
	HostAttributesDecoration      internal.HostDecoration `mapstructure:"host_attributes_decoration,omitempty"`
}

func (c *Config) Validate() error {
	if err := c.CollectorAttributesDecoration.Validate(); err != nil {
		return err
	}

	if c.MaxSizeMib < 0 {
		return fmt.Errorf("invalid configuration: 'max_size_mib' must be greater than or equal to zero, got: %d", c.MaxSizeMib)
	}

	return nil
}
