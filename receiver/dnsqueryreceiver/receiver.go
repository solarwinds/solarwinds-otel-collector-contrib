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

// Package dnsqueryreceiver scrapes DNS query metrics by issuing live DNS
// queries against configured servers and recording the result code, response
// code, and round-trip time as OpenTelemetry gauge metrics.
package dnsqueryreceiver

import "go.opentelemetry.io/collector/scraper"

// compile-time assertion that dnsQueryScraper satisfies the scraper.Metrics interface.
var _ scraper.Metrics = (*dnsQueryScraper)(nil)
