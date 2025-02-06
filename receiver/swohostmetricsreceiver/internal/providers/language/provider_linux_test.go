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

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/stretchr/testify/assert"
)

func Test_Provide_ProvidesLanguageNameAndChannelIsClosedAfterDelivery(t *testing.T) {
	commandOutput := "LANGUAGE=en_US"

	expectedLanguage := Language{
		Name: "en_US",
	}

	sut := provider{
		cli: providers.CreateCommandLineExecutorMock(commandOutput, "", nil),
	}

	ch := sut.Provide()
	actualLanguage := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedLanguage, actualLanguage)
	assert.False(t, open, "channel must be closed")
}

func Test_Provide_ProvidesEmptyLanguageNameWhenLocaleNotSetAndChannelIsClosedAfterDelivery(t *testing.T) {
	commandOutput := "LANGUAGE="

	expectedLanguage := Language{}

	sut := provider{
		cli: providers.CreateCommandLineExecutorMock(commandOutput, "", nil),
	}

	ch := sut.Provide()
	actualLanguage := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedLanguage, actualLanguage)
	assert.False(t, open, "channel must be closed")
}

func Test_Provide_FailsAndProvidesEmptyObjectAndChannelIsClosedAfterDelivery(t *testing.T) {
	expectedLanguage := Language{}

	sut := provider{
		cli: providers.CreateCommandLineExecutorMock("", "", fmt.Errorf("kokoha error")),
	}

	ch := sut.Provide()
	actualLanguage := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedLanguage, actualLanguage)
	assert.False(t, open, "channel must be closed")
}

func Test_Provide_ErrorInStdErrItProvidesEmptyObjectAndChannelIsClosedAfterDelivery(t *testing.T) {
	expectedLanguage := Language{}

	sut := provider{
		cli: providers.CreateCommandLineExecutorMock("", "kokoha error", nil),
	}

	ch := sut.Provide()
	actualLanguage := <-ch
	_, open := <-ch // secondary receive

	assert.Equal(t, expectedLanguage, actualLanguage)
	assert.False(t, open, "channel must be closed")
}
