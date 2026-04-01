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

package dnsqueryreceiver

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/dnsqueryreceiver/internal/metadata"
)

// dnsQueryResult holds the result of a single DNS query for one (server, domain) pair.
type dnsQueryResult struct {
	server      string
	domain      string
	queryTimeMs float64
	resultCode  int64
	rcodeValue  int64
	resultStr   string
	rcodeStr    string
}

type dnsQueryScraper struct {
	config   *Config
	logger   *zap.Logger
	mb       *metadata.MetricsBuilder
}

// settings is not stored as a struct field — all needed values (logger, mb) are extracted at construction time.
func newDNSQueryScraper(cfg *Config, settings receiver.Settings) *dnsQueryScraper {
	return &dnsQueryScraper{
		config: cfg,
		logger: settings.Logger,
		mb:     metadata.NewMetricsBuilder(metadata.DefaultMetricsBuilderConfig(), settings),
	}
}

func (s *dnsQueryScraper) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (s *dnsQueryScraper) Shutdown(_ context.Context) error {
	return nil
}

func (s *dnsQueryScraper) ScrapeMetrics(ctx context.Context) (pmetric.Metrics, error) {
	type serverDomainPair struct{ server, domain string }

	pairs := make([]serverDomainPair, 0, len(s.config.Servers)*len(s.config.Domains))
	for _, srv := range s.config.Servers {
		for _, dom := range s.config.Domains {
			pairs = append(pairs, serverDomainPair{srv, dom})
		}
	}

	results := make([]dnsQueryResult, len(pairs))
	var wg sync.WaitGroup
	for i, p := range pairs {
		wg.Add(1)
		go func(idx int, server, domain string) {
			defer wg.Done()
			results[idx] = s.queryDNS(ctx, server, domain)
		}(i, p.server, p.domain)
	}
	wg.Wait()

	out := pmetric.NewMetrics()
	now := pcommon.NewTimestampFromTime(time.Now())

	for _, r := range results {
		rb := s.mb.NewResourceBuilder()
		rb.SetSwOtelcolDnsqueryServer(r.server)
		rb.SetSwOtelcolDnsqueryDomain(r.domain)
		rb.SetSwOtelcolDnsqueryRecordType(s.config.RecordType)

		s.mb.RecordDNSQueryQueryTimeMsDataPoint(now, r.queryTimeMs,
			r.domain, r.rcodeStr, s.config.RecordType, r.resultStr, r.server)
		s.mb.RecordDNSQueryResultCodeDataPoint(now, r.resultCode,
			r.domain, r.rcodeStr, s.config.RecordType, r.resultStr, r.server)
		s.mb.RecordDNSQueryRcodeValueDataPoint(now, r.rcodeValue,
			r.domain, r.rcodeStr, s.config.RecordType, r.resultStr, r.server)

		rm := s.mb.Emit(metadata.WithResource(rb.Emit()))
		rm.ResourceMetrics().MoveAndAppendTo(out.ResourceMetrics())
	}

	return out, nil
}

func (s *dnsQueryScraper) queryDNS(ctx context.Context, server, domain string) dnsQueryResult {
	result := dnsQueryResult{
		server:    server,
		domain:    domain,
		resultStr: "error",
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dnsTypeFromString(s.config.RecordType))
	m.RecursionDesired = true

	c := &dns.Client{
		Net:     s.config.Network,
		Timeout: s.config.Timeout,
	}
	addr := net.JoinHostPort(server, strconv.Itoa(s.config.Port))

	resp, rtt, err := c.ExchangeContext(ctx, m, addr)
	if err != nil {
		if isTimeout(ctx, err) {
			result.resultCode = 1
			result.resultStr = "timeout"
		} else {
			result.resultCode = 2
			s.logger.Warn("DNS query failed",
				zap.String("server", server),
				zap.String("domain", domain),
				zap.Error(err),
			)
		}
		return result
	}

	result.resultCode = 0
	result.resultStr = "success"
	result.queryTimeMs = float64(rtt.Nanoseconds()) / 1e6
	result.rcodeValue = int64(resp.Rcode)
	result.rcodeStr = dns.RcodeToString[resp.Rcode]
	return result
}

func dnsTypeFromString(recordType string) uint16 {
	if t, ok := dns.StringToType[recordType]; ok {
		return t
	}
	return dns.TypeNS
}

func isTimeout(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return true
	}
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return false
}
