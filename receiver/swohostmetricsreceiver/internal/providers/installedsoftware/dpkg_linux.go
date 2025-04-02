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

package installedsoftware

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/cli"
	"go.uber.org/zap"
)

type dpkgProvider struct {
	bash cli.CommandLineExecutor
}

const parsingRegexpPattern = "^ii[\\s]+([[:graph:]]+)[\\s]+([[:graph:]]+)[\\s]+.+$"

var parsingRegexp = regexp.MustCompile(parsingRegexpPattern)

var _ (Provider) = (*dpkgProvider)(nil)

func NewDpkgProvider() Provider {
	return createDpkgProvider(
		cli.NewBashCliExecutor(),
	)
}

func createDpkgProvider(
	bash cli.CommandLineExecutor,
) Provider {
	return &dpkgProvider{
		bash: bash,
	}
}

// GetSoftware implements InstalledSoftwareProvider.
func (provider *dpkgProvider) GetSoftware() ([]InstalledSoftware, error) {
	command := "dpkg -l | grep ^ii"

	stdout, _, err := provider.bash.ExecuteCommand(command)
	if err != nil {
		message := "DPKG based installed software can not be obtained"
		zap.L().Error(message, zap.Error(err))
		return []InstalledSoftware{}, fmt.Errorf(message+": %w", err)
	}

	result := provider.parse(stdout)
	return result, nil
}

func (*dpkgProvider) parse(commandOutput string) []InstalledSoftware {
	regexpGroupMembersCount := 3
	packageIndex := 1
	versionIndex := 2

	result := make([]InstalledSoftware, 0)
	pkgs := strings.Split(commandOutput, "\n")
	for _, pkg := range pkgs {
		groups := parsingRegexp.FindStringSubmatch(pkg)
		if len(groups) != regexpGroupMembersCount {
			continue
		}

		is := InstalledSoftware{
			Name:    groups[packageIndex],
			Version: groups[versionIndex],
		}
		result = append(result, is)
	}
	return result
}
