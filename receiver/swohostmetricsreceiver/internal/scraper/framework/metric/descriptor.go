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

import (
	"go.uber.org/zap"
)

// MetricEmitterCreateFunc is a functor for creation of
// metric emitter instance.
type EmitterCreateFunc func(*zap.Logger) Emitter

// MetricDescriptor represent description of one metric.
type Descriptor struct {
	// Creator function for creation of specific metric emitter for
	// metric represented by this descriptor.
	Create EmitterCreateFunc
}

func TraverseThroughMetricDescriptors(
	metricDescriptors map[string]Descriptor,
	enabledMetrics *Enabled,
	logger *zap.Logger,
) map[string]Emitter {
	mes := make(map[string]Emitter, 0)

	for mName, mDescriptor := range metricDescriptors {
		// Metric is not among enabled ones.
		if _, found := enabledMetrics.Metrics[mName]; !found {
			continue
		}

		// Metric is enabled by config, let's use it.
		logger.Sugar().Debugf("creating metric emitter for '%s", mName)
		me := mDescriptor.Create(logger)
		mes[me.Name()] = me
	}

	return mes
}
