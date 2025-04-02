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

package solarwindsextension

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"

	"github.com/solarwinds/solarwinds-otel-collector-releases/extension/solarwindsextension/internal"
	"github.com/solarwinds/solarwinds-otel-collector-releases/extension/solarwindsextension/internal/metadata"
)

func NewFactory() extension.Factory {
	return extension.NewFactory(
		metadata.Type,
		internal.NewDefaultConfig,
		createExtension,
		metadata.ExtensionStability)
}

func createExtension(ctx context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
	extCfg, ok := cfg.(*internal.Config)
	if !ok {
		return nil, fmt.Errorf("unexpected config type: %T", cfg)
	}
	return newExtension(ctx, set, extCfg)
}
