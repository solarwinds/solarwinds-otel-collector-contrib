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

package types

import "go.opentelemetry.io/collector/pdata/pmetric"

// Functor capable of providing metric slice.
type MetricsEmittingFunc func() (pmetric.MetricSlice, error)

// Interface prescribing what metric emitter (actual metric producer)
// needs to implement.
type MetricEmitterInterface interface {
	// callback for initializing metrics internals.
	Initialize() error

	// emitter call back capable of providing metric slice.
	GetEmittingFunction() MetricsEmittingFunc
}
