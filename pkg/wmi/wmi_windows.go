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

type executor struct{}

func NewExecutor() Executor {
	return &executor{}
}

// CreateQuery creates string WMI query based on the input.
// Note: Implementation of WMI always requires pointer to array of src.
func (*executor) CreateQuery(src interface{}, where string) string {
	return wmi.CreateQuery(src, where)
}

// Query executes WMI query and modifies dst with its result.
// Returns modified object and error in case query fails.
// Note: Implementation of WMI always requires pointer to array of dst.
func (*executor) Query(query string, dst interface{}) (interface{}, error) {
	err := wmi.Query(query, dst)
	return dst, err
}
