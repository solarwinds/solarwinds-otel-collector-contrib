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

package language

import (
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/cli"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"go.uber.org/zap"
)

const (
	localeCommand = "localectl"
	language      = "LANGUAGE"
)

type provider struct {
	cli    cli.CommandLineExecutor
	logger *zap.Logger
}

var _ providers.Provider[Language] = (*provider)(nil)

// Provide implements Provider.
func (lp *provider) Provide() <-chan Language {
	ch := make(chan Language)
	go func() {
		defer close(ch)
		stdout, err := cli.ProcessCommand(lp.cli, localeCommand, lp.logger)
		if err == nil {
			// localeCommand output is in format LANGUAGE=en_US
			parsedOutput := providers.ParseKeyValue(stdout, "=", []string{language})
			ch <- Language{
				Name: parsedOutput[language],
			}
		}
	}()
	return ch
}

func CreateLanguageProvider(logger *zap.Logger) providers.Provider[Language] {
	return &provider{
		cli:    &cli.BashCli{},
		logger: logger,
	}
}
