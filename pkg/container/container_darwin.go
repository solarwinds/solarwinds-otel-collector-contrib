//go:build darwin

package container

import (
	"fmt"
)

var errNotSupported = fmt.Errorf("darwin containers are not supported")

type containerInfo struct{}

func newContainerInfo() Provider {
	return &containerInfo{}
}

func (c *containerInfo) ReadContainerInstanceID() (string, error) {
	return "", errNotSupported
}

func (c *containerInfo) IsRunInContainerd() bool {
	return false
}
