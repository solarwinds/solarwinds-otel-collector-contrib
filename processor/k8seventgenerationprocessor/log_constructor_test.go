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
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
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

			relType, exists := attrs.Get(relationshipType)
			assert.True(t, exists, "relationship type should exist")
			assert.Equal(t, KubernetesResourceUsesImageRelationType, relType.Str())

			relAttrs, exists := attrs.Get(constants.AttributeOtelEntityRelationshipAttributes)
			assert.True(t, exists, "relationship attributes should exist")
			assert.Equal(t, pcommon.ValueTypeMap, relAttrs.Type(), "relationship attributes should be a map")

			imageTag, exists := relAttrs.Map().Get(constants.AttributeImageTag)
			assert.True(t, exists, "imageTag should exist in relationship attributes")
			assert.Equal(t, tt.expectedTag, imageTag.Str(), "imageTag value should match container image tag")
		})
	}
}

// collectEntityDigests extracts oci.manifest.digest values from ContainerImage entity log records.
func collectEntityDigests(t *testing.T, lrs plog.LogRecordSlice) []string {
	t.Helper()
	var digests []string
	for i := 0; i < lrs.Len(); i++ {
		ids := getMapValue(t, lrs.At(i).Attributes(), otelEntityId)
		digests = append(digests, getStringValue(t, ids, constants.AttributeOciManifestDigest))
	}
	return digests
}

func TestTransformContainersToContainerImageEntityLogs(t *testing.T) {
	const (
		validImageID  = "docker://sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		validDigest   = "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		validImageID2 = "containerd://sha256:b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"
		validDigest2  = "sha256:b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"
	)

	t.Run("single container emits one ContainerImage entity", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name: "nginx",
				Image: manifests.Image{
					ImageID: validImageID,
					Name:    "docker.io/library/nginx",
					Tag:     "1.25",
				},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		require.Equal(t, 1, result.Len())
		lr := result.At(0)
		assert.Equal(t, timestamp, lr.ObservedTimestamp())

		attrs := lr.Attributes()
		assert.Equal(t, entityState, getStringValue(t, attrs, otelEntityEventType))
		assert.Equal(t, "ContainerImage", getStringValue(t, attrs, swEntityType))
	})

	t.Run("dedup by digest within pod", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx-a": {
				Name:  "nginx-a",
				Image: manifests.Image{ImageID: validImageID, Name: "registry-a/nginx", Tag: "1.25"},
			},
			"nginx-b": {
				Name:  "nginx-b",
				Image: manifests.Image{ImageID: validImageID, Name: "registry-b/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		assert.Equal(t, 1, result.Len(), "same digest should emit only one ContainerImage entity")
		ids := getMapValue(t, result.At(0).Attributes(), otelEntityId)
		assert.Equal(t, validDigest, getStringValue(t, ids, constants.AttributeOciManifestDigest))
	})

	t.Run("distinct digests emit one entity per digest", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "nginx", Tag: "1.25"},
			},
			"redis": {
				Name:  "redis",
				Image: manifests.Image{ImageID: validImageID2, Name: "redis", Tag: "7.0"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		assert.Equal(t, 2, result.Len())
		digests := collectEntityDigests(t, result)
		assert.Contains(t, digests, validDigest)
		assert.Contains(t, digests, validDigest2)
	})

	t.Run("empty ImageID skipped", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"waiting": {
				Name:  "waiting",
				Image: manifests.Image{ImageID: "", Name: "nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		assert.Equal(t, 0, result.Len(), "container with empty ImageID should be skipped")
	})

	t.Run("malformed ImageID skipped with warning", func(t *testing.T) {
		core, recorded := observer.New(zapcore.WarnLevel)
		logger := zap.New(core)

		containers := map[string]manifests.Container{
			"bad-container": {
				Name:  "bad-container",
				Image: manifests.Image{ImageID: "not-a-valid-digest-format", Name: "nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, logger)

		assert.Equal(t, 0, result.Len(), "malformed ImageID should not produce entity")
		require.GreaterOrEqual(t, recorded.Len(), 1, "should log a warning for malformed ImageID")

		entry := recorded.All()[0]
		fieldMap := entry.ContextMap()
		assert.Equal(t, "bad-container", fieldMap["container"], "warning should include container name")
		assert.Equal(t, "not-a-valid-digest-format", fieldMap["imageID"], "warning should include invalid ImageID value")
	})

	t.Run("correct otel.entity.id attributes", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		require.Equal(t, 1, result.Len())
		ids := getMapValue(t, result.At(0).Attributes(), otelEntityId)

		// ContainerImage entity has ONLY oci.manifest.digest as ID (single-attribute identity)
		assert.Equal(t, 1, ids.Len(), "ContainerImage entity ID should contain only oci.manifest.digest")
		assert.Equal(t, validDigest, getStringValue(t, ids, constants.AttributeOciManifestDigest))
	})
}

// TestTransformContainersToContainerImageEntityLogs_ContainerTypes verifies
// init containers and main containers are processed. Ephemeral containers
// are excluded at PodManifest.GetContainers() level, so the transform
// function never sees them.

func TestTransformContainersToContainerImageEntityLogs_ContainerTypes(t *testing.T) {
	const (
		validImageID  = "docker://sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		validImageID2 = "containerd://sha256:b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"
	)

	t.Run("init containers processed", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"init-setup": {
				Name:            "init-setup",
				IsInitContainer: true,
				Image:           manifests.Image{ImageID: validImageID, Name: "busybox", Tag: "latest"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		assert.Equal(t, 1, result.Len(), "init containers should produce ContainerImage entities")
	})

	t.Run("main containers processed", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"app": {
				Name:            "app",
				IsInitContainer: false,
				Image:           manifests.Image{ImageID: validImageID, Name: "my-app", Tag: "v1.0"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		assert.Equal(t, 1, result.Len(), "main containers should produce ContainerImage entities")
	})

	t.Run("mixed init and main containers all processed", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"init-db": {
				Name:            "init-db",
				IsInitContainer: true,
				Image:           manifests.Image{ImageID: validImageID, Name: "db-init", Tag: "v1"},
			},
			"app": {
				Name:            "app",
				IsInitContainer: false,
				Image:           manifests.Image{ImageID: validImageID2, Name: "app", Tag: "v2"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		assert.Equal(t, 2, result.Len(), "both init and main container types produce entities")
	})

	t.Run("ephemeral containers excluded by manifest parser", func(t *testing.T) {
		// Ephemeral containers are filtered out at PodManifest.GetContainers() level
		// (PodSpec only parses spec.containers and spec.initContainers).
		// The transform function processes exactly what it receives — no ephemerals.
		containers := map[string]manifests.Container{
			"main": {
				Name:  "main",
				Image: manifests.Image{ImageID: validImageID, Name: "nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageEntityLogs(containers, timestamp, zap.NewNop())

		assert.Equal(t, 1, result.Len(), "only containers in the map are processed")
	})
}

// TestTransformContainersToContainerImageRelatesToLogs verifies that
// transformContainersToContainerImageRelatesToLogs emits RelatesTo
// relationship state events with the expected source/destination entity
// types, IDs, and deduplication behavior.

func TestTransformContainersToContainerImageRelatesToLogs(t *testing.T) {
	const (
		validImageID = "docker://sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		validDigest  = "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
	)

	t.Run("single container emits one RelatesTo relationship", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageRelatesToLogs(containers, timestamp, "test-cluster-uid", zap.NewNop())

		require.Equal(t, 1, result.Len())
		lr := result.At(0)
		assert.Equal(t, timestamp, lr.ObservedTimestamp())

		attrs := lr.Attributes()
		assert.Equal(t, relationshipUpdateEventType, getStringValue(t, attrs, otelEntityEventType))
		assert.Equal(t, "RelatesTo", getStringValue(t, attrs, relationshipType))
	})

	t.Run("dedup by digest and name within pod", func(t *testing.T) {
		// Two containers sharing the same digest AND same image name → one relationship
		containers := map[string]manifests.Container{
			"main": {
				Name:  "main",
				Image: manifests.Image{ImageID: validImageID, Name: "nginx", Tag: "1.25"},
			},
			"sidecar": {
				Name:  "sidecar",
				Image: manifests.Image{ImageID: validImageID, Name: "nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageRelatesToLogs(containers, timestamp, "test-cluster-uid", zap.NewNop())

		assert.Equal(t, 1, result.Len(), "identical {digest, name} pair should emit one relationship")
	})

	t.Run("same digest different names emits multiple relationships", func(t *testing.T) {
		// Same digest, different image names → two relationships (distinct KubernetesContainerImage destinations)
		containers := map[string]manifests.Container{
			"nginx-a": {
				Name:  "nginx-a",
				Image: manifests.Image{ImageID: validImageID, Name: "registry-a/nginx", Tag: "1.25"},
			},
			"nginx-b": {
				Name:  "nginx-b",
				Image: manifests.Image{ImageID: validImageID, Name: "registry-b/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageRelatesToLogs(containers, timestamp, "test-cluster-uid", zap.NewNop())

		assert.Equal(t, 2, result.Len(), "same digest with different names should emit two relationships")

		destNames := make(map[string]bool)
		for i := 0; i < result.Len(); i++ {
			destIds := getMapValue(t, result.At(i).Attributes(), relationshipDestEntityIds)
			destNames[getStringValue(t, destIds, "container.image.name")] = true
		}
		assert.True(t, destNames["registry-a/nginx"])
		assert.True(t, destNames["registry-b/nginx"])
	})

	t.Run("empty ImageID skipped", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"waiting": {
				Name:  "waiting",
				Image: manifests.Image{ImageID: "", Name: "nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageRelatesToLogs(containers, timestamp, "test-cluster-uid", zap.NewNop())

		assert.Equal(t, 0, result.Len(), "container with empty ImageID should be skipped")
	})

	// T009: supply clusterUID and assert sw.k8s.cluster.uid is present in destination entity IDs.
	t.Run("correct source and destination entity types and IDs with clusterUID", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageRelatesToLogs(containers, timestamp, "test-cluster-uid", zap.NewNop())

		require.Equal(t, 1, result.Len())
		attrs := result.At(0).Attributes()

		// Source: ContainerImage with {oci.manifest.digest}
		assert.Equal(t, "ContainerImage", getStringValue(t, attrs, srcEntityType))
		srcIds := getMapValue(t, attrs, relationshipSrcEntityIds)
		assert.Equal(t, 1, srcIds.Len(), "source entity should have only oci.manifest.digest")
		assert.Equal(t, validDigest, getStringValue(t, srcIds, constants.AttributeOciManifestDigest))

		// Destination: KubernetesContainerImage with {sw.k8s.cluster.uid, oci.manifest.digest, container.image.name}
		assert.Equal(t, k8sContainerImageEntityType, getStringValue(t, attrs, destEntityType))
		destIds := getMapValue(t, attrs, relationshipDestEntityIds)
		assert.Equal(t, 3, destIds.Len(), "destination entity should have sw.k8s.cluster.uid, oci.manifest.digest and container.image.name")
		assert.Equal(t, "test-cluster-uid", getStringValue(t, destIds, "sw.k8s.cluster.uid"),
			"RelatesTo destination ID must include sw.k8s.cluster.uid")
		assert.Equal(t, validDigest, getStringValue(t, destIds, constants.AttributeOciManifestDigest))
		assert.Equal(t, "docker.io/library/nginx", getStringValue(t, destIds, "container.image.name"))
	})

	// T010: skips records when clusterUID is empty.
	t.Run("skips records when clusterUID is empty", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageRelatesToLogs(containers, timestamp, "", zap.NewNop())

		assert.Equal(t, 0, result.Len(),
			"RelatesTo relationship emission must be skipped when clusterUID is empty")
	})
}

// TestTransformContainersToContainerImagePartialFailure verifies best-effort behavior:
// valid containers produce events even when other containers in the same pod
// have malformed or empty ImageIDs.

func TestTransformContainersToContainerImagePartialFailure(t *testing.T) {
	const (
		validImageID = "docker://sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		validDigest  = "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
	)

	t.Run("valid container succeeds while malformed and empty containers are skipped", func(t *testing.T) {
		core, recorded := observer.New(zapcore.WarnLevel)
		logger := zap.New(core)

		containers := map[string]manifests.Container{
			"good-container": {
				Name:  "good-container",
				Image: manifests.Image{ImageID: validImageID, Name: "nginx", Tag: "1.25"},
			},
			"bad-container": {
				Name:  "bad-container",
				Image: manifests.Image{ImageID: "malformed-no-digest", Name: "broken", Tag: "latest"},
			},
			"empty-container": {
				Name:  "empty-container",
				Image: manifests.Image{ImageID: "", Name: "pending", Tag: "v1"},
			},
		}

		entityResult := transformContainersToContainerImageEntityLogs(containers, timestamp, logger)

		assert.Equal(t, 1, entityResult.Len(), "only valid container should emit entity")
		ids := getMapValue(t, entityResult.At(0).Attributes(), otelEntityId)
		assert.Equal(t, validDigest, getStringValue(t, ids, constants.AttributeOciManifestDigest))

		relResult := transformContainersToContainerImageRelatesToLogs(containers, timestamp, "test-cluster-uid", logger)

		assert.Equal(t, 1, relResult.Len(), "only valid container should emit relationship")

		// Empty ImageID is silently skipped; malformed ImageID logs a warning.
		require.GreaterOrEqual(t, recorded.Len(), 1, "should log warning for malformed ImageID")

		foundMalformedWarning := false
		for _, entry := range recorded.All() {
			fieldMap := entry.ContextMap()
			if fieldMap["container"] == "bad-container" {
				foundMalformedWarning = true
				assert.Equal(t, "malformed-no-digest", fieldMap["imageID"],
					"warning should include invalid ImageID value")
				break
			}
		}
		assert.True(t, foundMalformedWarning, "should log a warning with container name 'bad-container' and its ImageID")
	})
}

// TestTransformContainersToContainerImageLogs tests the KubernetesContainerImage entity state event path.
// T005: entity ID map must include sw.k8s.cluster.uid when clusterUID is provided.
// T006: function must skip (return zero records) when clusterUID is empty.
func TestTransformContainersToContainerImageLogs(t *testing.T) {
	const (
		validImageID = "docker://sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		validDigest  = "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		clusterUID   = "test-cluster-uid"
	)

	t.Run("entity ID includes sw.k8s.cluster.uid (T005)", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageLogs(containers, timestamp, clusterUID)

		require.Equal(t, 1, result.Len())
		ids := getMapValue(t, result.At(0).Attributes(), otelEntityId)
		assert.Equal(t, clusterUID, getStringValue(t, ids, "sw.k8s.cluster.uid"),
			"KubernetesContainerImage entity ID must include sw.k8s.cluster.uid")
		assert.Equal(t, validDigest, getStringValue(t, ids, constants.AttributeOciManifestDigest))
		assert.Equal(t, "docker.io/library/nginx", getStringValue(t, ids, "container.image.name"))
	})

	t.Run("skips all records when clusterUID is empty (T006)", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageLogs(containers, timestamp, "")

		assert.Equal(t, 0, result.Len(),
			"KubernetesContainerImage entity emission must be skipped when clusterUID is empty")
	})
}

// TestTransformContainersToContainerImageRelationsLogs tests the KubernetesResourceUsesImage relationship path.
// T007: destination entity ID must include sw.k8s.cluster.uid when clusterUID is provided.
// T008: function must skip records when clusterUID is empty.
func TestTransformContainersToContainerImageRelationsLogs(t *testing.T) {
	const (
		validImageID = "docker://sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		validDigest  = "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
		clusterUID   = "test-cluster-uid"
	)

	md := manifests.PodMetadata{
		Name:      "test-pod",
		Namespace: "test-namespace",
	}

	t.Run("destination entity ID includes sw.k8s.cluster.uid (T007)", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageRelationsLogs(containers, md, timestamp, clusterUID)

		require.Equal(t, 1, result.Len())
		attrs := result.At(0).Attributes()
		destIds := getMapValue(t, attrs, relationshipDestEntityIds)
		assert.Equal(t, clusterUID, getStringValue(t, destIds, "sw.k8s.cluster.uid"),
			"KubernetesResourceUsesImage destination entity ID must include sw.k8s.cluster.uid")
		assert.Equal(t, validDigest, getStringValue(t, destIds, constants.AttributeOciManifestDigest))
		assert.Equal(t, "docker.io/library/nginx", getStringValue(t, destIds, "container.image.name"))
	})

	t.Run("skips records when clusterUID is empty (T008)", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToContainerImageRelationsLogs(containers, md, timestamp, "")

		assert.Equal(t, 0, result.Len(),
			"KubernetesResourceUsesImage relationship emission must be skipped when clusterUID is empty")
	})
}
