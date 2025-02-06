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

package shared

import (
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers"
)

type generatingAttributeGenerator struct {
	Payload Attributes
}

var _ AttributesGenerator = (*generatingAttributeGenerator)(nil)

func CreateAttributesGeneratorMock(
	payload Attributes,
) AttributesGenerator {
	return &generatingAttributeGenerator{
		Payload: payload,
	}
}

// Generate implements shared.AttributesGenerator.
func (g *generatingAttributeGenerator) Generate() AttributesChannel {
	ch := make(AttributesChannel)
	go func() {
		ch <- g.Payload
		close(ch)
	}()
	return ch
}

type providerMock[T any] struct {
	providedObject T
}

func CreateProviderMock[T any](d T) providers.Provider[T] {
	return &providerMock[T]{
		providedObject: d,
	}
}

// Provide implements domain.DomainProvider.
func (m *providerMock[T]) Provide() <-chan T {
	ch := make(chan T)
	go func() {
		ch <- m.providedObject
		close(ch)
	}()

	return ch
}

type emptyProviderMock[T any] struct{}

func CreateEmptyProviderMock[T any]() providers.Provider[T] {
	return &emptyProviderMock[T]{}
}

// Provide implements domain.DomainProvider.
func (*emptyProviderMock[T]) Provide() <-chan T {
	ch := make(chan T)
	defer close(ch)
	return ch
}
