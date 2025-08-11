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

package processescount

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := Create(
		CreateWrapper(),
	)

	result := <-sut.Provide()

	fmt.Printf("Result: %+v\n", result)
}

func Test_Provide_WhenSucceedsReturnsCountAndChannelIsClosedAfterDelivery(t *testing.T) {
	expected := int64(1701)

	sut := Create(CreateSucceedingWrapper(expected))

	ch := sut.Provide()
	actualUptime := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expected, actualUptime.Count)
	assert.Nil(t, actualUptime.Error)
	assert.False(t, open, "channel must be closed")
}

func Test_Provide_WhenFailsReturnsZeroCountWithErrorAndChannelIsClosedAfterDelivery(t *testing.T) {
	expectedError := fmt.Errorf("kokoha happened")

	sut := Create(
		CreateFailingUptimeWrapper(expectedError),
	)

	ch := sut.Provide()
	actualCount := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedError, actualCount.Error)
	assert.Zero(t, actualCount.Count)
	assert.False(t, open, "channel must be closed")
}
