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
	conventions "go.opentelemetry.io/otel/semconv/v1.6.1"
)

const (
	k8sLogType             = "sw.k8s.log.type"
	clusterUidEnv          = "CLUSTER_UID"
	k8sContainerEntityType = "KubernetesContainer"
	entityState            = "entity_state"

	// Attributes for OTel entity events identification
	otelEntityEventAsLog = "otel.entity.event_as_log"
	otelEntityEventType  = "otel.entity.event.type"
	swEntityType         = "otel.entity.type"

	// Attributes for telemetry mapping
	otelEntityId    = "otel.entity.id"
	swK8sClusterUid = "sw.k8s.cluster.uid"
	swK8sWorkloadIp = "sw.k8s.workload.ip"
	serviceName     = "k8s.service.name"

	// Attributes containing additional information about container
	otelEntityAttributes = "otel.entity.attributes"
	k8sContainerStatus   = "sw.k8s.container.status"
	k8sContainerInit     = "sw.k8s.container.init"
	k8sContainerSidecar  = "sw.k8s.container.sidecar"
)

// addContainerResourceLog adds a new ResourceLogs to the provided Logs structure
// and sets required attributes on "resource" and "scopeLogs"
func addContainerResourceLog(ld plog.Logs, containersLogSlice plog.LogRecordSlice) {
	rl := ld.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr(k8sLogType, "entitystateevent")
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().Attributes().PutBool(otelEntityEventAsLog, true)
	containersLogSlice.CopyTo(sl.LogRecords())
}

// transformManifestToContainerLogs returns a new plog.LogRecordSlice and appends
// all LogRecords containing container information from the provided Manifest.
func transformManifestToContainerLogs(m *manifests.PodManifest, t pcommon.Timestamp) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	containers := m.GetContainers()
	for _, c := range containers {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		addContainerAttributes(lr.Attributes(), m.Metadata, c)
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
