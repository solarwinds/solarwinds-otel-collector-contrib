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

package installedsoftware

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/cli"
	"github.com/stretchr/testify/assert"
)

const dpkgPayload = `
ii  acpid                           1:2.0.34-1ubuntu2                       amd64        Advanced Configuration and Power Interface event daemon
ii  adduser                         3.137ubuntu1                            all          add and remove users and groups
ii  amd64-microcode                 3.20231019.1ubuntu2                     amd64        Processor microcode firmware for AMD CPUs
ii  apparmor                        4.0.0-beta3-0ubuntu3                    amd64        user-space parser utility for AppArmor
ii  apport                          2.28.1-0ubuntu2                         all          automatically generate crash reports for debugging
`

func Test_Dpkg_GetSoftware_CommandFailsErrorIsReturned(t *testing.T) {
	cli := cli.CreateNewCliExecutorMock("", "", fmt.Errorf("failed to process command"))
	sut := createDpkgProvider(cli)

	is, err := sut.GetSoftware()

	assert.Zero(t, len(is), "no installed software must be returned")
	assert.Error(t, err, "call must fail")
}

func Test_Dpkg_GetSoftware_CommandSucceedsSoftwareIsReturned(t *testing.T) {
	expectedSoftwareCount := 5

	cli := cli.CreateNewCliExecutorMock(dpkgPayload, "", nil)
	sut := createDpkgProvider(cli)

	is, err := sut.GetSoftware()

	assert.Equal(t, expectedSoftwareCount, len(is), "installed software must be in exact count")
	assert.NoError(t, err, "call must succeed")
	assert.Equal(t, is[0].Name, "acpid", "the first installed software must be equal")
	assert.Equal(t, is[4].Name, "apport", "the last installed software must be equal")
}

func Test_Dpkg_GetSoftware_UnsupportedFormatEmptyCollectionIsReturnedWithNoError(t *testing.T) {
	cli := cli.CreateNewCliExecutorMock("kokoha_unsupported", "", nil)
	sut := createDpkgProvider(cli)

	is, err := sut.GetSoftware()

	assert.Zero(t, len(is), "installed software collection must be empty")
	assert.NoError(t, err, "no error is returned")
}
