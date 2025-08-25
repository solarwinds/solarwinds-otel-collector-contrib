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
