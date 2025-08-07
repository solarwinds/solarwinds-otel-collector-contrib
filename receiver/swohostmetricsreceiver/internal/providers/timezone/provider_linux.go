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

package timezone

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/cli"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"go.uber.org/zap"
)

const (
	timeZoneCommand = "timedatectl | grep 'Time zone'"
	timeZone        = "Time zone"
)

type provider struct {
	cli    cli.CommandLineExecutor
	logger *zap.Logger
}

var _ providers.Provider[TimeZone] = (*provider)(nil)

func CreateTimeZoneProvider(logger *zap.Logger) providers.Provider[TimeZone] {
	return &provider{
		cli:    &cli.BashCli{},
		logger: logger,
	}
}

// Provide implements Provider.
func (dp *provider) Provide() <-chan TimeZone {
	ch := make(chan TimeZone)
	go func() {
		defer close(ch)
		stdout, err := cli.ProcessCommand(dp.cli, timeZoneCommand, dp.logger)
		// Bias field has to be set to invalid value because it
		// defaults to zero, which is valid in this context
		result := TimeZone{
			Bias: -1,
		}
		if err == nil {
			// The output contains result in format Time zone: xxx
			parsedOutput := providers.ParseKeyValue(stdout, ": ", []string{timeZone})
			result.Caption = parsedOutput[timeZone]
		}
		ch <- result
	}()
	return ch
}
