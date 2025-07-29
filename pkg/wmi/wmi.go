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
	"errors"
	"fmt"
)

var (
	ErrWmiNoResult       = errors.New("wmi query failed: no result returned")
	ErrWmiEmptyResult    = errors.New("wmi query failed: empty result returned")
	ErrWmiTooManyResults = errors.New("wmi query failed: more than single results returned")
)

type Executor interface {
	CreateQuery(interface{}, string) string
	Query(string, interface{}) (interface{}, error)
}

// QueryResult utilizes executor to get TResult.
func QueryResult[TResult interface{}](executor Executor) (TResult, error) {
	var model TResult
	query := executor.CreateQuery(&model, "")
	result, err := executor.Query(query, &model)
	if err != nil {
		var nilResult TResult
		return nilResult, fmt.Errorf("wmi query failed: %w", err)
	}

	return *result.(*TResult), err
}

// QuerySingleResult utilizes executor to get single result of TResult.
// If not exactly single result is returned from the query, error with
// nil result is returned.
func QuerySingleResult[TResult interface{}](executor Executor) (TResult, error) {
	var nilResult TResult
	res, err := QueryResult[[]TResult](executor)
	if err != nil {
		return nilResult, err
	}

	if res == nil {
		return nilResult, ErrWmiNoResult
	}

	if len(res) == 0 {
		return nilResult, ErrWmiEmptyResult
	}

	if len(res) > 1 {
		return nilResult, ErrWmiTooManyResults
	}

	return res[0], nil
}
