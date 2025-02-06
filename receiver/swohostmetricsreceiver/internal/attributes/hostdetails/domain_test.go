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
	"strconv"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/domain"
	"github.com/stretchr/testify/assert"
)

func Test_DomainAttributesGenerator_Functional(t *testing.T) {
	t.Skip("Only for manual run")

	sut := CreateDomainAttributesGenerator(
		domain.CreateDomainProvider(),
	)

	result := <-sut.Generate()

	fmt.Printf("Result %v\n", result)
}

func Test_Generate_DomainIsProvided_AttributesAreGenerated(t *testing.T) {
	expectedDomain := "kokoha"
	expectedFQDN := "kokoha.host"
	expectedDomainRole := 1
	expectedWorkgroup := "kokohagroup"

	providedDomain := domain.Domain{
		Domain:     expectedDomain,
		FQDN:       expectedFQDN,
		DomainRole: expectedDomainRole,
		Workgroup:  expectedWorkgroup,
	}

	expectedAttributes := shared.Attributes{
		"hostdetails.domain":           expectedDomain,
		"hostdetails.domain.fqdn":      expectedFQDN,
		"hostdetails.domain.role":      strconv.Itoa(expectedDomainRole),
		"hostdetails.domain.workgroup": expectedWorkgroup,
	}

	sut := CreateDomainAttributesGenerator(
		shared.CreateProviderMock(providedDomain), // send valid data
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(
		t,
		expectedAttributes,
		actualAttributes,
		"expected attributes are not the same as actual",
	)
}

func Test_Generate_DomainDataUnavailable_AttributesAreNotGenerated(t *testing.T) {
	sut := CreateDomainAttributesGenerator(
		shared.CreateEmptyProviderMock[domain.Domain](), // not sending any data just closing the channel
	)

	actualAttributes := <-sut.Generate()

	assert.Equal(
		t,
		shared.Attributes(nil),
		actualAttributes,
		"expected attributes are not the same as actual",
	)
}
