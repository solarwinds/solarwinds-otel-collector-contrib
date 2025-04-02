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
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers/infostat"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers/language"
	"github.com/stretchr/testify/assert"
)

func Test_OsDetailsAttributesGenerator_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := CreateOsDetailsAttributesGenerator(
		CreateInfoStatAttributesGenerator(
			infostat.CreateInfoStatProvider(),
		),
		CreateLanguageAttributesGenerator(
			language.CreateLanguageProvider(),
		),
	)

	result := <-sut.Generate()

	fmt.Printf("Result %+v\n", result)
}

func Test_Generate_AllGeneratorsGenerates_GeneratesFullAttributesCollection(t *testing.T) {
	infoStatAttributes := shared.Attributes{
		"domain.mock.empty": "",
		"domain.mock.full":  "any",
	}

	languageAttributes := shared.Attributes{
		"language.mock.empty": "",
		"language.mock.full":  "any",
	}

	expectedAttributes := shared.Attributes{
		"domain.mock.empty":   "",
		"domain.mock.full":    "any",
		"language.mock.empty": "",
		"language.mock.full":  "any",
	}

	sut := CreateOsDetailsAttributesGenerator(
		shared.CreateAttributesGeneratorMock(infoStatAttributes),
		shared.CreateAttributesGeneratorMock(languageAttributes),
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, expectedAttributes, actualAttributes)
}

func Test_Generate_SomeGeneratorsFails_GeneratesOnlyPartialCollection(t *testing.T) {
	infoStatAttributes := shared.Attributes{
		"infostat.mock.empty": "",
		"infostat.mock.full":  "any",
	}

	expectedAttributes := shared.Attributes{
		"infostat.mock.empty": "",
		"infostat.mock.full":  "any",
	}

	sut := CreateOsDetailsAttributesGenerator(
		shared.CreateAttributesGeneratorMock(infoStatAttributes),
		shared.CreateAttributesGeneratorMock(shared.Attributes(nil)), // language attributes
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, expectedAttributes, actualAttributes)
}

func Test_Generate_AllGeneratorsFails_GeneratesEmptyCollection(t *testing.T) {
	sut := CreateOsDetailsAttributesGenerator(
		shared.CreateAttributesGeneratorMock(shared.Attributes(nil)),
		shared.CreateAttributesGeneratorMock(shared.Attributes(nil)), // language attributes
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, shared.Attributes{}, actualAttributes)
}
