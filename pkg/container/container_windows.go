//go:build windows

package container

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

var errNotSupported = fmt.Errorf("windows containers are not fully supported")

type containerInfo struct{}

func newContainerInfo() Provider {
	return &containerInfo{}
}

// ReadContainerInstanceID detects if we're running inside of a docker container and returns the instance id if available
func (c *containerInfo) ReadContainerInstanceID() (string, error) {
	// Check for ContainerType in HKLM\System\CurrentControlSet\Control
	registryKeyControl, err := registry.OpenKey(registry.LOCAL_MACHINE, `System\CurrentControlSet\Control`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer registryKeyControl.Close()

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
	defer registryKeyCompName.Close()

	// Retrieve the ComputerName value
	// Use it as Container Identifier
	containerID, _, err := registryKeyCompName.GetStringValue("ComputerName")
	return containerID, err
}

func (c *containerInfo) IsRunInContainerd() bool {
	return false
}
