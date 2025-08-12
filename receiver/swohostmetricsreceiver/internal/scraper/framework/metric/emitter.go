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

package metric

import "go.opentelemetry.io/collector/pdata/pmetric"

// Result is structure representing result from
// metric emitter.
type Result struct {
	// Data contains metric slice on success.
	Data pmetric.MetricSlice
	// Error is filled on failure, otherwise nil is returned.
	Error error
}

// Emitter is prescription for metric emitter in
// scraping framework.
type Emitter interface {
	// Init initializes metric emitter. Returns error
	// when failed, otherwise nil is returned.
	Init() error
	// Emit produces pointer to emitted metric result.
	Emit() *Result
	// Name returns name of metric emitted by metric emitter.
	Name() string
}

type InitFunc func() error

func (f InitFunc) Init() error {
	if f == nil {
		return nil
	}
	return f()
}
