//go:build linux

package container

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

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
	var err error
	var containerID string

	switch {
	case c.isInDocker():
		containerID, err = c.readDockerInstanceID()
	case c.isInNonDockerContainer():
		containerID, err = c.readNonDockerContainerID()
	case c.isContainerInKubernetes():
		hostnameEnv, _ := os.LookupEnv("HOSTNAME")
		containerID = hostnameEnv
	default:
		containerID, err = "", errors.New("it is not a container instance")
	}

	if containerID == "" {
		err = fmt.Errorf("failed to detect container ID: %w", err)
	}

	return containerID, err
}

func (c *containerInfo) IsRunInContainerd() bool {
	data, err := os.ReadFile("/proc/self/cpuset")
	if err != nil {
		return false
	}

	// container running on containerD in k8s does not have such content in this file
	if strings.Contains(string(data), "/default/") {
		return true
	}

	data, err = os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return false
	}

	if strings.Contains(string(data), "containerd") {
		return true
	}

	return false
}

// readDockerInstanceID returns the docker instance id if available.
func (c *containerInfo) readDockerInstanceID() (string, error) {
	data, err := os.ReadFile("/proc/self/cpuset")
	if err != nil {
		return "", err
	}

	// Attempt to read the Docker instance name if it's running in a Kubernetes pod with cgroups V1
	if strings.Contains(string(data), "/kubepods/") {
		s := strings.Split(string(data), "/")
		if len(s) > 0 {
			return strings.TrimSpace(s[len(s)-1]), err
		}
	}

	// Attempt to read the Docker instance name using the default way on cgroups V1
	if strings.Contains(string(data), "/docker/") {
		s := strings.Split(string(data), "/")
		if len(s) > 0 {
			return strings.TrimSpace(s[2]), err
		}
	}

	// Attempt to read the Docker instance on cgroups V2
	data, err = os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}

	if strings.Contains(string(data), "docker") {
		lines := strings.Split(string(data), "\n")
		var line string
		for _, l := range lines {
			if strings.Contains(l, "containers") {
				line = l
			}
		}
		if line == "" {
			return "", fmt.Errorf("unable to read Docker on cgroup v2 instance identifier ")
		}

		words := strings.Split(line, "/")
		for i, word := range words {
			if strings.Contains(word, "containers") && len(words) > i+1 {
				return words[i+1], nil
			}
		}
	}

	return "", fmt.Errorf("unable to read Docker instance identifier")
}

// readNonDockerContainerID now returns identifier only for containerd and Podman containers.
//
//nolint:gocognit
func (c *containerInfo) readNonDockerContainerID() (string, error) { //nolint:funlen
	data, err := os.ReadFile("/proc/self/cpuset")
	if err != nil {
		c.logger.Info("error reading cpuset")
	}

	// Attempt to read the Containerd instance name using the default way on cgroups V1
	if strings.Contains(string(data), "/default/") {
		tokenized := strings.Split(string(data), "/")
		if len(tokenized) > 0 {
			return strings.TrimSpace(tokenized[2]), err
		}
	}

	// Attempt to read the Podman container name using the default way on cgroups V1
	if strings.Contains(string(data), "libpod") {
		tokenized := strings.Split(string(data), "-")
		if len(tokenized) > 0 {
			return strings.TrimSpace(strings.ReplaceAll(tokenized[1], ".scope", "")), err
		}
	}

	if strings.Contains(string(data), "cri-containerd-") {
		tokenized := strings.Split(string(data), "/")
		if len(tokenized) > 0 {
			return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(tokenized[len(tokenized)-1], ".scope", ""), "cri-containerd-", "")), err
		}
	}

	// Attempt to read the Containerd instance name if it's running in a Kubernetes pod with cgroups V1
	if strings.Contains(string(data), "/kubepods/") {
		s := strings.Split(string(data), "/")
		if len(s) > 0 {
			return strings.TrimSpace(s[len(s)-1]), err
		}
	}

	// Attempt to read the Podman container name using the default way on cgroups V2
	data, err = os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		c.logger.Info("error reading mountinfo")
	}

	if strings.Contains(string(data), "containers") {
		lines := strings.Split(string(data), "\n")
		var line string
		for _, l := range lines {
			if strings.Contains(l, "containers") {
				line = l
			}
		}
		if line != "" {
			words := strings.Split(line, "/")
			for _, word := range words {
				dst := make([]byte, hex.DecodedLen(64))

				if _, err := hex.Decode(dst, []byte(word)); err == nil {
					return word, nil
				}
			}
		}
	}

	data, err = os.ReadFile("/proc/self/cgroup")
	if err != nil {
		c.logger.Error("error reading cgroup")
	}

	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		words := strings.Split(l, "/")
		for _, word := range words {
			dst := make([]byte, hex.DecodedLen(64))

			if _, err := hex.Decode(dst, []byte(word)); err == nil {
				return word, nil
			}
		}
	}

	return "", fmt.Errorf("unable to read container instance identifier")
}

func (c *containerInfo) isContainerInKubernetes() bool {
	if hostnameEnv, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST"); ok && hostnameEnv != "" {
		return true
	}

	data, err := os.ReadFile("/proc/self/mounts")
	if err != nil {
		return false
	}

	return strings.Contains(string(data), "kubernetes")
}

func (c *containerInfo) isInNonDockerContainer() bool {
	data, err := os.ReadFile("/proc/self/mounts")
	if err == nil {
		if strings.Contains(string(data), "container") && !strings.Contains(string(data), "docker/containers") {
			return true
		}
	}

	_, err = os.Stat("/.containerenv")
	if err == nil {
		return true
	}

	_, err = os.Stat("/proc/2/status")
	if err != nil {
		return true
	}

	data, err = os.ReadFile("/proc/2/status")
	if err != nil {
		return false
	}

	return !strings.Contains(string(data), "kthread")
}

func (c *containerInfo) isInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
