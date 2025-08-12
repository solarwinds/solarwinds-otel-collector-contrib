//go:build darwin

package container

import (
	"go.uber.org/zap"
)

type containerInfo struct{}

func newProvider() Provider {
	return &containerInfo{}
}

func (c *containerInfo) ReadContainerInstanceID() (string, error) {
	zap.L().Warn("darwin containers are not supported")
	return "", nil
}

func (c *containerInfo) IsRunInContainerd() bool {
	return false
}
