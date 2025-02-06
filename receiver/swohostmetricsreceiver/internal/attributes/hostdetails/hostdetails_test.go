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

package hostdetails

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/domain"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/model"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/timezone"
	"github.com/stretchr/testify/assert"
)

func Test_HostDetailsAttributesGenerator_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := CreateHostDetailsAttributesGenerator(
		CreateDomainAttributesGenerator(
			domain.CreateDomainProvider(),
		),
		CreateModelAttributesGenerator(
			model.CreateModelProvider(),
		),
		CreateTimeZoneAttributesGenerator(
			timezone.CreateTimeZoneProvider(),
		),
	)

	result := <-sut.Generate()

	fmt.Printf("Result %+v\n", result)
}

func Test_Generate_AllGeneratorsGenerates_GeneratesFullAttributesCollection(t *testing.T) {
	domainAttributes := shared.Attributes{
		"domain.mock.empty": "",
		"domain.mock.full":  "any",
	}
	modelAttributes := shared.Attributes{
		"model.mock.empty": "",
		"model.mock.full":  "any",
	}
	timezoneAttributes := shared.Attributes{
		"timezone.mock.empty": "",
		"timezone.mock.full":  "any",
	}

	expectedAttributes := shared.Attributes{
		"domain.mock.empty":   "",
		"domain.mock.full":    "any",
		"model.mock.empty":    "",
		"model.mock.full":     "any",
		"timezone.mock.empty": "",
		"timezone.mock.full":  "any",
	}

	sut := CreateHostDetailsAttributesGenerator(
		shared.CreateAttributesGeneratorMock(domainAttributes),
		shared.CreateAttributesGeneratorMock(modelAttributes),
		shared.CreateAttributesGeneratorMock(timezoneAttributes),
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, expectedAttributes, actualAttributes)
}

func Test_Generate_SomeGeneratorsFails_GeneratesOnlyPartialCollection(t *testing.T) {
	domainAttributes := shared.Attributes{
		"domain.mock.empty": "",
		"domain.mock.full":  "any",
	}
	timezoneAttributes := shared.Attributes{
		"timezone.mock.empty": "",
		"timezone.mock.full":  "any",
	}

	expectedAttributes := shared.Attributes{
		"domain.mock.empty":   "",
		"domain.mock.full":    "any",
		"timezone.mock.empty": "",
		"timezone.mock.full":  "any",
	}

	sut := CreateHostDetailsAttributesGenerator(
		shared.CreateAttributesGeneratorMock(domainAttributes),
		shared.CreateAttributesGeneratorMock(nil), // model attributes
		shared.CreateAttributesGeneratorMock(timezoneAttributes),
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, expectedAttributes, actualAttributes)
}

func Test_Generate_AllGeneratorsFails_GeneratesEmptyCollection(t *testing.T) {
	sut := CreateHostDetailsAttributesGenerator(
		shared.CreateAttributesGeneratorMock(nil),
		shared.CreateAttributesGeneratorMock(nil),
		shared.CreateAttributesGeneratorMock(nil),
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, shared.Attributes{}, actualAttributes)
}
