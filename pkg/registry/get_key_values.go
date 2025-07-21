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

package registry

type GetKeyValuesTypeFunc func(rootKey RootKey, parent, keyName string, names []string) (map[string]string, error)

func GetKeyValues(rootKey RootKey, parent, keyName string, names []string) (map[string]string, error) {
	regReader, err := NewReader(rootKey, parent)
	if err != nil {
		return nil, err
	}

	return regReader.GetKeyValues(keyName, names)
}

type GetKeyUIntValueTypeFunc func(rootKey RootKey, parent, keyName string, name string) (uint64, error)

func GetKeyUIntValue(rootKey RootKey, parent, keyName string, name string) (uint64, error) {
	regReader, err := NewReader(rootKey, parent)
	if err != nil {
		return 0, err
	}

	return regReader.GetKeyUIntValue(keyName, name)
}
