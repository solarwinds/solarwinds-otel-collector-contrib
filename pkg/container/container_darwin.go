//go:build darwin

package container

import (
	"fmt"
)

type containerInfo struct{}

func newProvider() Provider {
	return &containerInfo{}
}

func (c *containerInfo) ReadContainerInstanceID() (string, error) {
	return "", fmt.Errorf("darwin containers are not supported")
}

func (c *containerInfo) IsRunInContainerd() bool {
	return false
}
