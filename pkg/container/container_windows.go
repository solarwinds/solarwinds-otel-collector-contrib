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

//go:build windows

package container

import (
	"go.uber.org/zap"
	"golang.org/x/sys/windows/registry"
)

type containerInfo struct {
	logger *zap.Logger
}

func NewProvider(logger *zap.Logger) Provider {
	return &containerInfo{
		logger: logger,
	}
}

// ReadContainerInstanceID detects if we're running inside a docker container and returns the instance id if available
func (c *containerInfo) ReadContainerInstanceID() (string, error) {
	// Check for ContainerType in HKLM\System\CurrentControlSet\Control
	registryKeyControl, err := registry.OpenKey(registry.LOCAL_MACHINE, `System\CurrentControlSet\Control`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := registryKeyControl.Close(); closeErr != nil {
			c.logger.Error("failed to close registry key", zap.Error(closeErr))
		}
	}()

	// Retrieve the ContainerType value (to make sure it exists)
	// If it does, we are running in a Windows Container
	_, _, err = registryKeyControl.GetIntegerValue("ContainerType")
	if err != nil {
		return "", err
	}

	// Check for ComputerName in HKLM\System\CurrentControlSet\Control\ComputerName\ComputerName
	registryKeyCompName, err := registry.OpenKey(registry.LOCAL_MACHINE, `System\CurrentControlSet\Control\ComputerName\ComputerName`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}

	defer func() {
		if closeErr := registryKeyCompName.Close(); closeErr != nil {
			c.logger.Error("failed to close registry key", zap.Error(closeErr))
		}
	}()

	// Retrieve the ComputerName value
	// Use it as Container Identifier
	containerID, _, err := registryKeyCompName.GetStringValue("ComputerName")
	return containerID, err
}

func (c *containerInfo) IsRunInContainerd() bool {
	return false
}
