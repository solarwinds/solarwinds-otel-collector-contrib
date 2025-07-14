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
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpointSliceManifestParsing(t *testing.T) {
	loadTestManifest := func(filepath string) EndpointSliceManifest {
		t.Helper()

		data, err := os.ReadFile(filepath)
		require.NoError(t, err, "Failed to read test data file")

		var manifest EndpointSliceManifest
		err = json.Unmarshal(data, &manifest)
		require.NoError(t, err, "Failed to unmarshal JSON data")

		return manifest
	}

	t.Run("Manifest with all data", func(t *testing.T) {
		manifest := loadTestManifest("./testdata/endpointslice_alldata.json")

		// Test the GetAddresses method
		addresses := manifest.GetAddresses()
		assert.ElementsMatch(t, []string{"192.168.1.1", "192.168.1.2"}, addresses, "Expected addresses to match")

		// Test extracting name and namespace
		assert.Equal(t, "test-name", manifest.Metadata.Labels.ServiceName, "Expected service name to match")
		assert.Equal(t, "test-namespace", manifest.Metadata.Namespace, "Expected service namespace to match")
	})

	t.Run("Manifest with missing data", func(t *testing.T) {
		manifest := loadTestManifest("./testdata/endpointslice_missingdata.json")

		// Test the GetAddresses method
		addresses := manifest.GetAddresses()
		assert.Empty(t, addresses, "Expected no addresses in the manifest")

		// Test extracting name
		assert.Empty(t, manifest.Metadata.Labels.ServiceName, "Expected no service name in the manifest")
	})

	t.Run("Manifest with only necessary data", func(t *testing.T) {
		manifest := loadTestManifest("./testdata/endpointslice_onlynecessarydata.json")

		// Test the GetAddresses method
		addresses := manifest.GetAddresses()
		assert.ElementsMatch(t, []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}, addresses, "Expected addresses to match")

		// Test extracting name and namespace
		assert.Equal(t, "test-name", manifest.Metadata.Labels.ServiceName, "Expected service name to match")
		assert.Equal(t, "test-namespace", manifest.Metadata.Namespace, "Expected service namespace to match")
	})
}
