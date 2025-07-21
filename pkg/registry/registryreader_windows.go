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

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type windowsRegistryReader struct {
	RootPath string
	RootKey  registry.Key
}

var _ (Reader) = (*windowsRegistryReader)(nil)

func NewReader(rootKey RootKey, rootPath string) (Reader, error) {
	if len(rootPath) == 0 {
		return nil, fmt.Errorf("rootPath can't be empty")
	}
	key, err := mapRootKey(rootKey)
	if err != nil {
		return nil, err
	}
	newInstance := windowsRegistryReader{RootPath: rootPath, RootKey: key}

	return newInstance, nil
}

func mapRootKey(rootKey RootKey) (registry.Key, error) {
	var key registry.Key
	switch rootKey {
	case LocalMachineKey:
		return registry.LOCAL_MACHINE, nil
	case CurrentUserKey:
		return registry.CURRENT_USER, nil
	default:
		return key, fmt.Errorf("undefined rootKey: %d", rootKey)
	}
}

func (reader windowsRegistryReader) GetSubKeys() ([]string, error) {
	uninstallRootKey, err := registry.OpenKey(reader.RootKey, reader.RootPath, registry.READ)
	if err != nil {
		return nil, err
	}
	defer uninstallRootKey.Close()

	uninstallKeys, err := uninstallRootKey.ReadSubKeyNames(0)
	if err != nil {
		return nil, err
	}

	return uninstallKeys, nil
}

func (reader windowsRegistryReader) GetKeyValues(keyName string, names []string) (map[string]string, error) {
	key, err := reader.openKey(keyName)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	values := map[string]string{}
	for _, name := range names {
		value, err := reader.getStringValue(&key, name)
		if err != nil {
			values[name] = ""
		} else {
			values[name] = value
		}
	}

	return values, nil
}

func (reader windowsRegistryReader) GetKeyUIntValue(keyName string, name string) (uint64, error) {
	key, err := reader.openKey(keyName)
	if err != nil {
		return 0, err
	}
	defer key.Close()

	value, err := reader.getIntValue(&key, name)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (reader windowsRegistryReader) openKey(keyName string) (registry.Key, error) {
	path := fmt.Sprintf("%s\\%s", reader.RootPath, keyName)
	return registry.OpenKey(reader.RootKey, path, registry.READ)
}

func (windowsRegistryReader) getIntValue(key *registry.Key, name string) (uint64, error) {
	value, _, err := key.GetIntegerValue(name)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (windowsRegistryReader) getStringValue(key *registry.Key, name string) (string, error) {
	value, _, err := key.GetStringValue(name)
	if err != nil {
		return "", err
	}

	return value, nil
}
