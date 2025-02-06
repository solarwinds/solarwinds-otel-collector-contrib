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
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/providers/language"
)

// Following keys will be used in metric as attributes.
const (
	languageLCID        = "osdetails.language.lcid"
	languageName        = "osdetails.language.name"
	languageDisplayName = "osdetails.language.displayname"
	languageCount       = 3
)

type generator struct {
	LanguageProvider providers.Provider[language.Language]
}

var _ shared.AttributesGenerator = (*generator)(nil)

func CreateLanguageAttributesGenerator(
	lp providers.Provider[language.Language],
) shared.AttributesGenerator {
	return &generator{
		LanguageProvider: lp,
	}
}

// Generate implements shared.AttributesGenerator.
func (g *generator) Generate() shared.AttributesChannel {
	ch := make(shared.AttributesChannel)
	go g.generateInternal(ch)
	return ch
}

func (g *generator) generateInternal(ch shared.AttributesChannel) {
	defer close(ch)

	lCh := g.LanguageProvider.Provide()
	for l := range lCh {
		atts := processLanguageAttributes(l)
		ch <- atts
	}
}

func processLanguageAttributes(l language.Language) shared.Attributes {
	m := make(shared.Attributes, languageCount)
	if l.Name != "" {
		m[languageName] = l.Name
	}
	if l.DisplayName != "" {
		m[languageDisplayName] = l.DisplayName
	}
	if l.LCID != 0 {
		m[languageLCID] = strconv.Itoa(l.LCID)
	}
	return m
}
