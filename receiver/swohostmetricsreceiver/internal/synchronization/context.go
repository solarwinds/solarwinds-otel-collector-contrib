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

import "context"

func IsContextClosed(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// Cancels given context ctx by given cancel function cancelFn
// passing cause of cancellation only if context ctx is not closed.
// If context was already closed nothing is done here.
func CancelContextWithCauseIfNotClosed(
	ctx context.Context,
	cancelFn context.CancelCauseFunc,
	cause error,
) {
	if !IsContextClosed(ctx) {
		cancelFn(cause)
	}
}
