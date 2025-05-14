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

package providers

import "strings"

// ParseKeyValue takes text, splits it to separate lines, then
// is trying to find keys in lines split by the passed separator.
func ParseKeyValue(text, separator string, keys []string) map[string]string {
	attributesMap := make(map[string]string)
	for _, line := range strings.Split(text, "\n") {
		splitAttribute := strings.Split(strings.TrimSpace(line), separator)
		// correct result should consist of key and value property
		if len(splitAttribute) != 2 {
			continue
		}

		// Add wanted keys only to the map
		for _, key := range keys {
			if key == splitAttribute[0] {
				attributesMap[key] = splitAttribute[1]
				break
			}
		}
	}
	return attributesMap
}
