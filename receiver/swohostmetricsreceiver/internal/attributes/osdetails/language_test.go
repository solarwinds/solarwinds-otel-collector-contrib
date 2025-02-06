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

package osdetails

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/stretchr/testify/assert"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/language"
)

func Test_LanguageAttributesGenerator_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := CreateLanguageAttributesGenerator(
		language.CreateLanguageProvider(),
	)

	result := <-sut.Generate()

	fmt.Printf("Result %+v\n", result)
}

func Test_Generate_LanguageIsProvided_AttributesAreGenerated(t *testing.T) {
	providedLanguage := language.Language{
		LCID:        123,
		Name:        "expected-Name",
		DisplayName: "expected-DisplayName",
	}

	expectedAttributes := shared.Attributes{
		"osdetails.language.lcid":        strconv.Itoa(providedLanguage.LCID),
		"osdetails.language.displayname": providedLanguage.DisplayName,
		"osdetails.language.name":        providedLanguage.Name,
	}

	sut := CreateLanguageAttributesGenerator(
		shared.CreateProviderMock(providedLanguage), // send valid data
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(
		t,
		expectedAttributes,
		actualAttributes,
		"expected attributes are not the same as actual",
	)
}

func Test_Generate_LanguageDataUnavailable_AttributesAreNotGenerated(t *testing.T) {
	sut := CreateLanguageAttributesGenerator(
		shared.CreateEmptyProviderMock[language.Language](), // not sending any data just closing the channel
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(
		t,
		shared.Attributes(nil),
		actualAttributes,
		"expected attributes are not the same as actual",
	)
}
