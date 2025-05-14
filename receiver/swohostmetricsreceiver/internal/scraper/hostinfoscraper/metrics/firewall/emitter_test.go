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

//go:build !integration

package firewall

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Emitter_Functional(t *testing.T) {
	t.Skip("This test should be run manually")

	sut := NewEmitter()
	if err := sut.Init(); err != nil {
		assert.Fail(t, "initialization must not fail")
	}
	er := sut.Emit()
	if er.Error != nil {
		assert.Fail(t, "metric emitation must not fail")
	}
	fmt.Printf("Result: %+v", er.Data)
}
