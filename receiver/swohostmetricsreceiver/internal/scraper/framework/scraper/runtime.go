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

package scraper

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
)

// Runtime represents current set (map) of initialized and
// allocated scope emitters for usage in scraper manager.
type Runtime struct {
	// ScopeEmitters represents allocated and required
	// scope emitters for given scraper.
	ScopeEmitters map[string]scope.Emitter
}

func createScraperRuntime(
	scraperDescriptor *Descriptor,
	enabledMetrics *metric.Enabled,
	logger *zap.Logger,
) (*Runtime, error) {
	// Traverse and assembly scope descriptors.
	ses := scope.TraverseThroughScopeDescriptors(
		scraperDescriptor.ScopeDescriptors,
		enabledMetrics,
		logger,
	)

	// No schedule scope emitters.
	if len(ses) == 0 {
		return nil, fmt.Errorf("no scheduled scope emitters for scraper '%s'", scraperDescriptor.Type)
	}

	sr := new(Runtime)
	sr.ScopeEmitters = ses
	return sr, nil
}
