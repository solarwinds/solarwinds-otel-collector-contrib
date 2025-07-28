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

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/registry"
	"go.uber.org/zap"
)

type windowsProvider struct {
	registryReaders []registry.Reader
	logger          *zap.Logger
}

var _ (Provider) = (*windowsProvider)(nil)

func NewInstalledSoftwareProvider(logger *zap.Logger) Provider {
	return createInstalledSoftwareProvider(
		createRegistryReaders(logger),
		logger,
	)
}

func createInstalledSoftwareProvider(
	registryReaders []registry.Reader,
	logger *zap.Logger,
) Provider {
	return &windowsProvider{
		registryReaders: registryReaders,
		logger:          logger,
	}
}

func createRegistryReaders(logger *zap.Logger) []registry.Reader {
	// 64-bit reader
	uninstallRootPath64bit := `Software\Microsoft\Windows\CurrentVersion\Uninstall`
	registryReader64bit, err := registry.NewReader(registry.LocalMachineKey, uninstallRootPath64bit)
	if err != nil {
		logger.Error(
			"64-bit registry reader for path can not be created",
			zap.String("path", uninstallRootPath64bit),
			zap.Error(err),
		)
		return make([]registry.Reader, 0)
	}
	// 32-bit reader
	uninstallRootPath32bit := `Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`
	registryReader32bit, err := registry.NewReader(registry.LocalMachineKey, uninstallRootPath32bit)
	if err != nil {
		logger.Error(
			"32-bit registry reader for path can not be created",
			zap.String("path", uninstallRootPath32bit),
			zap.Error(err),
		)
		return make([]registry.Reader, 0)
	}
	return []registry.Reader{registryReader64bit, registryReader32bit}
}

// GetSoftware implements Provider.
func (provider *windowsProvider) GetSoftware() ([]InstalledSoftware, error) {
	// From initialization there were no registred registry readers.
	if len(provider.registryReaders) == 0 {
		m := "no register reader registered"
		provider.logger.Error(m)
		return make([]InstalledSoftware, 0), fmt.Errorf("%s", m)
	}

	values := []InstalledSoftware{}
	for _, registryReader := range provider.registryReaders {
		readerValues, err := processRegistryReader(registryReader, provider.logger)
		if err != nil {
			return values, err
		}
		values = append(values, readerValues...)
	}
	return values, nil
}

func processRegistryReader(registryReader registry.Reader, logger *zap.Logger) ([]InstalledSoftware, error) {
	values := []InstalledSoftware{}
	uninstallKeys, err := registryReader.GetSubKeys()
	if err != nil {
		message := "Windows installed software can not be obtained, failed to get installed software registry keys"
		logger.Error(message, zap.Error(err))
		return values, fmt.Errorf("%s:%w", message, err)
	}

	for _, swKeyName := range uninstallKeys {
		registryValues, err := registryReader.GetKeyValues(swKeyName, []string{"DisplayName", "DisplayVersion", "Publisher", "InstallDate"})
		if err != nil {
			message := fmt.Sprintf("unable to read from the %s registry key", swKeyName)
			logger.Warn(message, zap.Error(err))
			continue
		}

		value := InstalledSoftware{
			Name:        registryValues["DisplayName"],
			Version:     registryValues["DisplayVersion"],
			Publisher:   registryValues["Publisher"],
			InstallDate: formatDate(registryValues["InstallDate"]),
		}

		if value.Name != "" {
			values = append(values, value)
		}
	}

	return values, nil
}

func formatDate(inputDate string) string {
	const dateLength int = 8
	outputDate := ""
	if len(inputDate) == dateLength {
		outputDate = fmt.Sprintf(
			"%s-%s-%s",
			inputDate[:4],
			inputDate[4:6],
			inputDate[6:])
	}

	return outputDate
}
