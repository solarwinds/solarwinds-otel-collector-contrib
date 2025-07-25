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

package infostat

import (
	"fmt"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"

	"github.com/shirou/gopsutil/v3/host"
	"go.uber.org/zap"
)

type InfoStat struct {
	Hostname             string
	BootTime             uint64
	Os                   string
	Platform             string
	PlatformFamily       string
	PlatformVersion      string
	KernelVersion        string
	KernelArchitecture   string
	VirtualizationSystem string
	VirtualizationRole   string
	HostID               string
}

type provider struct {
	internalExecutor infoStatExecutor
	logger           *zap.Logger
}

var _ providers.Provider[InfoStat] = (*provider)(nil)

func CreateInfoStatProvider(logger *zap.Logger) providers.Provider[InfoStat] {
	return &provider{
		internalExecutor: &executor{},
		logger:           logger,
	}
}

// Wrapper for host.Info() implementation.
type infoStatExecutor interface {
	Getinfo() (*host.InfoStat, error)
}
type executor struct{}

var _ infoStatExecutor = (*executor)(nil)

// Getinfo implements infoStatExecutor.
func (*executor) Getinfo() (*host.InfoStat, error) {
	return host.Info()
}

// Provide implements Provider.
func (is *provider) Provide() <-chan InfoStat {
	ch := make(chan InfoStat)
	go is.provideInternal(ch)
	return ch
}

func (is *provider) provideInternal(ch chan<- InfoStat) {
	defer close(ch)

	infoStat, err := is.internalExecutor.Getinfo()
	if err != nil {
		is.logger.Error("InfoStat command execution failed", zap.Error(err))
		return
	}

	infoStatDetails := InfoStat{
		Hostname:             infoStat.Hostname,
		BootTime:             infoStat.BootTime,
		Os:                   infoStat.OS,
		Platform:             infoStat.Platform,
		PlatformFamily:       infoStat.PlatformFamily,
		PlatformVersion:      infoStat.PlatformVersion,
		KernelVersion:        infoStat.KernelVersion,
		KernelArchitecture:   infoStat.KernelArch,
		VirtualizationSystem: infoStat.VirtualizationSystem,
		VirtualizationRole:   infoStat.VirtualizationRole,
		HostID:               infoStat.HostID,
	}

	is.logger.Debug(fmt.Sprintf("InfoStat provided %+v", infoStatDetails))
	ch <- infoStatDetails
}
