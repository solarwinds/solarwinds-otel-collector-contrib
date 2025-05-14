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

type InstalledUpdate struct {
	Caption     string
	HotFixID    string
	InstalledOn string
	InstalledBy string
	Description string
}

type Provider interface {
	GetUpdates() ([]InstalledUpdate, error)
}

// Null Object pattern provider, returns nothing, but does not fail.
type noUpdatesProvider struct{}

var _ (Provider) = (*noUpdatesProvider)(nil)

// Create new instance of NullProvider, whitch returns no updates.
func createNoUpdatesProvider() Provider {
	return new(noUpdatesProvider)
}

func (noUpdatesProvider) GetUpdates() ([]InstalledUpdate, error) {
	return []InstalledUpdate{}, nil
}
