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
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// Scraper represents general prescription for scraping.
// It mimics functions used in OTEL collector in scraper helper.
// Start function mimics component.StartFunc by signature.
// Shutdown function mimics component.ShutdownFunc.
// Scraper function mimics scraperhelper.ScrapeFunc.
type Scraper interface {
	Start(context.Context, component.Host) error
	Shutdown(context.Context) error
	Scrape(context.Context) (pmetric.Metrics, error)
}
