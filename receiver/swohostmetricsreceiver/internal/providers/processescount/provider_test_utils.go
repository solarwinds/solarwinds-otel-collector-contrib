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

type succeedingCounter struct {
	value int64
}

var _ ProcessCounter = (*succeedingCounter)(nil)

func CreateSucceedingCounter(
	value int64,
) ProcessCounter {
	return &succeedingCounter{
		value: value,
	}
}

func (w *succeedingCounter) GetCount() (int64, error) {
	return w.value, nil
}

type failingCounter struct {
	err error
}

var _ ProcessCounter = (*failingCounter)(nil)

func CreateFailingCounter(
	err error,
) ProcessCounter {
	return &failingCounter{
		err: err,
	}
}

func (w *failingCounter) GetCount() (int64, error) {
	return 0, w.err
}
