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

package synchronization

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsContextClosed_ReturnsTrueOnClosedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := IsContextClosed(ctx)

	assert.True(t, result)
}

func Test_IsContextClosed_ReturnsFalseOnOpenedContext(t *testing.T) {
	ctx := context.Background()

	result := IsContextClosed(ctx)

	assert.False(t, result)
}

func Test_CancelContextWithCauseIfNotClosed_CancelOpenedContext(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())

	// context is not yet close
	assert.Nil(t, ctx.Err())

	cancellingCause := "just kill it"
	CancelContextWithCauseIfNotClosed(ctx, cancel, errors.New(cancellingCause))

	// context is closed
	assert.NotNil(t, ctx.Err())

	assert.Contains(t, ctx.Err().Error(), "context canceled")
	assert.Contains(t, context.Cause(ctx).Error(), cancellingCause)
}

func Test_CancelContextWithCauseIfNotClosed_NotCancelOnAlreadyClosedContext(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())

	// cancel for the first time
	primaryCancellationCause := "kill it immediately"
	cancel(errors.New(primaryCancellationCause))

	// attempt to cancel it again => intended to be skipped
	secondaryCancellationCause := "kill it again"
	CancelContextWithCauseIfNotClosed(ctx, cancel, errors.New(secondaryCancellationCause))

	// must contains primary cancellation cause and not secondary
	// => second cancellation was not processed
	assert.Contains(t, context.Cause(ctx).Error(), primaryCancellationCause)
	assert.NotContains(t, context.Cause(ctx).Error(), secondaryCancellationCause)
}
