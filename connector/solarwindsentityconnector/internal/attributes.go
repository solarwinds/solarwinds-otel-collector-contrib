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

package internal

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"strings"
)

type Attributes struct {
	Source      map[string]pcommon.Value
	Destination map[string]pcommon.Value
	Common      map[string]pcommon.Value
}

func IdentifyAttributes(resourceAttrs pcommon.Map, srcPrefix, destPrefix string) Attributes {
	attrs := Attributes{
		Source:      make(map[string]pcommon.Value),
		Destination: make(map[string]pcommon.Value),
		Common:      make(map[string]pcommon.Value),
	}
	for k, v := range resourceAttrs.All() {
		switch {
		case srcPrefix != "" && strings.HasPrefix(k, srcPrefix):
			attrs.Source[getWithoutPrefix(srcPrefix, k)] = v
		case destPrefix != "" && strings.HasPrefix(k, destPrefix):
			attrs.Destination[getWithoutPrefix(destPrefix, k)] = v
		default:
			attrs.Common[k] = v
		}
	}

	return attrs
}

func getWithoutPrefix(prefix, key string) string {
	return strings.TrimPrefix(key, prefix)
}
