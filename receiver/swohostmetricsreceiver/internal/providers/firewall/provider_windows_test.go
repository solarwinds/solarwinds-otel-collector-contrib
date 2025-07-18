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

package firewall

import (
	"fmt"
	registry2 "github.com/solarwinds/solarwinds-otel-collector-contrib/tools/registry"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Provide_ProvidesFirewallContainerOnSucceedingCliCommand(t *testing.T) {
	expectedProfiles := []Profile{
		{Name: "Domain", Enabled: 1},
		{Name: "Private", Enabled: 0},
		{Name: "Public", Enabled: 1},
	}

	sut := provider{
		getRegistryValue: getMockRegistryValueFunc(map[string]uint64{
			"StandardProfile": 0,
			"DomainProfile":   1,
			"PublicProfile":   1,
		}),
	}

	ch := sut.Provide()
	actualModel := <-ch
	_, open := <-ch

	assert.ElementsMatch(t, expectedProfiles, actualModel.FirewallProfiles, "exact model must be provided")
	assert.NoError(t, actualModel.Error)
	assert.False(t, open, "channel must be closed at the end")
}

func Test_Provide_ProvidesErrorfullFirewallContainerOnFailingCommandWithPartialResult(t *testing.T) {
	expectedProfiles := []Profile{
		{Name: "Domain", Enabled: 1},
		{Name: "Public", Enabled: 0},
	}

	sut := provider{
		getRegistryValue: getMockRegistryValueFunc(map[string]uint64{
			"DomainProfile": 1,
			"PublicProfile": 0,
		}),
	}

	ch := sut.Provide()
	actualModel := <-ch
	_, open := <-ch

	assert.ElementsMatch(t, expectedProfiles, actualModel.FirewallProfiles, "partial profiles must be provided")
	assert.ErrorContains(t, actualModel.Error, "failed to get profile 'StandardProfile'", "error must be provided in result")
	assert.False(t, open, "channel must be closed at the end")
}

func getMockRegistryValueFunc(results map[string]uint64) registry2.GetKeyUIntValueTypeFunc {
	return func(_ registry2.RootKey, _, keyName string, _ string) (uint64, error) {
		result, exists := results[keyName]
		if !exists {
			return 0, fmt.Errorf("test failure getting %s", keyName)
		}
		return result, nil
	}
}
