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

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers/model"
	"github.com/stretchr/testify/assert"
)

func Test_ModelAttributesGenerator_Functional(t *testing.T) {
	t.Skip("This test should be run only manually")

	sut := CreateModelAttributesGenerator(
		model.CreateModelProvider(),
	)

	result := <-sut.Generate()

	fmt.Printf("Result: %+v\n", result)
}

func Test_Generate_OnFullModel_AllAttributesAreGenerated(t *testing.T) {
	modelName := "USS Enterprise"
	modelManufacturer := "Utopia Planitia Shipyards"
	modelSerialNumber := "NCC-1701E"

	providedModel := model.Model{
		SerialNumber: modelSerialNumber,
		Manufacturer: modelManufacturer,
		Name:         modelName,
	}

	expectedAttributes := shared.Attributes{
		"hostdetails.model.manufacturer": modelManufacturer,
		"hostdetails.model.name":         modelName,
		"hostdetails.model.serialnumber": modelSerialNumber,
	}

	sut := CreateModelAttributesGenerator(
		shared.CreateProviderMock(
			providedModel,
		),
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, expectedAttributes, actualAttributes)
}

func Test_Generate_OnPartialModel_SomeAttributesAreGenerated(t *testing.T) {
	modelName := "USS Enterprise"
	modelManufacturer := "Utopia Planitia Shipyards"
	modelSerialNumber := ""

	providedModel := model.Model{
		SerialNumber: modelSerialNumber,
		Manufacturer: modelManufacturer,
		Name:         modelName,
	}

	expectedAttributes := shared.Attributes{
		"hostdetails.model.manufacturer": modelManufacturer,
		"hostdetails.model.name":         modelName,
	}

	sut := CreateModelAttributesGenerator(
		shared.CreateProviderMock(
			providedModel,
		),
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(
		t,
		expectedAttributes,
		actualAttributes,
		"expected attributes are not the same as actual",
	)
}

func Test_Generate_OnNoModel_NilAttributesAreGenerated(t *testing.T) {
	sut := CreateModelAttributesGenerator(
		shared.CreateEmptyProviderMock[model.Model](),
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(t, shared.Attributes(nil), actualAttributes)
}
