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

package manifests

import (
	"errors"

	"github.com/distribution/reference"
)

type PodManifest struct {
	Metadata PodMetadata `json:"metadata"`
	Status   PodStatus   `json:"status"`
	Spec     PodSpec     `json:"spec"`
}

type PodMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PodStatus struct {
	ContainerStatuses     []statusContainer
	InitContainerStatuses []statusContainer
	Conditions            []PodCondition
}

type PodCondition struct {
	Timestamp string `json:"lastTransitionTime"`
}

type specContainer struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type PodSpec struct {
	Containers     []specContainer `json:"containers"`
	InitContainers []struct {
		specContainer
		RestartPolicy string `json:"restartPolicy"`
	} `json:"initContainers"`
}

type statusContainer struct {
	Name        string         `json:"name"`
	ContainerId string         `json:"containerID"`
	State       map[string]any `json:"state"`
	ImageID     string         `json:"imageID"`
}

type Container struct {
	Name               string
	ContainerId        string
	State              string
	IsInitContainer    bool
	IsSidecarContainer bool
	Image              Image
}

type Image struct {
	ImageID string
	Name    string
	Tag     string
}

// GetContainers returns a map of containers from the manifest. Data of each container
// are merged from "spec" and "status" parts of the manifest.
func (m *PodManifest) GetContainers() map[string]Container {
	containers := make(map[string]Container, 0)
	for _, c := range m.Spec.Containers {
		containers[c.Name] = Container{
			Name:               c.Name,
			IsInitContainer:    false,
			IsSidecarContainer: false,
		}
	}

	for _, ic := range m.Spec.InitContainers {
		containers[ic.Name] = Container{
			Name:               ic.Name,
			IsInitContainer:    true,
			IsSidecarContainer: ic.RestartPolicy == "Always",
		}
	}

	m.fillContainerImages(containers)
	m.Status.fillStates(containers)
	return containers
}

func (m *PodManifest) fillContainerImages(containers map[string]Container) {
	processSpec := func(sc specContainer) {
		if c, ok := containers[sc.Name]; ok {
			if name, tag, err := parseImageNameAndTag(sc.Image); err == nil {
				c.Image.Name = name
				c.Image.Tag = tag
				containers[sc.Name] = c
			}
		}
	}

	processStatus := func(sc statusContainer) {
		if c, ok := containers[sc.Name]; ok {
			c.Image.ImageID = sc.ImageID
			containers[sc.Name] = c
		}
	}

	for _, c := range m.Spec.Containers {
		processSpec(c)
	}

	for _, c := range m.Spec.InitContainers {
		processSpec(c.specContainer)
	}

	for _, c := range m.Status.ContainerStatuses {
		processStatus(c)
	}

	for _, c := range m.Status.InitContainerStatuses {
		processStatus(c)
	}
}

// fillStates fills the basic and init container states from the "status" part of the manifest.
func (s *PodStatus) fillStates(containers map[string]Container) {
	for _, c := range s.ContainerStatuses {
		c.fillContainer(containers)
	}

	for _, ic := range s.InitContainerStatuses {
		ic.fillContainer(containers)
	}
}

// fillContainer fills the container with additional information from "status" part of manifest and
// updates the container in the containers map.
func (sc *statusContainer) fillContainer(containers map[string]Container) {
	c, ok := containers[sc.Name]
	if !ok {
		return
	}

	c.ContainerId = sc.ContainerId
	c.State = getState(sc.State)
	containers[sc.Name] = c
}

// getState parse the state of the container from the "state" part of the manifest.
// The state is the processor is looking for is the key in the map. The value of status key
// is ignored.
func getState(state map[string]any) string {
	for key := range state {
		return key
	}
	return ""
}

func parseImageNameAndTag(image string) (name, tag string, err error) {
	ref, err := reference.Parse(image)
	if err != nil {
		return "", "", err
	}

	if namedTagged, ok := ref.(reference.NamedTagged); ok {
		return namedTagged.Name(), namedTagged.Tag(), nil
	} else if namedOnly, ok := ref.(reference.Named); ok {
		return namedOnly.Name(), "latest", nil
	} else {
		return "", "", errors.New("cannot retrieve image name and tag")
	}
}
