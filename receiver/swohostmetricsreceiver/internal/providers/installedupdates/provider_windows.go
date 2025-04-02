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

package installedupdates

import (
	"fmt"
	"strings"

	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/wmi"
	"go.uber.org/zap"
)

type windowsProvider struct {
	wmi wmi.Executor
}

var _ (Provider) = (*windowsProvider)(nil)

func NewProvider() Provider {
	return createWindowsProvider(
		wmi.NewExecutor(),
	)
}

func createWindowsProvider(
	wmi wmi.Executor,
) Provider {
	return &windowsProvider{
		wmi: wmi,
	}
}

// Win32_QuickFixEngineering represents actual Hot Fix WMI Object
// with subset of fields required for scraping.
type Win32_QuickFixEngineering struct {
	Caption     string
	Description string
	HotFixID    string
	InstalledBy string
	InstalledOn string
}

func (provider *windowsProvider) GetUpdates() ([]InstalledUpdate, error) {
	result, err := wmi.QueryResult[[]Win32_QuickFixEngineering](provider.wmi)
	if err != nil {
		message := "Invalid installed updates output."
		zap.L().Error(message, zap.Error(err))

		return []InstalledUpdate{}, fmt.Errorf("%s %w", message, err)
	}

	var updates []InstalledUpdate
	for _, update := range result {
		updates = append(updates, InstalledUpdate{
			Caption:     update.Caption,
			HotFixID:    update.HotFixID,
			InstalledOn: update.InstalledOn,
			InstalledBy: update.InstalledBy,
			Description: update.Description,
		})
	}

	return formatDates(updates), nil
}

// Format date from mm/dd/yyyy to yyyy-mm-dd for all updates.
func formatDates(updates []InstalledUpdate) []InstalledUpdate {
	formattedUpdates := []InstalledUpdate{}
	for _, update := range updates {
		formattedUpdate := InstalledUpdate{
			Caption:     update.Caption,
			Description: update.Description,
			HotFixID:    update.HotFixID,
			InstalledOn: formatDate(update.InstalledOn),
			InstalledBy: update.InstalledBy,
		}
		formattedUpdates = append(formattedUpdates, formattedUpdate)
	}

	return formattedUpdates
}

// format date from mm/dd/yyyy to yyyy-mm-dd.
func formatDate(inputDate string) string {
	outputDate := ""
	dateParts := strings.Split(inputDate, "/")
	if len(dateParts) == 3 {
		outputDate = fmt.Sprintf("%s-%s-%s",
			dateParts[2],
			dateParts[0],
			dateParts[1])
	}

	return outputDate
}
