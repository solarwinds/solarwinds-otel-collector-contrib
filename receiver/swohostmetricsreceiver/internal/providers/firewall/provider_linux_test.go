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

package firewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_Provide_ProvidesEmptyFirewallProfileCollectionWithNoErrors(t *testing.T) {
	sut := CreateFirewallProvider(zap.NewNop())
	ch := sut.Provide()
	actualModel := <-ch
	_, open := <-ch

	assert.Nil(t, actualModel.FirewallProfiles, "unsupported provider must return no data")
	assert.Nil(t, actualModel.Error, "unsupported provider must return no error")
	assert.False(t, open, "channel must be closed afterward")
}
