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
	"strconv"
	"strings"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/cli"
	"go.uber.org/zap"
)

const (
	delimiter        = ";"
	command          = "rpm -qa --qf '%{NAME}" + delimiter + "%{VERSION}" + delimiter + "%{INSTALLTIME}\n'"
	nameIndex        = 0
	versionIndex     = 1
	installTimeIndex = 2
)

type rpmProvider struct {
	bash   cli.CommandLineExecutor
	logger *zap.Logger
}

var _ (Provider) = (*rpmProvider)(nil)

func NewRpmProvider(logger *zap.Logger) Provider {
	return createRpmProvider(
		cli.NewBashCliExecutor(),
		logger,
	)
}

func createRpmProvider(
	bash cli.CommandLineExecutor,
	logger *zap.Logger,
) Provider {
	return &rpmProvider{
		bash:   bash,
		logger: logger,
	}
}

// GetSoftware implements InstalledSoftwareProvider.
func (provider *rpmProvider) GetSoftware() ([]InstalledSoftware, error) {
	stdout, _, err := provider.bash.ExecuteCommand(command)
	if err != nil {
		return []InstalledSoftware{}, fmt.Errorf("RPM based installed software can not be obtained: %w", err)
	}

	result := provider.parse(stdout)
	return result, nil
}

func (*rpmProvider) parse(stdout string) []InstalledSoftware {
	result := make([]InstalledSoftware, 0)
	pkgs := strings.Split(stdout, "\n")
	for _, pkg := range pkgs {
		group := strings.Split(pkg, delimiter)
		if len(group) < 3 {
			continue
		}

		is := InstalledSoftware{
			Name:        group[nameIndex],
			Version:     group[versionIndex],
			InstallDate: assemblyTimestamp(group[installTimeIndex]),
			Publisher:   "",
		}
		result = append(result, is)
	}
	return result
}

func assemblyTimestamp(unixFormat string) string {
	seconds, err := strconv.ParseInt(unixFormat, 10, 64)
	if err != nil {
		return ""
	}
	t := time.Unix(seconds, 0)
	formatedDate := fmt.Sprintf(
		"%s-%s-%s",
		transformWithPredfinedLength(t.Year(), 4),
		transformWithPredfinedLength(int(t.Month()), 2),
		transformWithPredfinedLength(t.Day(), 2))
	return formatedDate
}

func transformWithPredfinedLength(number int, predefinedLength int) string {
	result := strconv.Itoa(number)
	currentLength := len(result)
	if currentLength < predefinedLength {
		prefix := strings.Repeat("0", predefinedLength-currentLength)
		result = prefix + result
	}
	return result
}
