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
	"context"
	"testing"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/manifests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"
)

var (
	timestamp = pcommon.NewTimestampFromTime(time.Now())

	// some parts from original manifests that are not used by the algorithm are removed for simplicity
	endpointManifest      = `{"apiVersion":"v1","kind":"Endpoints","metadata":{"name":"test-name","namespace":"test-namespace"},"subsets":[{"addresses":[{"ip":"192.168.1.1"},{"ip":"192.168.1.2"}]}]}`
	endpointSliceManifest = `{"apiVersion":"discovery.k8s.io/v1","kind":"EndpointSlice","endpoints":[{"addresses":["192.168.1.3","192.168.1.4"]}],"metadata":{"labels":{"kubernetes.io/service-name":"test-name"},"name":"test-name-00001","namespace":"test-namespace"},"addressType":"IPv4"}`
	podManifest           = `{"apiVersion":"v1","kind":"Pod","metadata":{"annotations":{"checksum/config":"123456","swo.cloud.solarwinds.com/cluster-uid":"test-cluster-uid"},"creationTimestamp":"2025-02-04T11:28:27Z","generateName":"test-generate-name","labels":{"some-label":"test-label"},"managedFields":[{"apiVersion":"v1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:swo.cloud.solarwinds.com/cluster-uid":{}},"f:generateName":{},"f:labels":{"f:app.kubernetes.io/instance":{}}},"f:spec":{"f:affinity":{},"f:containers":{}}},"manager":"test-manager","operation":"Update","time":"2025-02-04T11:28:27Z"}],"name":"test-pod-name","namespace":"test-namespace","ownerReferences":[{"apiVersion":"apps/v1","blockOwnerDeletion":true,"controller":true,"kind":"DaemonSet","name":"test","uid":"123456789"}],"resourceVersion":"1.2.3","uid":"123456789"},"spec":{"containers":[{"args":["--warning"],"env":[{"name":"EBPF_NET_CLUSTER_NAME","value":"cluster name"}],"image":"test-container-id","imagePullPolicy":"IfNotPresent","name":"test-container-name","resources":{"requests":{"memory":"50Mi"}},"securityContext":{"privileged":true},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","volumeMounts":[]}],"dnsPolicy":"ClusterFirstWithHostNet","enableServiceLinks":true,"hostNetwork":true,"hostPID":true,"initContainers":[{"command":["sh","-c","some command;"],"env":[{"name":"EBPF_NET_INTAKE_HOST","value":"test"}],"image":"test-image-container-image","imagePullPolicy":"IfNotPresent","name":"test-init-container-name","resources":{},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","volumeMounts":[]}],"nodeName":"test-node","nodeSelector":{"kubernetes.io/os":"linux"},"preemptionPolicy":"PreemptLowerPriority","priority":0,"restartPolicy":"Always","schedulerName":"test-scheduler","securityContext":{"fsGroup":0,"runAsGroup":0,"runAsUser":0},"serviceAccount":"test-service-account","serviceAccountName":"test-service-account-name","terminationGracePeriodSeconds":30,"tolerations":[{"effect":"NoSchedule","operator":"Exists"}],"volumes":[{"hostPath":{"path":"/","type":"Directory"},"name":"host"}]},"status":{"conditions":[{"lastProbeTime":null,"lastTransitionTime":"2025-02-04T11:31:42Z","message":"containers with unready status","reason":"ContainersNotReady","status":"False","type":"ContainersReady"},{"lastProbeTime":null,"lastTransitionTime":"2025-02-04T11:28:27Z","status":"True","type":"PodScheduled"}],"containerStatuses":[{"containerID":"test-container-id","image":"test-container-image-id","imageID":"test-container-image-id","lastState":{"terminated":{"containerID":"container-id","exitCode":255,"finishedAt":"2025-02-04T11:30:24Z","reason":"Error","startedAt":"2025-02-04T11:29:10Z"}},"name":"test-container-name","ready":false,"restartCount":1,"started":false,"state":{"terminated":{"containerID":"test-container-id","exitCode":255,"finishedAt":"2025-02-04T11:31:41Z","reason":"Error","startedAt":"2025-02-04T11:30:25Z"}}}],"hostIP":"1.2.3.4","hostIPs":[{"ip":"1.2.3.4"}],"initContainerStatuses":[{"containerID":"test-init-container-id","image":"test-init-container-image","imageID":"test-init-container-image-id","lastState":{},"name":"test-init-container-name","ready":true,"restartCount":0,"started":false,"state":{"terminated":{"containerID":"test-init-container-id","exitCode":0,"finishedAt":"2025-02-04T11:29:09Z","reason":"Completed","startedAt":"2025-02-04T11:28:27Z"}}}],"phase":"Running","podIP":"1.2.3.4","podIPs":[{"ip":"1.2.3.4"}],"qosClass":"Burstable","startTime":"2025-02-04T11:28:27Z"}}`
)

func TestEmptyResourceLogs(t *testing.T) {
	// processor does not decorate empty Log structure
	consumer, err := startAndConsumeLogs(t, plog.NewLogs())

	assert.NoError(t, err)
	assert.Len(t, consumer.AllLogs(), 1)
	l := consumer.AllLogs()[0]
	assert.Equal(t, 0, l.ResourceLogs().Len())
}

func TestEmptyLogRecords(t *testing.T) {
	// processor does not decorate empty Log Records
	consumer, err := startAndConsumeLogs(t, generateLogs())

	assert.NoError(t, err)
	assert.Len(t, consumer.AllLogs(), 1)
	l := consumer.AllLogs()[0]
	assert.Equal(t, 0, l.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().Len())
}

func TestDifferentKindBody(t *testing.T) {
	// processor does not decorate Log Records with different kind than Pod/Endpoints/EndpointSlice
	consumer, err := startAndConsumeLogs(t, generateManifestLogs("Deployment", ""))
	assert.NoError(t, err)

	assert.Len(t, consumer.AllLogs(), 1)
	lr := consumer.AllLogs()[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	assert.Equal(t, "", lr.Body().Str())
	assert.Equal(t, "Deployment", getStringValue(t, lr.Attributes(), "k8s.object.kind"))
}

func TestContainerAndServiceMappingsEndUpInDifferentResources(t *testing.T) {
	combinedMessage := plog.NewLogs()
	generateManifestLogs("Pod", podManifest).ResourceLogs().MoveAndAppendTo(combinedMessage.ResourceLogs())
	generateManifestLogs("Endpoints", endpointManifest).ResourceLogs().MoveAndAppendTo(combinedMessage.ResourceLogs())
	generateManifestLogs("EndpointSlice", endpointSliceManifest).ResourceLogs().MoveAndAppendTo(combinedMessage.ResourceLogs())

	consumer, err := startAndConsumeLogs(t, combinedMessage)
	assert.NoError(t, err)

	allLogs := consumer.AllLogs()
	assert.Len(t, allLogs, 1)

	allResourceLogs := allLogs[0].ResourceLogs()
	assert.Equal(t, 5, allResourceLogs.Len())

	assert.Equal(t, combinedMessage.ResourceLogs().At(0), allResourceLogs.At(0), "First should be the original resource for Pod")
	assert.Equal(t, combinedMessage.ResourceLogs().At(1), allResourceLogs.At(1), "Second should be the original resource for Endpoints")
	assert.Equal(t, combinedMessage.ResourceLogs().At(2), allResourceLogs.At(2), "Third should be the original resource for EndpointSlice")

	assert.Equal(t, "entitystateevent", getStringValue(t, allResourceLogs.At(3).Resource().Attributes(), "sw.k8s.log.type"), "Fourth should be the resource for containers")
	assert.Equal(t, "serviceendpointsmapping", getStringValue(t, allResourceLogs.At(4).Resource().Attributes(), "sw.k8s.log.type"), "Fifth should be the resource for service mappings")
}

func TestPodManifests(t *testing.T) {

	verifyNewLog := func(t *testing.T, newLog plog.ResourceLogs, expectedContainers map[string]manifests.Container) {
		// resource
		assert.Equal(t, 1, newLog.Resource().Attributes().Len())
		assert.Equal(t, "entitystateevent", getStringValue(t, newLog.Resource().Attributes(), "sw.k8s.log.type"))

		// scope logs
		sl := newLog.ScopeLogs().At(0)
		assert.Equal(t, 1, newLog.ScopeLogs().Len())
		assert.Equal(t, 1, sl.Scope().Attributes().Len())
		assert.Equal(t, true, getBoolValue(t, sl.Scope().Attributes(), "otel.entity.event_as_log"))

		// log records
		assert.Equal(t, len(expectedContainers), sl.LogRecords().Len())
		for _, lr := range sl.LogRecords().All() {
			assert.Equal(t, timestamp, lr.ObservedTimestamp())

			attrs := lr.Attributes()
			assert.Equal(t, 4, attrs.Len())
			assert.Equal(t, "", lr.Body().Str())

			eventType := getStringValue(t, attrs, "otel.entity.event.type")
			assert.Equal(t, "entity_state", eventType)
			entityType := getStringValue(t, attrs, "otel.entity.type")
			assert.Equal(t, "KubernetesContainer", entityType)

			ids := getMapValue(t, attrs, "otel.entity.id")
			assert.Equal(t, "test-pod-name", getStringValue(t, ids, "k8s.pod.name"))
			assert.Equal(t, "test-namespace", getStringValue(t, ids, "k8s.namespace.name"))
			assert.Equal(t, "test-cluster-uid", getStringValue(t, ids, "sw.k8s.cluster.uid"))
			containerName := getStringValue(t, ids, "k8s.container.name")
			c, exists := expectedContainers[containerName]
			assert.True(t, exists, "Container was not expected: %s", containerName)

			otherAttrs := getMapValue(t, attrs, "otel.entity.attributes")
			assert.Equal(t, c.ContainerId, getStringValue(t, otherAttrs, "container.id"))
			assert.Equal(t, c.State, getStringValue(t, otherAttrs, "sw.k8s.container.status"))
			assert.Equal(t, c.IsInitContainer, getBoolValue(t, otherAttrs, "sw.k8s.container.init"), "Unexpected value for sw.k8s.container.init attribute: %s", containerName)
			assert.Equal(t, c.IsSidecarContainer, getBoolValue(t, otherAttrs, "sw.k8s.container.sidecar"), "Unexpected value for sw.k8s.container.sidecar attribute: %s", containerName)
		}
	}

	t.Run("TestInvalidManifest", func(t *testing.T) {
		consumer, err := startAndConsumeLogs(t, generateManifestLogs("Pod", ""))
		assert.Error(t, err)
		assert.Len(t, consumer.AllLogs(), 0)
	})

	t.Run("TestPodLogBody", func(t *testing.T) {
		t.Setenv("CLUSTER_UID", "test-cluster-uid")
		l := generateManifestLogs("Pod", podManifest)

		consumer, err := startAndConsumeLogs(t, l)
		assert.NoError(t, err)

		result := consumer.AllLogs()
		assert.Len(t, result, 1)
		assert.Equal(t, 2, result[0].ResourceLogs().Len())

		origLog := result[0].ResourceLogs().At(0)
		verifyOriginalLog(t, origLog, podManifest)

		newLog := result[0].ResourceLogs().At(1)
		verifyNewLog(t, newLog, map[string]manifests.Container{
			"test-container-name": {
				Name:            "test-container-name",
				ContainerId:     "test-container-id",
				State:           "terminated",
				IsInitContainer: false,
			},
			"test-init-container-name": {
				Name:            "test-init-container-name",
				ContainerId:     "test-init-container-id",
				State:           "terminated",
				IsInitContainer: true,
			},
		})
	})
}

func TestEndpointAndEndpointSliceManifests(t *testing.T) {

	verifyNewLog := func(t *testing.T, newLog plog.ResourceLogs, expectedEndpoints []string) {
		// resource
		assert.Equal(t, 1, newLog.Resource().Attributes().Len())
		assert.Equal(t, "serviceendpointsmapping", getStringValue(t, newLog.Resource().Attributes(), "sw.k8s.log.type"))

		// scope logs
		sl := newLog.ScopeLogs().At(0)
		assert.Equal(t, 1, newLog.ScopeLogs().Len())
		assert.Equal(t, 0, sl.Scope().Attributes().Len())

		// log records
		assert.Equal(t, len(expectedEndpoints), sl.LogRecords().Len())
		for _, lr := range sl.LogRecords().All() {
			assert.Equal(t, timestamp, lr.ObservedTimestamp())

			attrs := lr.Attributes()
			assert.Equal(t, 4, attrs.Len())
			assert.Equal(t, "", lr.Body().Str())

			assert.Equal(t, "test-name", getStringValue(t, attrs, "k8s.service.name"))
			assert.Equal(t, "test-namespace", getStringValue(t, attrs, "k8s.namespace.name"))
			workloadIp := getStringValue(t, attrs, "sw.k8s.workload.ip")
			assert.Contains(t, expectedEndpoints, workloadIp, "Unexpected endpoint IP: %s", workloadIp)
			assert.Equal(t, "test-cluster-uid", getStringValue(t, attrs, "sw.k8s.cluster.uid"))
		}
	}

	t.Run("TestInvalidEndpointManifest", func(t *testing.T) {
		_, err := startAndConsumeLogs(t, generateManifestLogs("Endpoints", ""))
		assert.Error(t, err)
	})

	t.Run("TestInvalidEndpointSliceManifest", func(t *testing.T) {
		_, err := startAndConsumeLogs(t, generateManifestLogs("EndpointSlice", ""))
		assert.Error(t, err)
	})

	t.Run("TestServiceMappingExtractionFromEndpoint", func(t *testing.T) {
		t.Setenv("CLUSTER_UID", "test-cluster-uid")
		l := generateManifestLogs("Endpoints", endpointManifest)

		consumer, err := startAndConsumeLogs(t, l)
		assert.NoError(t, err)

		result := consumer.AllLogs()
		assert.Len(t, result, 1)
		assert.Equal(t, 2, result[0].ResourceLogs().Len())

		origLog := result[0].ResourceLogs().At(0)
		verifyOriginalLog(t, origLog, endpointManifest)

		newLog := result[0].ResourceLogs().At(1)
		verifyNewLog(t, newLog, []string{"192.168.1.1", "192.168.1.2"})
	})

	t.Run("TestServiceMappingExtractionFromEndpointSlice", func(t *testing.T) {
		t.Setenv("CLUSTER_UID", "test-cluster-uid")
		l := generateManifestLogs("EndpointSlice", endpointSliceManifest)

		consumer, err := startAndConsumeLogs(t, l)
		assert.NoError(t, err)

		result := consumer.AllLogs()
		assert.Len(t, result, 1)
		assert.Equal(t, 2, result[0].ResourceLogs().Len())

		origLog := result[0].ResourceLogs().At(0)
		verifyOriginalLog(t, origLog, endpointSliceManifest)

		newLog := result[0].ResourceLogs().At(1)
		verifyNewLog(t, newLog, []string{"192.168.1.3", "192.168.1.4"})
	})
}

func startAndConsumeLogs(t *testing.T, logs plog.Logs) (consumer *consumertest.LogsSink, consumeLogsErr error) {
	t.Helper()
	ctx := context.Background()
	consumer = new(consumertest.LogsSink)
	processor, err := createLogsProcessor(ctx, processortest.NewNopSettings(processortest.NopType), createDefaultConfig(), consumer)
	require.NoError(t, err)

	err = processor.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)

	consumeLogsErr = processor.ConsumeLogs(ctx, logs)
	return
}

func generateLogs() plog.Logs {
	l := plog.NewLogs()
	ls := l.ResourceLogs().AppendEmpty()
	ls.ScopeLogs().AppendEmpty()
	return l
}

func generateManifestLogs(objKind string, manifest string) plog.Logs {
	l := generateLogs()
	rl := l.ResourceLogs().At(0)
	rl.Resource().Attributes().PutBool("ORIGINAL_LOG", true)
	sl := rl.ScopeLogs().At(0)
	lr := sl.LogRecords().AppendEmpty()
	lr.Attributes().PutStr("k8s.object.kind", objKind)
	lr.Body().SetStr(manifest)
	lr.SetObservedTimestamp(timestamp)
	return l
}

func verifyOriginalLog(t *testing.T, origLog plog.ResourceLogs, expectedBody string) {
	assert.Equal(t, 1, origLog.Resource().Attributes().Len())
	assert.Equal(t, true, getBoolValue(t, origLog.Resource().Attributes(), "ORIGINAL_LOG"))
	assert.Equal(t, 1, origLog.ScopeLogs().Len())
	origBody := origLog.ScopeLogs().At(0).LogRecords().At(0).Body().Str()
	assert.Equal(t, expectedBody, origBody)
}

func getStringValue(t *testing.T, attrs pcommon.Map, key string) string {
	t.Helper()
	return getAttrValue(t, attrs, key).AsString()
}

func getMapValue(t *testing.T, attrs pcommon.Map, key string) pcommon.Map {
	t.Helper()
	return getAttrValue(t, attrs, key).Map()
}

func getBoolValue(t *testing.T, attrs pcommon.Map, key string) bool {
	t.Helper()
	return getAttrValue(t, attrs, key).Bool()
}

func getAttrValue(t *testing.T, attrs pcommon.Map, key string) pcommon.Value {
	t.Helper()
	value, ok := attrs.Get(key)
	if !ok {
		require.Fail(t, "Attribute not found", "Key: %s", key)
	}
	return value
}
