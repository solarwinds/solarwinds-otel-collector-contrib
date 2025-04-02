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

import (
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func CreateCustomScopeEmitter(
	name string,
	mes map[string]metric.Emitter,
) Emitter {
	return CreateDefaultScopeEmitter(name, mes)
}

type emitterMock struct {
	name string
}

var _ Emitter = (*emitterMock)(nil)

// Emit implements ScopeEmitter.
func (s *emitterMock) Emit() *Result {
	return &Result{
		Data:  pmetric.ScopeMetricsSlice{},
		Error: nil,
	}
}

// Init implements ScopeEmitter.
func (s *emitterMock) Init() error {
	return nil
}

// Name implements ScopeEmitter.
func (s *emitterMock) Name() string {
	return s.name
}
