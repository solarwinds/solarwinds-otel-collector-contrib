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
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/dnsqueryreceiver/internal/metadata"
)

// startTestDNSServer starts a UDP DNS server on 127.0.0.1:0 using the given handler.
// Returns the bound address ("host:port") and registers cleanup via t.Cleanup.
func startTestDNSServer(t *testing.T, handler dns.HandlerFunc) string {
	t.Helper()

	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &dns.Server{
		PacketConn: pc,
		Net:        "udp",
		Handler:    handler,
	}

	started := make(chan struct{})
	srv.NotifyStartedFunc = func() { close(started) }

	go func() { _ = srv.ActivateAndServe() }()

	t.Cleanup(func() { _ = srv.Shutdown() })

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("DNS test server did not start in time")
	}

	return pc.LocalAddr().String()
}

func noErrorHandler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Rcode = dns.RcodeSuccess
	_ = w.WriteMsg(m)
}

func nxdomainHandler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Rcode = dns.RcodeNameError
	_ = w.WriteMsg(m)
}

func scraperFromCfg(cfg *Config) *dnsQueryScraper {
	return newDNSQueryScraper(cfg, receivertest.NewNopSettings(metadata.Type))
}

// parseAddr splits "host:port" and returns host and int port.
func parseAddr(t *testing.T, addr string) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)
	return host, port
}

// datapointsForMetric collects all gauge data points for the named metric across all ResourceMetrics.
func datapointsForMetric(metrics pmetric.Metrics, name string) []pmetric.NumberDataPoint {
	var out []pmetric.NumberDataPoint
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				if m.Name() == name {
					if m.Type() != pmetric.MetricTypeGauge {
						continue
					}
					for l := 0; l < m.Gauge().DataPoints().Len(); l++ {
						out = append(out, m.Gauge().DataPoints().At(l))
					}
				}
			}
		}
	}
	return out
}

// findRMForMetric returns the first ResourceMetrics containing the named metric.
func findRMForMetric(t *testing.T, metrics pmetric.Metrics, name string) pmetric.ResourceMetrics {
	t.Helper()
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				if sm.Metrics().At(k).Name() == name {
					return rm
				}
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return pmetric.NewResourceMetrics()
}

func TestScrape_Success(t *testing.T) {
	addr := startTestDNSServer(t, noErrorHandler)
	host, port := parseAddr(t, addr)

	cfg := defaultTestConfig()
	cfg.Servers = []string{host}
	cfg.Port = port
	cfg.Domains = []string{"."}
	cfg.RecordType = "NS"

	metrics, err := scraperFromCfg(cfg).ScrapeMetrics(context.Background())
	require.NoError(t, err)

	// 1 (server, domain) pair → 1 ResourceMetrics containing 3 metrics
	require.Equal(t, 1, metrics.ResourceMetrics().Len())

	rcDPs := datapointsForMetric(metrics, "dns_query.result_code")
	require.Len(t, rcDPs, 1)
	require.Equal(t, int64(0), rcDPs[0].IntValue())
	requireAttrStr(t, rcDPs[0].Attributes(), "dns_query.rcode", "NOERROR")
	requireAttrStr(t, rcDPs[0].Attributes(), "dns_query.result", "success")

	rvalDPs := datapointsForMetric(metrics, "dns_query.rcode_value")
	require.Len(t, rvalDPs, 1)
	require.Equal(t, int64(0), rvalDPs[0].IntValue())

	timeDPs := datapointsForMetric(metrics, "dns_query.query_time_ms")
	require.Len(t, timeDPs, 1)
	require.GreaterOrEqual(t, timeDPs[0].DoubleValue(), 0.0)

	rm := findRMForMetric(t, metrics, "dns_query.result_code")
	requireAttrStr(t, rm.Resource().Attributes(), "sw.otelcol.dnsquery.server", host)
	requireAttrStr(t, rm.Resource().Attributes(), "sw.otelcol.dnsquery.domain", ".")
	requireAttrStr(t, rm.Resource().Attributes(), "sw.otelcol.dnsquery.record_type", "NS")
}

func TestScrape_NXDOMAIN(t *testing.T) {
	addr := startTestDNSServer(t, nxdomainHandler)
	host, port := parseAddr(t, addr)

	cfg := defaultTestConfig()
	cfg.Servers = []string{host}
	cfg.Port = port
	cfg.Domains = []string{"."}
	cfg.RecordType = "NS"

	metrics, err := scraperFromCfg(cfg).ScrapeMetrics(context.Background())
	require.NoError(t, err)

	rcDPs := datapointsForMetric(metrics, "dns_query.result_code")
	require.Len(t, rcDPs, 1)
	// NXDOMAIN is a valid DNS response — still result_code=0.
	require.Equal(t, int64(0), rcDPs[0].IntValue())
	requireAttrStr(t, rcDPs[0].Attributes(), "dns_query.rcode", "NXDOMAIN")

	rvalDPs := datapointsForMetric(metrics, "dns_query.rcode_value")
	require.Len(t, rvalDPs, 1)
	require.Equal(t, int64(3), rvalDPs[0].IntValue())
}

func TestScrape_Timeout(t *testing.T) {
	// Open a UDP listener that never reads or responds — causes client timeout.
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = pc.Close() })

	host, port := parseAddr(t, pc.LocalAddr().String())

	cfg := defaultTestConfig()
	cfg.Servers = []string{host}
	cfg.Port = port
	cfg.Domains = []string{"."}
	cfg.RecordType = "NS"
	cfg.Timeout = 100 * time.Millisecond

	metrics, err := scraperFromCfg(cfg).ScrapeMetrics(context.Background())
	require.NoError(t, err)

	rcDPs := datapointsForMetric(metrics, "dns_query.result_code")
	require.Len(t, rcDPs, 1)
	require.Equal(t, int64(1), rcDPs[0].IntValue(), "timeout should produce result_code=1")
	requireAttrStr(t, rcDPs[0].Attributes(), "dns_query.rcode", "")
	requireAttrStr(t, rcDPs[0].Attributes(), "dns_query.result", "timeout")

	timeDPs := datapointsForMetric(metrics, "dns_query.query_time_ms")
	require.Len(t, timeDPs, 1)
	require.Equal(t, 0.0, timeDPs[0].DoubleValue())

	rvalDPs := datapointsForMetric(metrics, "dns_query.rcode_value")
	require.Len(t, rvalDPs, 1)
	require.Equal(t, int64(0), rvalDPs[0].IntValue())
}

func TestScrape_Error(t *testing.T) {
	// Find a free UDP port by binding and immediately releasing it.
	// On Linux, an ICMP "port unreachable" is returned for unbound UDP ports,
	// which the DNS client surfaces as a connection error (result_code=2).
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := pc.LocalAddr().String()
	require.NoError(t, pc.Close())
	_, port := parseAddr(t, addr)

	cfg := defaultTestConfig()
	cfg.Servers = []string{"127.0.0.1"}
	cfg.Port = port
	cfg.Domains = []string{"."}
	cfg.RecordType = "NS"
	cfg.Timeout = 500 * time.Millisecond

	metrics, err := scraperFromCfg(cfg).ScrapeMetrics(context.Background())
	require.NoError(t, err)

	rcDPs := datapointsForMetric(metrics, "dns_query.result_code")
	require.Len(t, rcDPs, 1)
	require.Equal(t, int64(2), rcDPs[0].IntValue(), "connection error should produce result_code=2")
	requireAttrStr(t, rcDPs[0].Attributes(), "dns_query.result", "error")

	timeDPs := datapointsForMetric(metrics, "dns_query.query_time_ms")
	require.Len(t, timeDPs, 1)
	require.Equal(t, 0.0, timeDPs[0].DoubleValue())
}

func TestScrape_ConcurrentPairs(t *testing.T) {
	addr := startTestDNSServer(t, noErrorHandler)
	host, port := parseAddr(t, addr)

	cfg := defaultTestConfig()
	cfg.Servers = []string{host, host}
	cfg.Port = port
	cfg.Domains = []string{".", "example.com."}
	cfg.RecordType = "NS"

	metrics, err := scraperFromCfg(cfg).ScrapeMetrics(context.Background())
	require.NoError(t, err)

	// 2 servers × 2 domains = 4 (server, domain) pairs → 4 ResourceMetrics
	require.Equal(t, 4, metrics.ResourceMetrics().Len())

	type pair struct{ server, domain string }
	seen := make(map[pair]bool)
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		attrs := metrics.ResourceMetrics().At(i).Resource().Attributes()
		srv, sOk := attrs.Get("sw.otelcol.dnsquery.server")
		dom, dOk := attrs.Get("sw.otelcol.dnsquery.domain")
		if sOk && dOk {
			seen[pair{srv.Str(), dom.Str()}] = true
		}
	}

	for _, dom := range []string{".", "example.com."} {
		p := pair{host, dom}
		require.True(t, seen[p], "expected pair %+v in resource attributes", p)
	}
}

func TestLifecycle_Shutdown(t *testing.T) {
	addr := startTestDNSServer(t, noErrorHandler)
	host, port := parseAddr(t, addr)

	cfg := defaultTestConfig()
	cfg.Servers = []string{host}
	cfg.Port = port

	recv, err := NewFactory().CreateMetrics(
		context.Background(),
		receivertest.NewNopSettings(metadata.Type),
		cfg,
		consumertest.NewNop(),
	)
	require.NoError(t, err)
	require.NoError(t, recv.Start(context.Background(), newMdatagenNopHost()))
	require.NoError(t, recv.Shutdown(context.Background()))
}

// requireAttrStr is a package-local helper that reads a string attribute from a pcommon.Map.
func requireAttrStr(t *testing.T, attrs pcommon.Map, key, want string) {
	t.Helper()
	v, ok := attrs.Get(key)
	require.True(t, ok, "attribute %q not present", key)
	require.Equal(t, want, v.Str(), fmt.Sprintf("attribute %q", key))
}
