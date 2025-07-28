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

package domain

import (
	"fmt"
	"sync"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/cli"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/synchronization"
	"go.uber.org/zap"
)

const (
	fqdnCommand   = "hostname --fqdn"
	domainCommand = "hostname -d"
)

type provider struct {
	cli    cli.CommandLineExecutor
	logger *zap.Logger
}

var _ providers.Provider[Domain] = (*provider)(nil)

// Provide implements Provider.
func (dp *provider) Provide() <-chan Domain {
	ch := make(chan Domain)
	go dp.provideInternal(ch)
	return ch
}

func CreateDomainProvider(logger *zap.Logger) providers.Provider[Domain] {
	return &provider{
		cli:    &cli.BashCli{},
		logger: logger,
	}
}

func (dp *provider) provideInternal(ch chan Domain) {
	defer close(ch)
	var wg sync.WaitGroup

	wg.Add(2)
	fqdnCh := dp.executeLinuxCommand(fqdnCommand)
	dCh := dp.executeLinuxCommand(domainCommand)

	// synchronization channel for breaking endless cycle
	terminationCh := synchronization.ActivateSupervisingRoutine(&wg)

	domain := Domain{
		DomainRole: -1,
	}
loop:
	for {
		select {
		case fqdn, ok := <-fqdnCh:
			if ok {
				domain.FQDN = fqdn
			} else {
				fqdnCh = nil
				wg.Done()
			}
		case d, ok := <-dCh:
			if ok {
				domain.Domain = d
			} else {
				dCh = nil
				wg.Done()
			}
		case <-terminationCh:
			break loop
		}
	}

	dp.logger.Debug(fmt.Sprintf("domain provider result: %+v", domain))
	wg.Wait()
	ch <- domain
}

// executeLinuxCommand applies passed command and fills the channel with
// the result from stdout. This string is already in the expected form
// without the need to parse the value.
func (dp *provider) executeLinuxCommand(command string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		stdout, err := cli.ProcessCommand(dp.cli, command, dp.logger)
		if err == nil {
			ch <- stdout
		}
	}()
	return ch
}
