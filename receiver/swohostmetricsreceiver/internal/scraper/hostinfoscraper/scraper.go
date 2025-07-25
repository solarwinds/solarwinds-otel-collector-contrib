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

package hostinfoscraper

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/metric"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scope"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/framework/scraper"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper/metrics/cpustats"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper/metrics/firewall"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper/metrics/uptime"
	lastloggeduser "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/scraper/hostinfoscraper/metrics/user/lastlogged"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/types"
	"go.uber.org/zap"
)

const (
	hostinfoScopeName = "otelcol/swohostmetricsreceiver/hostinfo"
	userScopeName     = "otelcol/swohostmetricsreceiver/hostinfo/user"
	cpuStatsScopeName = "otelcol/swohostmetricsreceiver/cpustats"
)

type Scraper struct {
	scraper.Manager
	config *types.ScraperConfig
}

var _ scraper.Scraper = (*Scraper)(nil)

func NewHostInfoScraper(
	scraperConfig *types.ScraperConfig,
	logger *zap.Logger,
) (*Scraper, error) {
	descriptor := &scraper.Descriptor{
		Type: ScraperType(),
		ScopeDescriptors: map[string]scope.Descriptor{
			hostinfoScopeName: {
				ScopeName: hostinfoScopeName,
				MetricDescriptors: map[string]metric.Descriptor{
					firewall.MetricName: {
						Create: firewall.NewEmitter,
						Logger: logger,
					},
					uptime.MetricName: {
						Create: uptime.NewEmitter,
						Logger: logger,
					},
				},
			},
			userScopeName: {
				ScopeName: userScopeName,
				MetricDescriptors: map[string]metric.Descriptor{
					lastloggeduser.MetricName: {
						Create: lastloggeduser.NewEmitter,
						Logger: logger,
					},
				},
			},
			cpuStatsScopeName: {
				ScopeName: cpuStatsScopeName,
				MetricDescriptors: map[string]metric.Descriptor{
					cpustats.MetricNameCPUTime: {
						Create: cpustats.NewEmitter(cpustats.MetricNameCPUTime),
						Logger: logger,
					},
					cpustats.MetricNameProcs: {
						Create: cpustats.NewEmitter(cpustats.MetricNameProcs),
						Logger: logger,
					},
					cpustats.MetricNameCurrentProcs: {
						Create: cpustats.NewEmitter(cpustats.MetricNameCurrentProcs),
						Logger: logger,
					},
					cpustats.MetricNameIntr: {
						Create: cpustats.NewEmitter(cpustats.MetricNameIntr),
						Logger: logger,
					},
					cpustats.MetricNameCtxt: {
						Create: cpustats.NewEmitter(cpustats.MetricNameCtxt),
						Logger: logger,
					},
					cpustats.MetricNameNumCores: {
						Create: cpustats.NewEmitter(cpustats.MetricNameNumCores),
						Logger: logger,
					},
				},
			},
		},
	}

	managerConfig := &scraper.ManagerConfig{
		ScraperConfig: scraperConfig,
	}

	s := &Scraper{
		Manager: scraper.NewScraperManager(logger),
		config:  scraperConfig,
	}

	if err := s.Init(descriptor, managerConfig); err != nil {
		return nil, err
	}

	return s, nil
}
