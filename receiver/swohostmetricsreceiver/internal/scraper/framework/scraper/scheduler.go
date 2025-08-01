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
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
)

// Schedule prescribes how transformation of
// configuration oriented structure to runtime
// configuration should look like.
type Scheduler interface {
	// Schedule method process the transformation mentioned above.
	// Provider descriptor pointer together with scraper configuration
	// is "translated" into Runtime object. When fails error is returned,
	// otherwise nil is return in place of error.
	Schedule(
		*Descriptor,
		*types.ScraperConfig,
		*zap.Logger,
	) (*Runtime, error)
}

type scheduler struct{}

var _ Scheduler = (*scheduler)(nil)

func NewScraperScheduler() Scheduler {
	return new(scheduler)
}

// Schedule implements ScraperScheduler.
func (*scheduler) Schedule(
	descriptor *Descriptor,
	config *types.ScraperConfig,
	logger *zap.Logger,
) (*Runtime, error) {
	sn := descriptor.Type

	// Obtains enabled metrics for scheduled scraper.
	enabledMetrics, err := metric.GetEnabledMetrics(sn.String(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled metrics for scraper '%s': %w", sn, err)
	}

	// Assembly Scraper runtime based on enabled metrics.
	scraperRuntime, err := createScraperRuntime(descriptor, enabledMetrics, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create scraper runtime for scraper '%s': %w", sn, err)
	}

	logger.Debug("scheduling of scraper finished successfully", zap.String("scraper", sn.String()))
	return scraperRuntime, nil
}
