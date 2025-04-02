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

package hostdetails

import (
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/attributes/shared"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers"
	"github.com/solarwinds/solarwinds-otel-collector-releases/receiver/swohostmetricsreceiver/internal/providers/model"
)

// Following keys will be used in metric as attributes.
const (
	modelManufacturer = "hostdetails.model.manufacturer"
	modelSerialnumber = "hostdetails.model.serialnumber"
	modelName         = "hostdetails.model.name"
	modelCount        = 3
)

type generator struct {
	ModelProvider providers.Provider[model.Model]
}

var _ shared.AttributesGenerator = (*generator)(nil)

func CreateModelAttributesGenerator(
	mp providers.Provider[model.Model],
) shared.AttributesGenerator {
	return &generator{
		ModelProvider: mp,
	}
}

// Generate implements shared.AttributesGenerator.
func (g *generator) Generate() shared.AttributesChannel {
	ch := make(shared.AttributesChannel)
	go g.generateInternal(ch)
	return ch
}

func (g *generator) generateInternal(
	ch shared.AttributesChannel,
) {
	defer close(ch)

	mCh := g.ModelProvider.Provide()
	for m := range mCh {
		atts := processModelAttributes(m)
		ch <- atts
	}
}

func processModelAttributes(model model.Model) shared.Attributes {
	m := make(shared.Attributes, modelCount)
	if model.Manufacturer != "" {
		m[modelManufacturer] = model.Manufacturer
	}
	if model.Name != "" {
		m[modelName] = model.Name
	}
	if model.SerialNumber != "" {
		m[modelSerialnumber] = model.SerialNumber
	}
	return m
}
