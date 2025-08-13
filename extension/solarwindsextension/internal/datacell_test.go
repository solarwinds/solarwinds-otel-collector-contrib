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

package internal

import (
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigValidateDataCenters verifies mappings
// for data centers (the mapping is case-insensitive).
func TestConfigValidateDataCenters(t *testing.T) {
	type test struct {
		dataCenter string
		url        string
		ok         bool
	}

	tests := []test{
		{dataCenter: "na-01", url: "otel.collector.na-01.cloud.solarwinds.com:443", ok: true},
		{dataCenter: "na-02", url: "otel.collector.na-02.cloud.solarwinds.com:443", ok: true},
		{dataCenter: "eu-01", url: "otel.collector.eu-01.cloud.solarwinds.com:443", ok: true},
		{dataCenter: "ap-01", url: "otel.collector.ap-01.cloud.solarwinds.com:443", ok: true},
		{dataCenter: "NA-01", url: "otel.collector.na-01.cloud.solarwinds.com:443", ok: true},
		{dataCenter: "apj-01", url: "", ok: false},
	}

	for _, tc := range tests {
		// Try to find a dataCenter URL for its ID.
		url, err := config.LookupDataCenterURL(tc.dataCenter)

		if tc.ok { // A URL should be returned.
			require.NoError(t, err)
			assert.Equal(t, tc.url, url)
		} else { // It should fail.
			assert.ErrorContains(t, err, "unknown data center ID")
		}
	}
}
