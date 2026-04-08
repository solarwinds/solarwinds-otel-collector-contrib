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
	godigest "github.com/opencontainers/go-digest"
	"go.uber.org/zap"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/constants"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/manifests"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	conventions "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	k8sLogType                              = "sw.k8s.log.type"
	k8sContainerEntityType                  = "KubernetesContainer"
	k8sContainerImageEntityType             = "KubernetesContainerImage"
	KubernetesResourceUsesImageRelationType = "KubernetesResourceUsesImage"
	entityState                             = "entity_state"
	relationshipUpdateEventType             = "entity_relationship_state"

	containerImageEntityType = "ContainerImage"
	relatesToRelationType    = "RelatesTo"

	otelEntityEventAsLog = "otel.entity.event_as_log"
	otelEntityEventType  = "otel.entity.event.type"
	swEntityType         = "otel.entity.type"

	otelEntityId    = "otel.entity.id"
	swK8sClusterUid = "sw.k8s.cluster.uid"
	swK8sWorkloadIp = "sw.k8s.workload.ip"
	serviceName     = "k8s.service.name"

	relationshipSrcEntityIds  = "otel.entity_relationship.source_entity.id"
	relationshipDestEntityIds = "otel.entity_relationship.destination_entity.id"
	relationshipType          = "otel.entity_relationship.type"
	srcEntityType             = "otel.entity_relationship.source_entity.type"
	destEntityType            = "otel.entity_relationship.destination_entity.type"

	otelEntityAttributes               = "otel.entity.attributes"
	k8sContainerStatus                 = "sw.k8s.container.status"
	k8sContainerInit                   = "sw.k8s.container.init"
	k8sContainerSidecar                = "sw.k8s.container.sidecar"
	k8sContainerDeployedByK8sCollector = "sw.k8s.deployedbycollector"
)

// imageIdentity uniquely identifies a container image based on digest and name only,
// excluding tag to prevent duplicate entities for the same image with different tags
type imageIdentity struct {
	digest string
	name   string
}

// addEntityStateEventResourceLog adds a new ResourceLogs to the provided Logs structure
// and sets required attributes on "resource" and "scopeLogs"
func addEntityStateEventResourceLog(ld plog.Logs, containersLogSlice plog.LogRecordSlice) {
	rl := ld.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr(k8sLogType, "entitystateevent")
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().Attributes().PutBool(otelEntityEventAsLog, true)
	containersLogSlice.CopyTo(sl.LogRecords())
}

// transformContainersToContainerLogs returns a new [plog.LogRecordSlice] and appends
// all LogRecords containing container information from the provided Containers.
func transformContainersToContainerLogs(containers map[string]manifests.Container, md manifests.PodMetadata, t pcommon.Timestamp, clusterUID string) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	for _, c := range containers {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		addContainerAttributes(lr.Attributes(), md, c, clusterUID)
	}

	return lrs
}

// addContainerAttributes sets attributes on the provided map for the given Metadata and Container.
func addContainerAttributes(attrs pcommon.Map, md manifests.PodMetadata, c manifests.Container, clusterUID string) {
	attrs.PutStr(otelEntityEventType, entityState)
	attrs.PutStr(swEntityType, k8sContainerEntityType)

	tm := attrs.PutEmptyMap(otelEntityId)
	tm.PutStr(string(conventions.K8SPodNameKey), md.Name)
	tm.PutStr(string(conventions.K8SNamespaceNameKey), md.Namespace)
	tm.PutStr(string(conventions.K8SContainerNameKey), c.Name)
	tm.PutStr(swK8sClusterUid, clusterUID)

	ea := attrs.PutEmptyMap(otelEntityAttributes)
	ea.PutStr(string(conventions.ContainerIDKey), c.ContainerId)
	ea.PutStr(k8sContainerStatus, c.State)
	ea.PutBool(k8sContainerInit, c.IsInitContainer)
	ea.PutBool(k8sContainerSidecar, c.IsSidecarContainer)

	if c.IsDeployedByK8sCollector {
		ea.PutBool(k8sContainerDeployedByK8sCollector, c.IsDeployedByK8sCollector)
	}
}

// addServiceMappingsResourceLog adds a new ResourceLogs to the provided Logs structure
// and sets required attributes on "resource" and "scopeLogs"
func addServiceMappingsResourceLog(ld plog.Logs, serviceMappingsLogSlice plog.LogRecordSlice) {
	rl := ld.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr(k8sLogType, "serviceendpointsmapping")
	lrs := rl.ScopeLogs().AppendEmpty().LogRecords()
	serviceMappingsLogSlice.CopyTo(lrs)
}

func transformManifestToServiceMappingLogs(m manifests.ServiceMapping, t pcommon.Timestamp, clusterUID string) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	for _, addr := range m.GetAddresses() {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		attrs := lr.Attributes()
		attrs.PutStr(serviceName, m.GetServiceName())
		attrs.PutStr(string(conventions.K8SNamespaceNameKey), m.GetNamespace())
		attrs.PutStr(swK8sWorkloadIp, addr)
		attrs.PutStr(swK8sClusterUid, clusterUID)
	}

	return lrs
}

// transformContainersToContainerImageLogs returns a new [plog.LogRecordSlice] and appends
// all LogRecords containing container image information from the provided Containers.
func transformContainersToContainerImageLogs(containers map[string]manifests.Container, t pcommon.Timestamp, clusterUID string) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	if clusterUID == "" {
		return lrs
	}

	processedValidImages := make(map[imageIdentity]struct{}, len(containers))
	for _, c := range containers {
		if c.Image.ImageID == "" {
			continue
		}
		identity := imageIdentity{
			digest: extractSha256Digest(c.Image.ImageID),
			name:   c.Image.Name,
		}
		if _, seen := processedValidImages[identity]; seen {
			continue
		}
		processedValidImages[identity] = struct{}{}

		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		addContainerImageAttributes(lr.Attributes(), c.Image, clusterUID)
	}

	return lrs
}

// addContainerImageAttributes sets attributes on the provided map for the given Metadata and Image.
func addContainerImageAttributes(attrs pcommon.Map, i manifests.Image, clusterUID string) {
	attrs.PutStr(otelEntityEventType, entityState)
	attrs.PutStr(swEntityType, k8sContainerImageEntityType)

	tm := attrs.PutEmptyMap(otelEntityId)
	tm.PutStr(swK8sClusterUid, clusterUID)
	tm.PutStr(constants.AttributeOciManifestDigest, extractSha256Digest(i.ImageID))
	tm.PutStr(string(conventions.ContainerImageNameKey), i.Name)

	entityAttrs := attrs.PutEmptyMap(otelEntityAttributes)
	tags := entityAttrs.PutEmptySlice(constants.AttributeContainerImageTags)
	if i.Tag != "" {
		tags.AppendEmpty().SetStr(i.Tag)
	}
}

// transformContainersToContainerImageRelationsLogs returns a new [plog.LogRecordSlice] and appends
// all LogRecords containing information about relations between containers and images from the provided Containers.
func transformContainersToContainerImageRelationsLogs(containers map[string]manifests.Container, md manifests.PodMetadata, t pcommon.Timestamp, clusterUID string) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	if clusterUID == "" {
		return lrs
	}

	for _, c := range containers {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		addContainerImageRelationAttributes(lr.Attributes(), md, c, clusterUID)
	}

	return lrs
}

// addContainerImageRelationAttributes sets attributes on the provided map for the given Metadata and Container.
func addContainerImageRelationAttributes(attrs pcommon.Map, md manifests.PodMetadata, c manifests.Container, clusterUID string) {
	attrs.PutStr(otelEntityEventType, relationshipUpdateEventType)
	attrs.PutStr(relationshipType, KubernetesResourceUsesImageRelationType)

	srcIds := attrs.PutEmptyMap(relationshipSrcEntityIds)
	destIds := attrs.PutEmptyMap(relationshipDestEntityIds)

	attrs.PutStr(srcEntityType, k8sContainerEntityType)
	srcIds.PutStr(string(conventions.K8SPodNameKey), md.Name)
	srcIds.PutStr(string(conventions.K8SNamespaceNameKey), md.Namespace)
	srcIds.PutStr(string(conventions.K8SContainerNameKey), c.Name)
	srcIds.PutStr(swK8sClusterUid, clusterUID)

	attrs.PutStr(destEntityType, k8sContainerImageEntityType)
	destIds.PutStr(swK8sClusterUid, clusterUID)
	destIds.PutStr(constants.AttributeOciManifestDigest, extractSha256Digest(c.Image.ImageID))
	destIds.PutStr(string(conventions.ContainerImageNameKey), c.Image.Name)

	relAttrs := attrs.PutEmptyMap(constants.AttributeOtelEntityRelationshipAttributes)
	relAttrs.PutStr(constants.AttributeImageTag, c.Image.Tag)
}

// isValidDigest validates that a digest string is a valid OCI digest
// (e.g., "sha256:<64 hex chars>" or "sha512:<128 hex chars>"),
// consistent with the output of extractSha256Digest.
func isValidDigest(digest string) bool {
	return godigest.Digest(digest).Validate() == nil
}

// transformContainersToContainerImageEntityLogs emits ContainerImage entity state events
// for each unique image digest found in the provided containers.
// Deduplication is per-pod (no shared state across pods).
func transformContainersToContainerImageEntityLogs(containers map[string]manifests.Container, t pcommon.Timestamp, logger *zap.Logger) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()
	seenDigests := make(map[string]struct{}, len(containers))

	for _, c := range containers {
		if c.Image.ImageID == "" {
			logger.Warn("skipping container with empty ImageID",
				zap.String("container", c.Name),
			)
			continue
		}

		digest := extractSha256Digest(c.Image.ImageID)
		if digest == "" || !isValidDigest(digest) {
			logger.Warn("skipping container with malformed ImageID",
				zap.String("container", c.Name),
				zap.String("imageID", c.Image.ImageID),
			)
			continue
		}

		if _, seen := seenDigests[digest]; seen {
			continue
		}
		seenDigests[digest] = struct{}{}

		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		attrs := lr.Attributes()
		attrs.PutStr(otelEntityEventType, entityState)
		attrs.PutStr(swEntityType, containerImageEntityType)

		ids := attrs.PutEmptyMap(otelEntityId)
		ids.PutStr(constants.AttributeOciManifestDigest, digest)
	}

	return lrs
}

// transformContainersToContainerImageRelatesToLogs emits RelatesTo relationship
// state events linking ContainerImage (source) to KubernetesContainerImage (destination)
// for each unique {digest, name} pair in the provided containers.
func transformContainersToContainerImageRelatesToLogs(containers map[string]manifests.Container, t pcommon.Timestamp, clusterUID string, logger *zap.Logger) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	if clusterUID == "" {
		return lrs
	}

	seenIdentities := make(map[imageIdentity]struct{}, len(containers))

	for _, c := range containers {
		if c.Image.ImageID == "" {
			logger.Warn("skipping container with empty ImageID for RelatesTo relationship",
				zap.String("container", c.Name),
			)
			continue
		}

		digest := extractSha256Digest(c.Image.ImageID)
		if digest == "" || !isValidDigest(digest) {
			logger.Warn("skipping container with malformed ImageID for RelatesTo relationship",
				zap.String("container", c.Name),
				zap.String("imageID", c.Image.ImageID),
			)
			continue
		}

		identity := imageIdentity{digest: digest, name: c.Image.Name}
		if _, seen := seenIdentities[identity]; seen {
			continue
		}
		seenIdentities[identity] = struct{}{}

		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		attrs := lr.Attributes()
		attrs.PutStr(otelEntityEventType, relationshipUpdateEventType)
		attrs.PutStr(relationshipType, relatesToRelationType)

		// Source: ContainerImage identified by digest only
		attrs.PutStr(srcEntityType, containerImageEntityType)
		srcIds := attrs.PutEmptyMap(relationshipSrcEntityIds)
		srcIds.PutStr(constants.AttributeOciManifestDigest, digest)

		// Destination: KubernetesContainerImage identified by {clusterUid, digest, name}
		attrs.PutStr(destEntityType, k8sContainerImageEntityType)
		destIds := attrs.PutEmptyMap(relationshipDestEntityIds)
		destIds.PutStr(swK8sClusterUid, clusterUID)
		destIds.PutStr(constants.AttributeOciManifestDigest, digest)
		destIds.PutStr(string(conventions.ContainerImageNameKey), c.Image.Name)
	}

	return lrs
}
