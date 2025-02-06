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

package discovery

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/cli"
)

func Test_Rpm_Functional(t *testing.T) {
	t.Skip("This test should be ran manually only")

	sut := NewRpmDiscovery()
	result := sut.Discover()
	assert.True(t, result)
}

func Test_RpmDiscover_OnCommandSuccessReturnsTrue(t *testing.T) {
	cliMock := cli.CreateNewCliExecutorMock("<whatever_command_output>", "", nil)
	sut := createRpmDiscovery(cliMock)

	actual := sut.Discover()
	assert.True(t, actual, "discovery must be positive")
}

func Test_RpmDiscover_OnCommandFailureReturnsFalse(t *testing.T) {
	cliMock := cli.CreateNewCliExecutorMock("", "", fmt.Errorf("discovery command failure"))
	sut := createRpmDiscovery(cliMock)

	actual := sut.Discover()
	assert.False(t, actual, "discovery must be negatove")
}
