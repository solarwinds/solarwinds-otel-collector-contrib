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

package timezone

import (
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/wmi"
)

type provider struct {
	wmi wmi.Executor
}

var _ providers.Provider[TimeZone] = (*provider)(nil)

func CreateTimeZoneProvider() providers.Provider[TimeZone] {
	return &provider{
		wmi: wmi.NewExecutor(),
	}
}

// Win32_TimeZone represents actual Time Zone WMI Object
// with subset of fields required for scraping.
type Win32_TimeZone struct {
	Bias         int32
	Caption      string
	StandardName string
}

// Provide implements Provider.
func (tp *provider) Provide() <-chan TimeZone {
	ch := make(chan TimeZone)
	go func() {
		defer close(ch)
		result, err := wmi.QuerySingleResult[Win32_TimeZone](tp.wmi)
		if err == nil {
			ch <- TimeZone{Bias: int(result.Bias), StandardName: result.StandardName, Caption: result.Caption}
		}
	}()
	return ch
}
