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

package manifests

import (
	"maps"
	"slices"
)

type EndpointSliceManifest struct {
	Metadata  EndpointSliceMetadata   `json:"metadata"`
	Endpoints []endpointSliceEndpoint `json:"endpoints"`
}

type EndpointSliceMetadata struct {
	Labels struct {
		ServiceName string `json:"kubernetes.io/service-name"`
	} `json:"labels"`
	Namespace string `json:"namespace"`
}

type endpointSliceEndpoint struct {
	Addresses []string `json:"addresses"`
}

// GetAddresses returns a slice of addresses from the EndpointSlice manifest.
func (m *EndpointSliceManifest) GetAddresses() []string {
	addresses := make(map[string]any, len(m.Endpoints))
	for _, endpoint := range m.Endpoints {
		for _, address := range endpoint.Addresses {
			if address != "" {
				addresses[address] = nil
			}
		}
	}

	return slices.Collect(maps.Keys(addresses))
}
