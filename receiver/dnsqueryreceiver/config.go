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
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

// Config defines the configuration for the dnsqueryreceiver.
type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`

	// Servers is the list of DNS servers to query. Required.
	Servers []string `mapstructure:"servers"`
	// Domains is the list of domains to query. Default: ["."]
	Domains []string `mapstructure:"domains"`
	// RecordType is the DNS record type to query. Default: "NS"
	RecordType string `mapstructure:"record_type"`
	// Network is the network protocol to use: "udp" or "tcp". Default: "udp"
	Network string `mapstructure:"network"`
	// Port is the DNS server port. Default: 53
	Port int `mapstructure:"port"`
	// Timeout is the query timeout. Default: 2s
	Timeout time.Duration `mapstructure:"timeout"`
}

var allowedRecordTypes = map[string]bool{
	"A": true, "AAAA": true, "ANY": true, "CNAME": true, "MX": true,
	"NS": true, "PTR": true, "SOA": true, "SPF": true, "SRV": true, "TXT": true,
}

var allowedNetworks = map[string]bool{
	"udp": true, "tcp": true,
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if len(c.Servers) == 0 {
		return errors.New("servers must not be empty")
	}
	for _, s := range c.Servers {
		if strings.TrimSpace(s) == "" {
			return errors.New("server entries must not be empty")
		}
	}
	if !allowedRecordTypes[c.RecordType] {
		return fmt.Errorf("record_type %q is not valid; allowed values: A, AAAA, ANY, CNAME, MX, NS, PTR, SOA, SPF, SRV, TXT", c.RecordType)
	}
	if !allowedNetworks[c.Network] {
		return fmt.Errorf("network %q is not valid; allowed values: udp, tcp", c.Network)
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port %d is not valid; must be between 1 and 65535", c.Port)
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	return nil
}
