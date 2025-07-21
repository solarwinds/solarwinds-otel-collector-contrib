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

package wmi

import (
	"reflect"
)

type ExecutorMock struct {
	errors  map[string]error
	results map[string]interface{}
}

func CreateWmiExecutorMock(results []interface{}, errors map[interface{}]error) Executor {
	mock := &ExecutorMock{
		errors:  make(map[string]error),
		results: make(map[string]interface{}),
	}

	for _, result := range results {
		mock.SetupResult(result)
	}

	for i, err := range errors {
		mock.SetupError(i, err)
	}

	return mock
}

func (m *ExecutorMock) SetupResult(i interface{}) {
	m.results[getImplementationName(i)] = i
}

func (m *ExecutorMock) SetupError(i interface{}, error error) {
	m.errors[getImplementationName(i)] = error
}

func (*ExecutorMock) CreateQuery(_ interface{}, _ string) string { return "not relevant" }

func (m *ExecutorMock) Query(_ string, dst interface{}) (interface{}, error) {
	err := m.errors[getImplementationName(dst)]
	if err != nil {
		return nil, err
	}

	result := m.results[getImplementationName(dst)]
	return result, nil
}

func getImplementationName(i interface{}) string {
	return reflect.TypeOf(i).String()
}
