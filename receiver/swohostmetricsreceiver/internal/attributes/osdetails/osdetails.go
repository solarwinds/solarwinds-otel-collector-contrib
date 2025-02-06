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
	"sync"

	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector/receiver/swohostmetricsreceiver/internal/synchronization"
)

type attributesGenerator struct {
	InfoStatAttributesGenerator shared.AttributesGenerator
	LanguageAttributesGenerator shared.AttributesGenerator
}

// Generate implements shared.AttributesGenerator.
func (g *attributesGenerator) Generate() shared.AttributesChannel {
	ch := make(shared.AttributesChannel)
	go g.generateInternal(ch)
	return ch
}

var _ shared.AttributesGenerator = (*attributesGenerator)(nil)

func CreateOsDetailsAttributesGenerator(
	infostat shared.AttributesGenerator,
	language shared.AttributesGenerator,
) shared.AttributesGenerator {
	return &attributesGenerator{
		InfoStatAttributesGenerator: infostat,
		LanguageAttributesGenerator: language,
	}
}

func (g *attributesGenerator) generateInternal(ch shared.AttributesChannel) {
	defer close(ch)

	// prepare for language attributes generator
	const attGeneratorsCount = 2
	var wg sync.WaitGroup
	wg.Add(attGeneratorsCount)

	terminationCh := synchronization.ActivateSupervisingRoutine(&wg)

	infoStatCh := g.InfoStatAttributesGenerator.Generate()
	languageCh := g.LanguageAttributesGenerator.Generate()

	allAttsCount := infoStatCount + languageCount
	osDetailsCh := make(shared.Attributes, allAttsCount)

loop:
	for {
		select {
		case isAtts, opened := <-infoStatCh:
			shared.ProcessReceivedAttributes(isAtts, osDetailsCh, opened, &infoStatCh, &wg)
		case lAtts, opened := <-languageCh:
			shared.ProcessReceivedAttributes(lAtts, osDetailsCh, opened, &languageCh, &wg)
		case <-terminationCh:
			break loop
		}
	}
	ch <- osDetailsCh
}
