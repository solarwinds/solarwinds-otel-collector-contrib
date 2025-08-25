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

package solarwindsprocessor

import (
	"fmt"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/extension/solarwindsextension"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/extensionfinder"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type ExtensionProvider interface {
	Init(logger *zap.Logger, extensionName string, host component.Host) (*solarwindsextension.SolarwindsExtension, error)
	SetAttributes(resourceAttributes *map[string]string) error
}

type extensionProvider struct {
	extension *solarwindsextension.SolarwindsExtension
}

var _ ExtensionProvider = (*extensionProvider)(nil)

// NewExtensionProvider creates a new instance of ExtensionProvider.
func NewExtensionProvider() ExtensionProvider {
	return &extensionProvider{}
}

func (ep *extensionProvider) Init(logger *zap.Logger, extensionName string, host component.Host) (*solarwindsextension.SolarwindsExtension, error) {
	if ep.extension == nil {
		swiExtension, err := extensionfinder.FindExtension[*solarwindsextension.SolarwindsExtension](
			logger,
			extensionName,
			host,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to find Solarwinds Extension %q", extensionName)
		}

		ep.extension = swiExtension
	}

	return ep.extension, nil
}

func (ep *extensionProvider) SetAttributes(resourceAttributes *map[string]string) error {
	if ep.extension == nil {
		return fmt.Errorf("solarwinds extension is nil")
	}

	cfg := ep.extension.GetCommonConfig()
	if cfg == nil {
		return fmt.Errorf("failed to get config from solarwinds extension")
	}

	attrs := *resourceAttributes
	// Send entity creation flag if the extension is configured to infer collector entity.
	if !cfg.WithoutEntity() {
		attrs[solarwindsextension.EntityCreation] = solarwindsextension.EntityCreationValue
	}

	// Set collector name.
	attrs[solarwindsextension.CollectorNameAttribute] = cfg.CollectorName()

	return nil
}
