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
	"testing"

	"github.com/stretchr/testify/require"
)

func defaultTestConfig() *Config {
	return createDefaultConfig().(*Config)
}

func TestValidate_EmptyServers(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{}
	err := cfg.Validate()
	require.ErrorContains(t, err, "servers must not be empty")
}

func TestValidate_BlankServer(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{""}
	err := cfg.Validate()
	require.Error(t, err)
}

func TestValidate_InvalidRecordType(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{"x"}
	cfg.RecordType = "BOGUS"
	err := cfg.Validate()
	require.ErrorContains(t, err, "record_type")
}

func TestValidate_InvalidNetwork(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{"x"}
	cfg.Network = "quic"
	err := cfg.Validate()
	require.ErrorContains(t, err, "network")
}

func TestValidate_InvalidPort_Zero(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{"x"}
	cfg.Port = 0
	err := cfg.Validate()
	require.ErrorContains(t, err, "port")
}

func TestValidate_InvalidPort_Negative(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{"x"}
	cfg.Port = -1
	err := cfg.Validate()
	require.ErrorContains(t, err, "port")
}

func TestValidate_InvalidPort_TooHigh(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{"x"}
	cfg.Port = 99999
	err := cfg.Validate()
	require.ErrorContains(t, err, "port")
}

func TestValidate_InvalidTimeout(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{"x"}
	cfg.Timeout = 0
	err := cfg.Validate()
	require.ErrorContains(t, err, "timeout")
}

func TestValidate_ValidMinimal(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.Servers = []string{"8.8.8.8"}
	err := cfg.Validate()
	require.NoError(t, err)
}

func TestValidate_AllRecordTypes(t *testing.T) {
	recordTypes := []string{"A", "AAAA", "ANY", "CNAME", "MX", "NS", "PTR", "SOA", "SPF", "SRV", "TXT"}
	for _, rt := range recordTypes {
		t.Run(rt, func(t *testing.T) {
			cfg := defaultTestConfig()
			cfg.Servers = []string{"8.8.8.8"}
			cfg.RecordType = rt
			err := cfg.Validate()
			require.NoError(t, err)
		})
	}
}
