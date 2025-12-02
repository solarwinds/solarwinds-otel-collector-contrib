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
	"os"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/manifests"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	conventions "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	k8sLogType                              = "sw.k8s.log.type"
	clusterUidEnv                           = "CLUSTER_UID"
	k8sContainerEntityType                  = "KubernetesContainer"
	k8sContainerImageEntityType             = "KubernetesContainerImage"
	KubernetesResourceUsesImageRelationType = "KubernetesResourceUsesImage"
	entityState                             = "entity_state"
	relationshipUpdateEventType             = "entity_relationship_state"

	// Attributes for OTel entity events identification
	otelEntityEventAsLog = "otel.entity.event_as_log"
	otelEntityEventType  = "otel.entity.event.type"
	swEntityType         = "otel.entity.type"

	// Attributes for telemetry mapping
	otelEntityId    = "otel.entity.id"
	swK8sClusterUid = "sw.k8s.cluster.uid"
	swK8sWorkloadIp = "sw.k8s.workload.ip"
	serviceName     = "k8s.service.name"

	// Relationship properties
	relationshipSrcEntityIds  = "otel.entity_relationship.source_entity.id"
	relationshipDestEntityIds = "otel.entity_relationship.destination_entity.id"
	relationshipType          = "otel.entity_relationship.type"
	srcEntityType             = "otel.entity_relationship.source_entity.type"
	destEntityType            = "otel.entity_relationship.destination_entity.type"

	// Attributes containing additional information about container
	otelEntityAttributes = "otel.entity.attributes"
	k8sContainerStatus   = "sw.k8s.container.status"
	k8sContainerInit     = "sw.k8s.container.init"
	k8sContainerSidecar  = "sw.k8s.container.sidecar"
)

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
func transformContainersToContainerLogs(containers map[string]manifests.Container, md manifests.PodMetadata, t pcommon.Timestamp) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	for _, c := range containers {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		addContainerAttributes(lr.Attributes(), md, c)
	}

	return lrs
}

// addContainerAttributes sets attributes on the provided map for the given Metadata and Container.
func addContainerAttributes(attrs pcommon.Map, md manifests.PodMetadata, c manifests.Container) {
	// Ingestion attributes
	attrs.PutStr(otelEntityEventType, entityState)
	attrs.PutStr(swEntityType, k8sContainerEntityType)

	// Telemetry mappings
	tm := attrs.PutEmptyMap(otelEntityId)
	tm.PutStr(string(conventions.K8SPodNameKey), md.Name)
	tm.PutStr(string(conventions.K8SNamespaceNameKey), md.Namespace)
	tm.PutStr(string(conventions.K8SContainerNameKey), c.Name)
	tm.PutStr(swK8sClusterUid, os.Getenv(clusterUidEnv))

	// Entity attributes
	ea := attrs.PutEmptyMap(otelEntityAttributes)
	ea.PutStr(string(conventions.ContainerIDKey), c.ContainerId)
	ea.PutStr(k8sContainerStatus, c.State)
	ea.PutBool(k8sContainerInit, c.IsInitContainer)
	ea.PutBool(k8sContainerSidecar, c.IsSidecarContainer)
}

// addServiceMappingsResourceLog adds a new ResourceLogs to the provided Logs structure
// and sets required attributes on "resource" and "scopeLogs"
func addServiceMappingsResourceLog(ld plog.Logs, serviceMappingsLogSlice plog.LogRecordSlice) {
	rl := ld.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr(k8sLogType, "serviceendpointsmapping")
	lrs := rl.ScopeLogs().AppendEmpty().LogRecords()
	serviceMappingsLogSlice.CopyTo(lrs)
}

func transformManifestToServiceMappingLogs(m manifests.ServiceMapping, t pcommon.Timestamp) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	for _, addr := range m.GetAddresses() {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		attrs := lr.Attributes()
		attrs.PutStr(serviceName, m.GetServiceName())
		attrs.PutStr(string(conventions.K8SNamespaceNameKey), m.GetNamespace())
		attrs.PutStr(swK8sWorkloadIp, addr)
		attrs.PutStr(swK8sClusterUid, os.Getenv(clusterUidEnv))
	}

	return lrs
}

// transformContainersToContainerImageLogs returns a new [plog.LogRecordSlice] and appends
// all LogRecords containing container image information from the provided Containers.
func transformContainersToContainerImageLogs(containers map[string]manifests.Container, t pcommon.Timestamp) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	var noImage manifests.Image
	uniqueImagesWithValidIDs := make(map[manifests.Image]struct{})
	for _, c := range containers {
		if c.Image == noImage {
			continue
		}
		if _, seen := uniqueImagesWithValidIDs[c.Image]; seen {
			continue
		}
		if c.Image.ImageID == "" {
			continue
		}
		uniqueImagesWithValidIDs[c.Image] = struct{}{}
	}

	for image := range uniqueImagesWithValidIDs {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		addContainerImageAttributes(lr.Attributes(), image)
	}

	return lrs
}

// addContainerImageAttributes sets attributes on the provided map for the given Metadata and Image.
func addContainerImageAttributes(attrs pcommon.Map, i manifests.Image) {
	// Ingestion attributes
	attrs.PutStr(otelEntityEventType, entityState)
	attrs.PutStr(swEntityType, k8sContainerImageEntityType)

	// Telemetry mappings
	tm := attrs.PutEmptyMap(otelEntityId)
	tm.PutStr(string(conventions.ContainerImageIDKey), i.ImageID)
	tm.PutStr(string(conventions.ContainerImageNameKey), i.Name)
	tm.PutStr(string(conventions.ContainerImageTagKey), i.Tag)
}

// transformContainersToContainerImageRelationsLogs returns a new [plog.LogRecordSlice] and appends
// all LogRecords containing information about relations between containers and images from the provided Containers.
func transformContainersToContainerImageRelationsLogs(containers map[string]manifests.Container, md manifests.PodMetadata, t pcommon.Timestamp) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	for _, c := range containers {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		addContainerImageRelationAttributes(lr.Attributes(), md, c)
	}

	return lrs
}

// addContainerImageRelationAttributes sets attributes on the provided map for the given Metadata and Container.
func addContainerImageRelationAttributes(attrs pcommon.Map, md manifests.PodMetadata, c manifests.Container) {
	// Ingestion attributes
	attrs.PutStr(otelEntityEventType, relationshipUpdateEventType)
	attrs.PutStr(relationshipType, KubernetesResourceUsesImageRelationType)

	srcIds := attrs.PutEmptyMap(relationshipSrcEntityIds)
	destIds := attrs.PutEmptyMap(relationshipDestEntityIds)

	// Telemetry mappings
	attrs.PutStr(srcEntityType, k8sContainerEntityType)
	srcIds.PutStr(string(conventions.K8SPodNameKey), md.Name)
	srcIds.PutStr(string(conventions.K8SNamespaceNameKey), md.Namespace)
	srcIds.PutStr(string(conventions.K8SContainerNameKey), c.Name)
	srcIds.PutStr(swK8sClusterUid, os.Getenv(clusterUidEnv))

	attrs.PutStr(destEntityType, k8sContainerImageEntityType)
	destIds.PutStr(string(conventions.ContainerImageIDKey), c.Image.ImageID)
	destIds.PutStr(string(conventions.ContainerImageNameKey), c.Image.Name)
	destIds.PutStr(string(conventions.ContainerImageTagKey), c.Image.Tag)
}
