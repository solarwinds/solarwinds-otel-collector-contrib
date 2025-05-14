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

package internal

import "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"

const (
	ArtificialMetricName = "swo.artifial.metric"
)

type ArtificialMetricEmitter struct {
	Emitter types.MetricsEmittingFunc
}

func (emitter *ArtificialMetricEmitter) Initialize() error {
	return nil
}

func (emitter *ArtificialMetricEmitter) GetEmittingFunction() types.MetricsEmittingFunc {
	return emitter.Emitter
}
