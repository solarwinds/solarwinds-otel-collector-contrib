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

package scope

import "github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/testutil"

type EmitterMock struct {
	emitResult *Result
	initResult error
	name       string
	EmitCC     testutil.CallsCounter
	InitCC     testutil.CallsCounter
	NameCC     testutil.CallsCounter
}

var _ Emitter = (*EmitterMock)(nil)

func CreateEmitterMock(
	emitResult *Result,
	initResult error,
	name string,
) *EmitterMock {
	return &EmitterMock{
		emitResult: emitResult,
		initResult: initResult,
		name:       name,
		EmitCC:     testutil.CallsCounter{},
		InitCC:     testutil.CallsCounter{},
		NameCC:     testutil.CallsCounter{},
	}
}

// Emit implements Emitter.
func (e *EmitterMock) Emit() *Result {
	e.EmitCC.IncrementCount()
	return e.emitResult
}

// Init implements Emitter.
func (e *EmitterMock) Init() error {
	e.InitCC.IncrementCount()
	return e.initResult
}

// Name implements Emitter.
func (e *EmitterMock) Name() string {
	e.NameCC.IncrementCount()
	return e.name
}
