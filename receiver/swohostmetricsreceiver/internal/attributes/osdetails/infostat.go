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

package osdetails

import (
	"strconv"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/infostat"
)

// Following keys will be used in metric as attributes.
const (
	hostname             = "osdetails.hostname"
	bootTime             = "osdetails.boottime"
	os                   = "osdetails.os"
	platform             = "osdetails.platform"
	platformFamily       = "osdetails.platform.family"
	platformVersion      = "osdetails.platform.version"
	kernelVersion        = "osdetails.kernel.version"
	kernelArchitecture   = "osdetails.kernel.architecture"
	virtualizationSystem = "osdetails.virtualization.system"
	virtualizationRole   = "osdetails.virtualization.role"
	hostID               = "osdetails.host.id"
	infoStatCount        = 11
)

type infoStatAttributesGenerator struct {
	InfoStatProvider providers.Provider[infostat.InfoStat]
}

var _ shared.AttributesGenerator = (*infoStatAttributesGenerator)(nil)

func CreateInfoStatAttributesGenerator(isp providers.Provider[infostat.InfoStat]) shared.AttributesGenerator {
	return &infoStatAttributesGenerator{
		InfoStatProvider: isp,
	}
}

// Generate implements shared.AttributesGenerator.
func (g *infoStatAttributesGenerator) Generate() shared.AttributesChannel {
	ch := make(shared.AttributesChannel)
	go g.generateInternal(ch)
	return ch
}

func (g *infoStatAttributesGenerator) generateInternal(ch shared.AttributesChannel) {
	defer close(ch)

	// activate provider and store its channel
	pCh := g.InfoStatProvider.Provide()

	// receive data and wait until provider's channel is done
	for t := range pCh {
		atts := processInfoStatAttributes(t)
		// when attribute are ready send it
		ch <- atts
	}
}

func processInfoStatAttributes(is infostat.InfoStat) shared.Attributes {
	m := make(shared.Attributes, infoStatCount)
	if len(is.Hostname) > 0 {
		m[hostname] = is.Hostname
	}
	if len(is.Os) > 0 {
		m[os] = is.Os
	}
	if len(is.Platform) > 0 {
		m[platform] = is.Platform
	}
	if len(is.PlatformFamily) > 0 {
		m[platformFamily] = is.PlatformFamily
	}
	if len(is.PlatformVersion) > 0 {
		m[platformVersion] = is.PlatformVersion
	}
	if len(is.KernelVersion) > 0 {
		m[kernelVersion] = is.KernelVersion
	}
	if len(is.KernelArchitecture) > 0 {
		m[kernelArchitecture] = is.KernelArchitecture
	}
	if len(is.VirtualizationSystem) > 0 {
		m[virtualizationSystem] = is.VirtualizationSystem
	}
	if len(is.VirtualizationRole) > 0 {
		m[virtualizationRole] = is.VirtualizationRole
	}
	if len(is.HostID) > 0 {
		m[hostID] = is.HostID
	}
	m[bootTime] = strconv.FormatUint(is.BootTime, 10)
	return m
}
