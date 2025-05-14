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

package shared

import "sync"

// Process incoming attributes from a particular channel.
// In case channel is close the channel pointer is set to nil
// and waitGroup gets the signal that this goroutine is done.
func ProcessReceivedAttributes(
	generatedAttrs Attributes,
	resultingAttrs Attributes,
	opened bool,
	ch *AttributesChannel,
	wg *sync.WaitGroup,
) {
	if !opened {
		*ch = nil
		wg.Done()
	} else {
		mergeAttributes(generatedAttrs, resultingAttrs)
	}
}

// Takes already existing attributes and new ones. It merges them.
func mergeAttributes(
	increment Attributes,
	result Attributes,
) {
	for k, v := range increment {
		result[k] = v
	}
}
