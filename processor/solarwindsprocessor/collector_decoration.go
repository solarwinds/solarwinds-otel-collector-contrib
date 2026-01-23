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

import "fmt"

type CollectorDecoration struct {
	Enabled bool `mapstructure:"enabled"`
	// Extension identifies a Solarwinds Extension to
	// use for obtaining required configuration.
	ExtensionName string `mapstructure:"extension"`
}

func (cd *CollectorDecoration) Validate() error {
	if cd.Enabled && cd.ExtensionName == "" {
		return fmt.Errorf("invalid configuration: 'extension' must be set in 'collector_attributes_decoration'")
	}
	return nil
}
