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
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	manifest = PodManifest{
		Metadata: PodMetadata{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
		Status: PodStatus{
			ContainerStatuses: []statusContainer{
				{
					Name:        "test-container",
					ContainerId: "test-container-id",
					State: map[string]any{
						"running": map[string]any{
							"startedAt": "2021-01-01T00:00:00Z",
						},
					},
					ImageID: "busybox@sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
				},
				{
					Name:        "test-container-missing-in-spec",
					ContainerId: "test-container-missing-in-spec-id",
					State: map[string]any{
						"running": map[string]any{
							"startedAt": "2021-01-01T00:00:00Z",
						},
					},
					ImageID: "busybox@sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
				},
			},
			InitContainerStatuses: []statusContainer{
				{
					Name:        "test-init-container",
					ContainerId: "test-init-container-id",
					State: map[string]any{
						"waiting": map[string]any{},
					},
					ImageID: "init-busybox@sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
				},
				{
					Name:        "test-sidecar-container",
					ContainerId: "test-sidecar-container-id",
					State: map[string]any{
						"terminated": map[string]any{},
					},
					ImageID: "sidecar-busybox@sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
				},
			},
		},
		Spec: PodSpec{
			Containers: []specContainer{
				{
					Name:  "test-container",
					Image: "busybox:latest",
				},
				{
					Name:  "test-container-missing-in-status",
					Image: "busybox:latest",
				},
			},
			InitContainers: []struct {
				specContainer
				RestartPolicy string `json:"restartPolicy"`
			}{
				{
					specContainer: specContainer{
						Name:  "test-init-container",
						Image: "init-busybox:latest",
					},
					RestartPolicy: "Smth",
				},
				{
					specContainer: specContainer{
						Name:  "test-sidecar-container",
						Image: "sidecar-busybox:latest",
					},
					RestartPolicy: "Always",
				},
			},
		},
	}
)

func TestGetContainer(t *testing.T) {
	containers := manifest.GetContainers()

	// container missing in spec should not be returned in the result
	assert.Len(t, containers, 4, "Expected 4 containers")

	// Basic container
	container, ok := containers["test-container"]
	assert.Truef(t, ok, "Expected container not found")
	assert.Equal(t, "test-container", container.Name)
	expectedContainer := Container{
		Name:               "test-container",
		ContainerId:        "test-container-id",
		State:              "running",
		IsInitContainer:    false,
		IsSidecarContainer: false,
		Image: Image{
			ImageID: "busybox@sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			Name:    "index.docker.io/library/busybox",
			Tag:     "latest",
		},
	}

	assert.Equal(t, expectedContainer, container)

	// init container
	initContainer, ok := containers["test-init-container"]
	assert.Truef(t, ok, "Expected container not found")
	assert.Equal(t, "test-init-container", initContainer.Name)
	expectedInitContainer := Container{
		Name:               "test-init-container",
		ContainerId:        "test-init-container-id",
		State:              "waiting",
		IsInitContainer:    true,
		IsSidecarContainer: false,
		Image: Image{
			ImageID: "init-busybox@sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			Name:    "index.docker.io/library/init-busybox",
			Tag:     "latest",
		},
	}

	assert.Equal(t, expectedInitContainer, initContainer)

	// sidecar container
	sidecarContainer, ok := containers["test-sidecar-container"]
	assert.Truef(t, ok, "Expected container not found")
	assert.Equal(t, "test-sidecar-container", sidecarContainer.Name)
	expectedSidecarContainer := Container{
		Name:               "test-sidecar-container",
		ContainerId:        "test-sidecar-container-id",
		State:              "terminated",
		IsInitContainer:    true,
		IsSidecarContainer: true,
		Image: Image{
			ImageID: "sidecar-busybox@sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			Name:    "index.docker.io/library/sidecar-busybox",
			Tag:     "latest",
		},
	}
	assert.Equal(t, expectedSidecarContainer, sidecarContainer)

	// container missing in status part of the manifest should be returned
	specOnlyContainer, ok := containers["test-container-missing-in-status"]
	assert.Truef(t, ok, "Expected container not found")
	assert.Equal(t, "test-container-missing-in-status", specOnlyContainer.Name)
	expectedSpecOnlyContainer := Container{
		Name:               "test-container-missing-in-status",
		IsInitContainer:    false,
		IsSidecarContainer: false,
		Image: Image{
			ImageID: "",
			Name:    "index.docker.io/library/busybox",
			Tag:     "latest",
		},
	}

	assert.Equal(t, expectedSpecOnlyContainer, specOnlyContainer)
}

func TestParseImageNameAndTag(t *testing.T) {
	tests := []struct {
		name          string
		image         string
		expectedName  string
		expectedTag   string
		expectedError bool
	}{
		{
			name:          "Docker Hub official image (implicit library prefix)",
			image:         "nginx",
			expectedName:  "index.docker.io/library/nginx",
			expectedTag:   "latest",
			expectedError: false,
		},
		{
			name:          "Docker Hub official image with tag",
			image:         "nginx:1.21",
			expectedName:  "index.docker.io/library/nginx",
			expectedTag:   "1.21",
			expectedError: false,
		},
		{
			name:          "Docker Hub user image",
			image:         "myuser/myimage:v1.0",
			expectedName:  "index.docker.io/myuser/myimage",
			expectedTag:   "v1.0",
			expectedError: false,
		},
		{
			name:          "Custom registry with tag",
			image:         "myregistry.io/myimage:v2.0",
			expectedName:  "myregistry.io/myimage",
			expectedTag:   "v2.0",
			expectedError: false,
		},
		{
			name:          "Custom registry without tag (defaults to latest)",
			image:         "gcr.io/project/image",
			expectedName:  "gcr.io/project/image",
			expectedTag:   "latest",
			expectedError: false,
		},
		{
			name:          "Docker Hub library prefix (Trivy format)",
			image:         "library/python",
			expectedName:  "index.docker.io/library/python",
			expectedTag:   "latest",
			expectedError: false,
		},
		{
			name:          "Docker Hub library prefix with tag",
			image:         "library/python:3.9",
			expectedName:  "index.docker.io/library/python",
			expectedTag:   "3.9",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, tag, err := parseImageNameAndTag(tt.image)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, name, "Image name mismatch")
				assert.Equal(t, tt.expectedTag, tag, "Image tag mismatch")
			}
		})
	}
}
