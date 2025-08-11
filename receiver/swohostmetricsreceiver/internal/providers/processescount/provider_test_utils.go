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

package processescount

type succeedingWrapper struct {
	value int64
}

var _ Wrapper = (*succeedingWrapper)(nil)

func CreateSucceedingWrapper(
	value int64,
) Wrapper {
	return &succeedingWrapper{
		value: value,
	}
}

// GetCount implements Wrapper.
func (w *succeedingWrapper) GetCount() (int64, error) {
	return w.value, nil
}

type failingWrapper struct {
	err error
}

var _ Wrapper = (*failingWrapper)(nil)

func CreateFailingUptimeWrapper(
	err error,
) Wrapper {
	return &failingWrapper{
		err: err,
	}
}

// GetCount implements Wrapper.
func (w *failingWrapper) GetCount() (int64, error) {
	return 0, w.err
}
