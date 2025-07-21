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
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/registry"

	"github.com/stretchr/testify/assert"
)

func Test_Provide_RetrievingLastLoggedUserProvidesDataAndChannelIsClosedAfterDelivery(t *testing.T) {
	registryValues := map[string]string{
		`LastLoggedOnUser`:        `TestDomain/TestUser`,
		`LastLoggedOnDisplayName`: `User Test`,
	}

	expectedResult := Data{
		Users: []User{{
			Name:        `TestDomain/TestUser`,
			DisplayName: `User Test`,
		}},
	}

	sut := provider{
		getRegistryValues: func(_ registry.RootKey, _, _ string, _ []string) (map[string]string, error) {
			return registryValues, nil
		},
	}

	ch := sut.Provide()
	actualResult := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedResult, actualResult)
	assert.False(t, open, `channel must be closed`)
}

func Test_Provide_RetrievingLastLoggedUserFailsProvidesErrorAndChannelIsClosedAfterDelivery(t *testing.T) {
	registryError := fmt.Errorf(`cardinal mistake`)
	expectedResult := Data{Error: registryError}

	sut := provider{
		getRegistryValues: func(_ registry.RootKey, _, _ string, _ []string) (map[string]string, error) {
			return nil, registryError
		},
	}

	ch := sut.Provide()
	actualResult := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedResult, actualResult)
	assert.False(t, open, `channel must be closed`)
}
