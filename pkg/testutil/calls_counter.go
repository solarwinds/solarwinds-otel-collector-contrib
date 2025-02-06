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

package testutil

import "sync/atomic"

type CallsCounter struct {
	calls atomic.Int32
}

func (c *CallsCounter) SetupInitialCount(initialCount int32) {
	c.calls.Store(initialCount)
}

func (c *CallsCounter) IncrementCount() {
	c.calls.Add(1)
}

func (c *CallsCounter) GetCallsCount() int32 {
	return c.calls.Load()
}
