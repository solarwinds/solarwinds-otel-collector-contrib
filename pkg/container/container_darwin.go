//go:build darwin

package container

import (
	"go.uber.org/zap"
)

type containerInfo struct {
	logger *zap.Logger
}

func NewProvider(logger *zap.Logger) Provider {
	return &containerInfo{
		logger: logger,
	}
}

func (c *containerInfo) ReadContainerInstanceID() (string, error) {
	c.logger.Warn("darwin containers are not supported")
	return "", nil
}

func (c *containerInfo) IsRunInContainerd() bool {
	c.logger.Warn("darwin containers are not supported")
	return false
}
