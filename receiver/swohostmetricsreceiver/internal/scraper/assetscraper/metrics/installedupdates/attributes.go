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

package installedupdates

import (
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/installedupdates"
)

// GetAttributes implements types.AttributesProvider.
func getAttributes(update installedupdates.InstalledUpdate) map[string]any {
	const (
		caption                 = "installedupdate.caption"
		hotfixid                = "installedupdate.hotfixid"
		installedon             = "installedupdate.installedon"
		installedby             = "installedupdate.installedby"
		description             = "installedupdate.description"
		maximumSizeOfAttributes = 5
	)

	attributes := make(map[string]any, maximumSizeOfAttributes)
	if len(update.Caption) > 0 {
		attributes[caption] = update.Caption
	}
	if len(update.HotFixID) > 0 {
		attributes[hotfixid] = update.HotFixID
	}
	if len(update.InstalledOn) > 0 {
		attributes[installedon] = update.InstalledOn
	}
	if len(update.InstalledBy) > 0 {
		attributes[installedby] = update.InstalledBy
	}
	if len(update.Description) > 0 {
		attributes[description] = update.Description
	}
	return attributes
}
