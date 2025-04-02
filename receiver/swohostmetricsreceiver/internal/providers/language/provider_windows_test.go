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

package language

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/registry"

	"github.com/stretchr/testify/assert"
)

func Test_Provide_ProvidesFullLanguageAndChannelIsClosedAfterDelivery(t *testing.T) {
	expectedLanguage := Language{
		LCID:        999,
		Name:        "test-short",
		DisplayName: "Test Display Name",
	}

	sut := provider{
		lcidProvider:    &mockLCIDProvider{lcid: 999},
		displayNamesMap: map[string]string{"test-short": "Test Display Name"},
		getRegistryValues: func(_ registry.RootKey, _, _ string, _ []string) (map[string]string, error) {
			return map[string]string{"LocaleName": "test-short"}, nil
		},
	}

	ch := sut.Provide()
	actualLanguage := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedLanguage, actualLanguage)
	assert.False(t, open, "channel must be closed")
}

func Test_Provide_ShortNameGettingFailsAndProvidesEmptyObjectAndChannelIsClosedAfterDelivery(t *testing.T) {
	expectedLanguage := Language{}

	sut := provider{
		lcidProvider:    &mockLCIDProvider{lcid: 999},
		displayNamesMap: map[string]string{},
		getRegistryValues: func(_ registry.RootKey, _, _ string, _ []string) (map[string]string, error) {
			return nil, fmt.Errorf("no short name for you")
		},
	}

	ch := sut.Provide()
	actualLanguage := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedLanguage, actualLanguage)
	assert.False(t, open, "channel must be closed")
}

// stubs.
type mockLCIDProvider struct {
	lcid uint32
}

func (m *mockLCIDProvider) GetUserDefaultLCID() uint32 {
	return m.lcid
}
