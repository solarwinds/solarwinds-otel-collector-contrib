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

package hardwareinventoryscraper

import (
	"context"
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/hardwareinventoryscraper/metrics/cpu"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/types"
	"github.com/stretchr/testify/assert"
)

func Test_Functional(t *testing.T) {
	t.Skip("This test must to be run manually only")

	sc := Config{
		ScraperConfig: types.ScraperConfig{
			Metrics: map[string]types.MetricSettingsConfig{
				cpu.Name: {Enabled: true},
			},
		},
	}

	s, err := NewHardwareInventoryScraper(&sc)
	assert.NoError(t, err, "scraper creation must not fail")
	err = s.Start(context.TODO(), nil)
	assert.NoError(t, err)
	m, err := s.Scrape(context.TODO())
	assert.NoError(t, err)
	fmt.Printf("Result: %+v\n", m)
}
