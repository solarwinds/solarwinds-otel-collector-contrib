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

package firewall

import (
	"fmt"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers/firewall"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type firewallProviderMock struct {
	fc firewall.Container
}

var _ providers.Provider[firewall.Container] = (*firewallProviderMock)(nil)

func CreateFirewallProviderMock(
	fc firewall.Container,
) providers.Provider[firewall.Container] {
	return &firewallProviderMock{
		fc: fc,
	}
}

// Provide implements providers.Provider.
func (m *firewallProviderMock) Provide() <-chan firewall.Container {
	ch := make(chan firewall.Container)
	go func() {
		defer close(ch)
		ch <- m.fc
	}()
	return ch
}

func produceFirewallProfileFromDataPoint(
	t *testing.T,
	dp pmetric.NumberDataPoint,
) firewall.Profile {
	v, b := dp.Attributes().Get(akFirewallProfile)
	assert.True(t, b, fmt.Sprintf("attribute %s must be available", akFirewallProfile))
	return firewall.Profile{Name: v.AsString(), Enabled: int(dp.IntValue())}
}
