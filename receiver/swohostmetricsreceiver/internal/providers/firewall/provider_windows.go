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
	"errors"
	"fmt"
	"sync"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/registry"
	"go.uber.org/zap"
)

const (
	firewallProfileKey = `SYSTEM\CurrentControlSet\services\SharedAccess\Parameters\FirewallPolicy`
	enabledValueName   = "EnableFirewall"
)

type provider struct {
	getRegistryValue registry.GetKeyUIntValueTypeFunc
}

var _ (providers.Provider[Container]) = (*provider)(nil)

func CreateFirewallProvider() providers.Provider[Container] {
	return &provider{
		getRegistryValue: registry.GetKeyUIntValue,
	}
}

// Provide implements providers.Provider.
func (fp *provider) Provide() <-chan Container {
	ch := make(chan Container)
	go func() {
		defer close(ch)
		result, err := fp.getFirewallProfiles()
		if err != nil {
			zap.L().Error("providing firewall profiles failed", zap.Error(err))
		}
		fc := Container{
			FirewallProfiles: result,
			Error:            err,
		}
		ch <- fc
	}()
	return ch
}

func getFirewallProfilesMapping() map[string]string {
	return map[string]string{
		"StandardProfile": "Private",
		"DomainProfile":   "Domain",
		"PublicProfile":   "Public",
	}
}

// getFirewallProfiles fetches individual profiles states from registry.
// Mapping of final profile names is based upon Get-NetFirewallProfile PS
// cmdlet outputs.
func (fp *provider) getFirewallProfiles() ([]Profile, error) {
	profileKeys := getFirewallProfilesMapping()
	errCh := make(chan error, len(profileKeys))
	profilesCh := make(chan Profile, len(profileKeys))

	wg := sync.WaitGroup{}
	for profileKey, profileType := range profileKeys {
		pk := profileKey
		pt := profileType
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := fp.getRegistryValue(registry.LocalMachineKey, firewallProfileKey, pk, enabledValueName)

			if err == nil {
				profilesCh <- Profile{Name: pt, Enabled: int(value)}
			} else {
				errCh <- fmt.Errorf("failed to get profile '%s' value: %w", pk, err)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	close(profilesCh)

	profiles, err := processFirewallProfilesChannels(errCh, profilesCh)
	return profiles, err
}

func processFirewallProfilesChannels(errCh chan error, profilesCh chan Profile) ([]Profile, error) {
	var err error
	for e := range errCh {
		err = errors.Join(e, err)
	}

	var profiles []Profile
	for p := range profilesCh {
		profiles = append(profiles, p)
	}

	return profiles, err
}
