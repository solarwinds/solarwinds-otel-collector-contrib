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

package loggedusers

import (
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/registry"
)

type provider struct {
	getRegistryValues registry.GetKeyValuesTypeFunc
}

var _ providers.Provider[Data] = (*provider)(nil)

func CreateProvider() providers.Provider[Data] {
	return &provider{getRegistryValues: registry.GetKeyValues}
}

func (dp *provider) Provide() <-chan Data {
	ch := make(chan Data)
	go dp.provideInternal(ch)
	return ch
}

func (dp *provider) provideInternal(ch chan Data) {
	defer close(ch)
	values, err := dp.getRegistryValues(
		registry.LocalMachineKey,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Authentication`,
		`LogonUI`,
		[]string{`LastLoggedOnUser`, `LastLoggedOnDisplayName`},
	)
	if err != nil {
		ch <- Data{Error: err}
		return
	}

	ch <- Data{
		Users: []User{{
			Name:        values[`LastLoggedOnUser`],
			DisplayName: values[`LastLoggedOnDisplayName`],
		}},
	}
}
