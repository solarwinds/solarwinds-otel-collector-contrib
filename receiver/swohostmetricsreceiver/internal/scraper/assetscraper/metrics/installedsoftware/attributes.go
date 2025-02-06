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

package installedsoftware

import "github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/installedsoftware"

// GetAttributes implements types.AttributesProvider.
func getAttributes(is installedsoftware.InstalledSoftware) map[string]any {
	const (
		name                    = "installedsoftware.name"
		date                    = "installedsoftware.date"
		publisher               = "installedsoftware.publisher"
		version                 = "installedsoftware.version"
		maximumSizeOfAttributes = 4
	)

	attributes := make(map[string]any, maximumSizeOfAttributes)
	if len(is.Name) > 0 {
		attributes[name] = is.Name
	}
	if len(is.InstallDate) > 0 {
		attributes[date] = is.InstallDate
	}
	if len(is.Publisher) > 0 {
		attributes[publisher] = is.Publisher
	}
	if len(is.Version) > 0 {
		attributes[version] = is.Version
	}
	return attributes
}
