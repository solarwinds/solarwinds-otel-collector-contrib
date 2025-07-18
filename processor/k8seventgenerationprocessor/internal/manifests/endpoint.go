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

type ServiceMapping interface {
	GetAddresses() []string
	GetServiceName() string
	GetNamespace() string
}

type EndpointManifest struct {
	Metadata endpointMetadata `json:"metadata"`
	Subsets  []endpointSubset `json:"subsets"`
}

type endpointMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type endpointSubset struct {
	Addresses         []endpointAddress `json:"addresses"`
	NotReadyAddresses []endpointAddress `json:"notReadyAddresses"`
}

type endpointAddress struct {
	Ip string `json:"ip"`
}

// GetAddresses returns a slice of addresses from the Endpoint manifest.
func (m *EndpointManifest) GetAddresses() []string {
	addresses := make(map[string]any, len(m.Subsets))
	for _, subset := range m.Subsets {
		for _, address := range subset.Addresses {
			addresses[address.Ip] = nil
		}
		for _, address := range subset.NotReadyAddresses {
			addresses[address.Ip] = nil
		}
	}

	return slices.Collect(maps.Keys(addresses))
}

func (m *EndpointManifest) GetServiceName() string {
	return m.Metadata.Name
}

func (m *EndpointManifest) GetNamespace() string {
	return m.Metadata.Namespace
}
