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
	"fmt"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/registry"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swohostmetricsreceiver/internal/providers"
	"go.uber.org/zap"
)

const (
	controlPanelKey  = "Control Panel"
	internationalKey = "International"
	localeValueName  = "LocaleName"
)

type provider struct {
	displayNamesMap   map[string]string
	getRegistryValues registry.GetKeyValuesTypeFunc
	lcidProvider      LCIDProvider
	logger            *zap.Logger
}

var _ providers.Provider[Language] = (*provider)(nil)

func CreateLanguageProvider(logger *zap.Logger) providers.Provider[Language] {
	return &provider{
		displayNamesMap:   getDisplayLanguages(),
		getRegistryValues: registry.GetKeyValues,
		lcidProvider:      NewWindowsLCIDProvider(),
		logger:            logger,
	}
}

// Provide implements Provider.
func (p *provider) Provide() <-chan Language {
	ch := make(chan Language)
	go func() {
		defer close(ch)
		lang, err := p.getLanguageInfo()
		if err != nil {
			p.logger.Error("failed to get language info", zap.Error(err))
		}

		if err == nil {
			ch <- *lang
		}
	}()
	return ch
}

func (p *provider) getLanguageInfo() (*Language, error) {
	names := []string{localeValueName}
	values, err := p.getRegistryValues(registry.CurrentUserKey, controlPanelKey, internationalKey, names)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user locale info: %w", err)
	}

	shortName := values[localeValueName]
	displayName, ok := p.displayNamesMap[shortName]
	if !ok {
		return nil, fmt.Errorf("display name not found for language '%s'", shortName)
	}

	lcid := p.lcidProvider.GetUserDefaultLCID()
	lang := &Language{
		LCID:        int(lcid),
		Name:        shortName,
		DisplayName: displayName,
	}

	return lang, nil
}
