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

package extensionfinder

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func FindExtension[E any](
	l *zap.Logger,
	extensionName string,
	host component.Host,
) (E, error) {
	extID := new(component.ID)
	if err := extID.UnmarshalText([]byte(extensionName)); err != nil {
		msg := fmt.Sprintf("failed to parse extension ID %q", extensionName)
		l.Error(msg, zap.Error(err))
		return *new(E), fmt.Errorf("%s: %w", msg, err)
	}

	ext, found := host.GetExtensions()[*extID]
	if !found {
		msg := fmt.Sprintf("extension %q not found", extensionName)
		l.Error(msg)
		return *new(E), fmt.Errorf("%s", msg)
	}

	castedExtension, castedOK := ext.(E)
	if !castedOK {
		msg := fmt.Sprintf("extension %q is not a %T", extensionName, *new(E))
		l.Error(msg)
		return *new(E), fmt.Errorf("%s", msg)
	}

	return castedExtension, nil
}
