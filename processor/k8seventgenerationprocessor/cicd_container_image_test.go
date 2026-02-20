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
	"time"

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

func cicdTestTimestamp() pcommon.Timestamp {
	return pcommon.NewTimestampFromTime(time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC))
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

// TestTransformContainersToCICDContainerImageLogs
//
// Verifies that transformContainersToCICDContainerImageLogs emits ContainerImage
// entity state log records with the expected attributes and timestamps.

func TestTransformContainersToCICDContainerImageLogs(t *testing.T) {
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
		ts := cicdTestTimestamp()

		result := transformContainersToCICDContainerImageLogs(containers, ts, zap.NewNop())

		require.Equal(t, 1, result.Len())
		lr := result.At(0)
		assert.Equal(t, ts, lr.ObservedTimestamp())

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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), zap.NewNop())

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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), zap.NewNop())

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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), zap.NewNop())

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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), logger)

		assert.Equal(t, 0, result.Len(), "malformed ImageID should not produce entity")
		require.GreaterOrEqual(t, recorded.Len(), 1, "should log a warning for malformed ImageID")

		// Verify warning includes required diagnostic fields (container name and invalid ImageID)
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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), zap.NewNop())

		require.Equal(t, 1, result.Len())
		ids := getMapValue(t, result.At(0).Attributes(), otelEntityId)

		// ContainerImage entity has ONLY oci.manifest.digest as ID (single-attribute identity)
		assert.Equal(t, 1, ids.Len(), "ContainerImage entity ID should contain only oci.manifest.digest")
		assert.Equal(t, validDigest, getStringValue(t, ids, constants.AttributeOciManifestDigest))
	})
}

// TestTransformContainersToCICDContainerImageLogs_ContainerTypes verifies
// init containers and main containers are processed. Ephemeral containers
// are excluded at PodManifest.GetContainers() level, so the transform
// function never sees them.

func TestTransformContainersToCICDContainerImageLogs_ContainerTypes(t *testing.T) {
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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), zap.NewNop())

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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), zap.NewNop())

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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), zap.NewNop())

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

		result := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), zap.NewNop())

		assert.Equal(t, 1, result.Len(), "only containers in the map are processed")
	})
}

// TestTransformContainersToCICDContainerImageRelatesToLogs verifies that
// transformContainersToCICDContainerImageRelatesToLogs emits RelatesTo
// relationship state events with the expected source/destination entity
// types, IDs, and deduplication behavior.

func TestTransformContainersToCICDContainerImageRelatesToLogs(t *testing.T) {
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
		ts := cicdTestTimestamp()

		result := transformContainersToCICDContainerImageRelatesToLogs(containers, ts, zap.NewNop())

		require.Equal(t, 1, result.Len())
		lr := result.At(0)
		assert.Equal(t, ts, lr.ObservedTimestamp())

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

		result := transformContainersToCICDContainerImageRelatesToLogs(containers, cicdTestTimestamp(), zap.NewNop())

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

		result := transformContainersToCICDContainerImageRelatesToLogs(containers, cicdTestTimestamp(), zap.NewNop())

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

		result := transformContainersToCICDContainerImageRelatesToLogs(containers, cicdTestTimestamp(), zap.NewNop())

		assert.Equal(t, 0, result.Len(), "container with empty ImageID should be skipped")
	})

	t.Run("correct source and destination entity types and IDs", func(t *testing.T) {
		containers := map[string]manifests.Container{
			"nginx": {
				Name:  "nginx",
				Image: manifests.Image{ImageID: validImageID, Name: "docker.io/library/nginx", Tag: "1.25"},
			},
		}

		result := transformContainersToCICDContainerImageRelatesToLogs(containers, cicdTestTimestamp(), zap.NewNop())

		require.Equal(t, 1, result.Len())
		attrs := result.At(0).Attributes()

		// Source: ContainerImage with {oci.manifest.digest}
		assert.Equal(t, "ContainerImage", getStringValue(t, attrs, srcEntityType))
		srcIds := getMapValue(t, attrs, relationshipSrcEntityIds)
		assert.Equal(t, 1, srcIds.Len(), "source entity should have only oci.manifest.digest")
		assert.Equal(t, validDigest, getStringValue(t, srcIds, constants.AttributeOciManifestDigest))

		// Destination: KubernetesContainerImage with {oci.manifest.digest, container.image.name}
		assert.Equal(t, k8sContainerImageEntityType, getStringValue(t, attrs, destEntityType))
		destIds := getMapValue(t, attrs, relationshipDestEntityIds)
		assert.Equal(t, 2, destIds.Len(), "destination entity should have oci.manifest.digest and container.image.name")
		assert.Equal(t, validDigest, getStringValue(t, destIds, constants.AttributeOciManifestDigest))
		assert.Equal(t, "docker.io/library/nginx", getStringValue(t, destIds, "container.image.name"))
	})
}

// TestTransformContainersToCICDPartialFailure verifies best-effort behavior:
// valid containers produce events even when other containers in the same pod
// have malformed or empty ImageIDs.

func TestTransformContainersToCICDPartialFailure(t *testing.T) {
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

		// ContainerImage entity emission: only good-container produces entity
		entityResult := transformContainersToCICDContainerImageLogs(containers, cicdTestTimestamp(), logger)

		assert.Equal(t, 1, entityResult.Len(), "only valid container should emit entity")
		ids := getMapValue(t, entityResult.At(0).Attributes(), otelEntityId)
		assert.Equal(t, validDigest, getStringValue(t, ids, constants.AttributeOciManifestDigest))

		// RelatesTo relationship emission: only good-container produces relationship
		relResult := transformContainersToCICDContainerImageRelatesToLogs(containers, cicdTestTimestamp(), logger)

		assert.Equal(t, 1, relResult.Len(), "only valid container should emit relationship")

		// Warning logged for malformed ImageID (empty ImageID is silently skipped, matching existing behavior).
		// Note: The spec also mentions logging errors and emitting metrics for actual emission failures
		// (memory pressure, serialization errors). Those failure modes require mocking plog internals
		// and are not testable at the unit level — they will be validated via integration/deployment tests.
		require.GreaterOrEqual(t, recorded.Len(), 1, "should log warning for malformed ImageID")

		// Verify warning log contents include required diagnostic fields
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
