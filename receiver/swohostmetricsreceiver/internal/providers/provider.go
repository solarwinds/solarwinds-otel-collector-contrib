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

package providers

type Data[T any] struct {
	Value T
	Error error
}

// Provider is a general interface for most of the data providers.
type Provider[T any] interface {
	// Provide returns a channel through which the data are provided.
	Provide() <-chan T
}

// ProviderFunc is a general interface for most of the inner data providers.
type ProviderFunc[T any] func() (T, error)

type provider[T any] struct {
	provide ProviderFunc[T]
}

var _ Provider[Data[any]] = (*provider[any])(nil)

func CreateDataProvider[T any](
	pf ProviderFunc[T],
) Provider[Data[T]] {
	return &provider[T]{provide: pf}
}

func (p *provider[T]) Provide() <-chan Data[T] {
	ch := make(chan Data[T])
	go p.provideInGoRoutine(ch)
	return ch
}

func (p *provider[T]) provideInGoRoutine(ch chan Data[T]) {
	defer close(ch)
	value, err := p.provide()
	ch <- Data[T]{value, err}
}
