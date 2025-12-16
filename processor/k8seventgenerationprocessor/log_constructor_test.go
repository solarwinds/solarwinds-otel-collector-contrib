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

package k8seventgenerationprocessor

import (
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/constants"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/manifests"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestExtractSha256Digest(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Full digest with sha256 prefix",
			input:    "sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			expected: "sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
		},
		{
			name:     "Docker image reference with sha256",
			input:    "busybox@sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			expected: "sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
		},
		{
			name:     "Docker-pullable format",
			input:    "docker-pullable://nginx@sha256:abc123def456",
			expected: "sha256:abc123def456",
		},
		{
			name:     "SHA512 algorithm",
			input:    "nginx@sha512:deadbeef12345678",
			expected: "sha512:deadbeef12345678",
		},
		{
			name:     "Uppercase SHA256",
			input:    "SHA256:abc123def456",
			expected: "sha256:abc123def456",
		},
		{
			name:     "Mixed case algorithm",
			input:    "ShA256:abc123def456",
			expected: "sha256:abc123def456",
		},
		{
			name:     "Uppercase SHA512",
			input:    "SHA512:deadbeef",
			expected: "sha512:deadbeef",
		},
		{
			name:     "Plain hash without prefix",
			input:    "355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			expected: "sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
		},
		{
			name:     "Docker image with uppercase algorithm",
			input:    "myimage@SHA256:abc123",
			expected: "sha256:abc123",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Real-world Kubernetes imageID format",
			input:    "docker.io/library/nginx@sha256:4c0fdaa8b6341bfdeca5f18f7837462c80cff90527ee35ef185571e1c327beac",
			expected: "sha256:4c0fdaa8b6341bfdeca5f18f7837462c80cff90527ee35ef185571e1c327beac",
		},
		{
			name:     "GCR image with digest",
			input:    "gcr.io/project/image@sha256:1234567890abcdef",
			expected: "sha256:1234567890abcdef",
		},
		{
			name:     "SHA384 algorithm",
			input:    "sha384:abcdef123456",
			expected: "sha384:abcdef123456",
		},
		{
			name:     "Image with port and digest",
			input:    "myregistry.io:5000/myimage@sha256:fedcba987654",
			expected: "sha256:fedcba987654",
		},
		{
			name:     "Docker runtime prefix",
			input:    "docker://sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			expected: "sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
		},
		{
			name:     "Containerd runtime prefix",
			input:    "containerd://sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			expected: "sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
		},
		{
			name:     "CRI-O runtime prefix",
			input:    "cri-o://sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
			expected: "sha256:355b3a1bf5609da364166913878a8508d4ba30572d02020a97028c75477e24ff",
		},
		{
			name:     "Docker runtime with image reference",
			input:    "docker://nginx@sha256:abc123",
			expected: "sha256:abc123",
		},
		{
			name:     "Containerd runtime with image reference",
			input:    "containerd://docker.io/library/nginx@sha256:4c0fdaa8b6341bfdeca5f18f7837462c80cff90527ee35ef185571e1c327beac",
			expected: "sha256:4c0fdaa8b6341bfdeca5f18f7837462c80cff90527ee35ef185571e1c327beac",
		},
		{
			name:     "Docker-pullable with image reference and digest",
			input:    "docker-pullable://gcr.io/project/image@sha256:1234567890abcdef",
			expected: "sha256:1234567890abcdef",
		},
		{
			name:     "CRI-O runtime with full image path",
			input:    "cri-o://myregistry.io:5000/myimage@sha256:fedcba987654",
			expected: "sha256:fedcba987654",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSha256Digest(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddContainerImageRelationAttributes(t *testing.T) {
	tests := []struct {
		name        string
		container   manifests.Container
		expectedTag string
	}{
		{
			name: "Container with tag",
			container: manifests.Container{
				Name: "test-container",
				Image: manifests.Image{
					ImageID: "sha256:abc123",
					Name:    "nginx",
					Tag:     "v1.2.3",
				},
			},
			expectedTag: "v1.2.3",
		},
		{
			name: "Container with latest tag",
			container: manifests.Container{
				Name: "test-container",
				Image: manifests.Image{
					ImageID: "sha256:def456",
					Name:    "redis",
					Tag:     "latest",
				},
			},
			expectedTag: "latest",
		},
		{
			name: "Container with empty tag",
			container: manifests.Container{
				Name: "test-container",
				Image: manifests.Image{
					ImageID: "sha256:ghi789",
					Name:    "busybox",
					Tag:     "",
				},
			},
			expectedTag: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := pcommon.NewMap()
			md := manifests.PodMetadata{
				Name:      "test-pod",
				Namespace: "test-namespace",
			}
			clusterUID := "test-cluster-uid"

			addContainerImageRelationAttributes(attrs, md, tt.container, clusterUID)

			// Verify relationship type
			relType, exists := attrs.Get(relationshipType)
			assert.True(t, exists, "relationship type should exist")
			assert.Equal(t, KubernetesResourceUsesImageRelationType, relType.Str())

			// Verify relationship attributes map exists
			relAttrs, exists := attrs.Get(constants.AttributeOtelEntityRelationshipAttributes)
			assert.True(t, exists, "relationship attributes should exist")
			assert.Equal(t, pcommon.ValueTypeMap, relAttrs.Type(), "relationship attributes should be a map")

			// Verify imageTag is present in relationship attributes
			imageTag, exists := relAttrs.Map().Get(constants.AttributeImageTag)
			assert.True(t, exists, "imageTag should exist in relationship attributes")
			assert.Equal(t, tt.expectedTag, imageTag.Str(), "imageTag value should match container image tag")
		})
	}
}
