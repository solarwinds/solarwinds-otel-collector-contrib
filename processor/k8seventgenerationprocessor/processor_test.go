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
	"os"
	"testing"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/constants"
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
	podManifest           = `{"apiVersion":"v1","kind":"Pod","metadata":{"annotations":{"checksum/config":"123456","swo.cloud.solarwinds.com/cluster-uid":"test-cluster-uid"},"creationTimestamp":"2025-02-04T11:28:27Z","generateName":"test-generate-name","labels":{"some-label":"test-label"},"managedFields":[{"apiVersion":"v1","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:swo.cloud.solarwinds.com/cluster-uid":{}},"f:generateName":{},"f:labels":{"f:app.kubernetes.io/instance":{}}},"f:spec":{"f:affinity":{},"f:containers":{}}},"manager":"test-manager","operation":"Update","time":"2025-02-04T11:28:27Z"}],"name":"test-pod-name","namespace":"test-namespace","ownerReferences":[{"apiVersion":"apps/v1","blockOwnerDeletion":true,"controller":true,"kind":"DaemonSet","name":"test","uid":"123456789"}],"resourceVersion":"1.2.3","uid":"123456789"},"spec":{"containers":[{"args":["--warning"],"env":[{"name":"EBPF_NET_CLUSTER_NAME","value":"cluster name"}],"image":"test-container-image:v2","imagePullPolicy":"IfNotPresent","name":"test-container-name","resources":{"requests":{"memory":"50Mi"}},"securityContext":{"privileged":true},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","volumeMounts":[]}],"dnsPolicy":"ClusterFirstWithHostNet","enableServiceLinks":true,"hostNetwork":true,"hostPID":true,"initContainers":[{"command":["sh","-c","some command;"],"env":[{"name":"EBPF_NET_INTAKE_HOST","value":"test"}],"image":"test-init-container-image","imagePullPolicy":"IfNotPresent","name":"test-init-container-name","resources":{},"terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File","volumeMounts":[]}],"nodeName":"test-node","nodeSelector":{"kubernetes.io/os":"linux"},"preemptionPolicy":"PreemptLowerPriority","priority":0,"restartPolicy":"Always","schedulerName":"test-scheduler","securityContext":{"fsGroup":0,"runAsGroup":0,"runAsUser":0},"serviceAccount":"test-service-account","serviceAccountName":"test-service-account-name","terminationGracePeriodSeconds":30,"tolerations":[{"effect":"NoSchedule","operator":"Exists"}],"volumes":[{"hostPath":{"path":"/","type":"Directory"},"name":"host"}]},"status":{"conditions":[{"lastProbeTime":null,"lastTransitionTime":"2025-02-04T11:31:42Z","message":"containers with unready status","reason":"ContainersNotReady","status":"False","type":"ContainersReady"},{"lastProbeTime":null,"lastTransitionTime":"2025-02-04T11:28:27Z","status":"True","type":"PodScheduled"}],"containerStatuses":[{"containerID":"test-container-id","image":"test-container-image:v2","imageID":"test-container-image-id","lastState":{"terminated":{"containerID":"container-id","exitCode":255,"finishedAt":"2025-02-04T11:30:24Z","reason":"Error","startedAt":"2025-02-04T11:29:10Z"}},"name":"test-container-name","ready":false,"restartCount":1,"started":false,"state":{"terminated":{"containerID":"test-container-id","exitCode":255,"finishedAt":"2025-02-04T11:31:41Z","reason":"Error","startedAt":"2025-02-04T11:30:25Z"}}}],"hostIP":"1.2.3.4","hostIPs":[{"ip":"1.2.3.4"}],"initContainerStatuses":[{"containerID":"test-init-container-id","image":"test-init-container-image","imageID":"test-init-container-image-id","lastState":{},"name":"test-init-container-name","ready":true,"restartCount":0,"started":false,"state":{"terminated":{"containerID":"test-init-container-id","exitCode":0,"finishedAt":"2025-02-04T11:29:09Z","reason":"Completed","startedAt":"2025-02-04T11:28:27Z"}}}],"phase":"Running","podIP":"1.2.3.4","podIPs":[{"ip":"1.2.3.4"}],"qosClass":"Burstable","startTime":"2025-02-04T11:28:27Z"}}`
)

func TestVulnerabilityReportManifest(t *testing.T) {

	verifyVulnerabilityEntity := func(t *testing.T, attrs pcommon.Map, expectedVulnID string, expectedSeverity string, expectedScore float64) {
		vulnIDMap := getMapValue(t, attrs, constants.AttributeOtelEntityID)
		assert.Equal(t, expectedVulnID, getStringValue(t, vulnIDMap, constants.AttributeVulnerabilityID))

		vulnEntityAttrs := getMapValue(t, attrs, constants.AttributeOtelEntityAttributes)

		// Check vulnerability.id is present
		assert.Equal(t, expectedVulnID, getStringValue(t, vulnEntityAttrs, constants.AttributeVulnerabilityID))

		// Check sw.entity.type and sw.entity.name are NOT present
		_, swEntityTypeExists := vulnEntityAttrs.Get(constants.AttributeSwEntityType)
		assert.False(t, swEntityTypeExists, "sw.entity.type should not be in entity attributes")
		_, swEntityNameExists := vulnEntityAttrs.Get(constants.AttributeSwEntityName)
		assert.False(t, swEntityNameExists, "sw.entity.name should not be in entity attributes")

		// Check vulnerability.description contains title text (not vulnerability.title)
		description, descExists := vulnEntityAttrs.Get(constants.AttributeVulnerabilityDescription)
		assert.True(t, descExists, "vulnerability.description should be present")
		assert.NotEmpty(t, description.Str())

		// Check vulnerability.title is NOT present
		_, titleExists := vulnEntityAttrs.Get("vulnerability.title")
		assert.False(t, titleExists, "vulnerability.title should not be present")

		// Check other attributes
		assert.Equal(t, expectedSeverity, getStringValue(t, vulnEntityAttrs, constants.AttributeVulnerabilitySeverity))
		assert.Equal(t, expectedScore, getAttrValue(t, vulnEntityAttrs, constants.AttributeVulnerabilityScoreBase).Double())

		// Check cwe.id is present
		assert.Equal(t, expectedVulnID, getStringValue(t, vulnEntityAttrs, constants.AttributeCweID))

		// Check vulnerability.reference is an array
		rawAttrs := vulnEntityAttrs.AsRaw()
		references, ok := rawAttrs[constants.AttributeVulnerabilityReference]
		assert.True(t, ok, "vulnerability.reference should be present")
		_, isSlice := references.([]any)
		assert.True(t, isSlice, "vulnerability.reference should be an array")
	}

	verifyWorkloadFinding := func(t *testing.T, attrs pcommon.Map, expectedVulnID string, expectedWorkloadName string, expectedWorkloadNameKey string, expectedNamespace string, expectedScannerName string, expectedEntityType string) {
		assert.Equal(t, constants.EntityTypeVulnerability, getStringValue(t, attrs, constants.AttributeOtelEntityRelationshipSrcType))
		srcIDMap := getMapValue(t, attrs, constants.AttributeOtelEntityRelationshipSourceEntityID)
		assert.Equal(t, expectedVulnID, getStringValue(t, srcIDMap, constants.AttributeVulnerabilityID))

		assert.Equal(t, expectedEntityType, getStringValue(t, attrs, constants.AttributeOtelEntityRelationshipDstType))
		destIDMap := getMapValue(t, attrs, constants.AttributeOtelEntityRelationshipDestinationEntityID)
		assert.Equal(t, expectedWorkloadName, getStringValue(t, destIDMap, expectedWorkloadNameKey))
		assert.Equal(t, expectedNamespace, getStringValue(t, destIDMap, "k8s.namespace.name"))

		relAttrs := getMapValue(t, attrs, constants.AttributeOtelEntityRelationshipAttributes)
		assert.Equal(t, expectedScannerName, getStringValue(t, relAttrs, constants.AttributeScannerName))
	}

	verifyImageFinding := func(t *testing.T, attrs pcommon.Map, expectedVulnID string, expectedImageID string, expectedImageName string, expectedImageTag string, expectedScannerName string) {
		srcIDMap := getMapValue(t, attrs, constants.AttributeOtelEntityRelationshipSourceEntityID)
		assert.Equal(t, expectedVulnID, getStringValue(t, srcIDMap, constants.AttributeVulnerabilityID))

		destIDMap := getMapValue(t, attrs, constants.AttributeOtelEntityRelationshipDestinationEntityID)
		assert.Equal(t, expectedImageID, getStringValue(t, destIDMap, "container.image.id"))
		assert.Equal(t, expectedImageName, getStringValue(t, destIDMap, "container.image.name"))
		assert.Equal(t, expectedImageTag, getStringValue(t, destIDMap, "container.image.tag"))

		relAttrs := getMapValue(t, attrs, constants.AttributeOtelEntityRelationshipAttributes)
		assert.Equal(t, expectedScannerName, getStringValue(t, relAttrs, constants.AttributeScannerName))
	}

	findLogByType := func(t *testing.T, logRecords plog.LogRecordSlice, eventType, entityType, relType string) plog.LogRecord {
		for i := 0; i < logRecords.Len(); i++ {
			lr := logRecords.At(i)
			attrs := lr.Attributes()

			et, _ := attrs.Get(constants.AttributeOtelEntityEventType)
			if et.Str() != eventType {
				continue
			}

			if eventType == constants.EventTypeEntityState {
				eType, _ := attrs.Get(constants.AttributeOtelEntityType)
				if eType.Str() == entityType {
					return lr
				}
			} else if eventType == constants.EventTypeEntityRelationshipState {
				rType, _ := attrs.Get(constants.AttributeOtelEntityRelationshipType)
				if rType.Str() == relType {
					destIDMap := getMapValue(t, attrs, constants.AttributeOtelEntityRelationshipDestinationEntityID)
					if entityType == "Workload" {
						// Check for any workload-specific name attribute
						if _, ok := destIDMap.Get("k8s.deployment.name"); ok {
							return lr
						}
						if _, ok := destIDMap.Get("k8s.daemonset.name"); ok {
							return lr
						}
						if _, ok := destIDMap.Get("k8s.statefulset.name"); ok {
							return lr
						}
					} else if entityType == "Image" {
						if _, ok := destIDMap.Get("container.image.id"); ok {
							return lr
						}
					}
				}
			}
		}
		return plog.LogRecord{}
	}

	verifyContainerImage := func(t *testing.T, attrs pcommon.Map, expectedImageID string, expectedImageName string, expectedImageTag string) {
		ids := getMapValue(t, attrs, "otel.entity.id")
		assert.Equal(t, expectedImageID, getStringValue(t, ids, "container.image.id"))
		assert.Equal(t, expectedImageName, getStringValue(t, ids, "container.image.name"))
		assert.Equal(t, expectedImageTag, getStringValue(t, ids, "container.image.tag"))
	}

	verifyNewLog := func(t *testing.T, newLog plog.ResourceLogs) {
		assert.Equal(t, 1, newLog.Resource().Attributes().Len())
		assert.Equal(t, "entitystateevent", getStringValue(t, newLog.Resource().Attributes(), "sw.k8s.log.type"))

		scopeLogs := newLog.ScopeLogs()
		assert.Equal(t, 1, scopeLogs.Len())

		logRecords := scopeLogs.At(0).LogRecords()
		// 1 Vulnerability Entity + 1 Container Image Entity + 1 Finding (Workload) + 1 Finding (Image) = 4 logs
		assert.Equal(t, 4, logRecords.Len())

		vulnLog := findLogByType(t, logRecords, constants.EventTypeEntityState, constants.EntityTypeVulnerability, "")
		require.NotEqual(t, plog.LogRecord{}, vulnLog)
		verifyVulnerabilityEntity(t, vulnLog.Attributes(), "CVE-2016-2781", "LOW", 6.5)

		containerImageLog := findLogByType(t, logRecords, constants.EventTypeEntityState, "KubernetesContainerImage", "")
		require.NotEqual(t, plog.LogRecord{}, containerImageLog)
		verifyContainerImage(t, containerImageLog.Attributes(), "sha256:83c025f0faa6799fab6645102a98138e39a9a7db2be3bc792c79d72659b1805d", "kube-proxy", "v1.32.2")

		workloadFindingLog := findLogByType(t, logRecords, constants.EventTypeEntityRelationshipState, "Workload", constants.RelationshipTypeVulnerabilityFinding)
		require.NotEqual(t, plog.LogRecord{}, workloadFindingLog)
		verifyWorkloadFinding(t, workloadFindingLog.Attributes(), "CVE-2016-2781", "kube-proxy", "k8s.daemonset.name", "kube-system", "Trivy", constants.EntityTypeKubernetesDaemonSet)

		imageFindingLog := findLogByType(t, logRecords, constants.EventTypeEntityRelationshipState, "Image", constants.RelationshipTypeVulnerabilityFinding)
		require.NotEqual(t, plog.LogRecord{}, imageFindingLog)
		verifyImageFinding(t, imageFindingLog.Attributes(), "CVE-2016-2781", "sha256:83c025f0faa6799fab6645102a98138e39a9a7db2be3bc792c79d72659b1805d", "kube-proxy", "v1.32.2", "Trivy")
	}

	t.Run("TestInvalidManifest", func(t *testing.T) {
		consumer, err := startAndConsumeLogs(t, generateManifestLogs("VulnerabilityReport", ""))
		assert.Error(t, err)
		assert.Len(t, consumer.AllLogs(), 0)
	})

	t.Run("TestVulnerabilityReportLogBody", func(t *testing.T) {
		jsonBytes, err := os.ReadFile("internal/manifests/testdata/vulnerability_report.json")
		require.NoError(t, err)

		consumer, err := startAndConsumeLogs(t, generateManifestLogs("VulnerabilityReport", string(jsonBytes)))
		assert.NoError(t, err)

		allLogs := consumer.AllLogs()
		assert.Len(t, allLogs, 1)

		allResourceLogs := allLogs[0].ResourceLogs()
		assert.Equal(t, 2, allResourceLogs.Len())

		origLog := allResourceLogs.At(0)
		verifyOriginalLog(t, origLog, string(jsonBytes))

		newLog := allResourceLogs.At(1)
		verifyNewLog(t, newLog)
	})
}

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

	verifyContainer := func(t *testing.T, attrs pcommon.Map, expectedContainers map[string]manifests.Container) {
		assert.Equal(t, 4, attrs.Len())

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

	verifyImage := func(t *testing.T, attrs pcommon.Map, expectedImages map[string]manifests.Image) {
		assert.Equal(t, 3, attrs.Len())

		ids := getMapValue(t, attrs, "otel.entity.id")
		imageID := getStringValue(t, ids, "container.image.id")
		image, exists := expectedImages[imageID]
		assert.True(t, exists, "Image was not expected: %s", imageID)
		assert.Equal(t, image.Name, getStringValue(t, ids, "container.image.name"))
		assert.Equal(t, image.Tag, getStringValue(t, ids, "container.image.tag"))
	}

	verifyNewLog := func(t *testing.T, newLog plog.ResourceLogs, expectedContainers map[string]manifests.Container, expectedImages map[string]manifests.Image) {
		// resource
		assert.Equal(t, 1, newLog.Resource().Attributes().Len())
		assert.Equal(t, "entitystateevent", getStringValue(t, newLog.Resource().Attributes(), "sw.k8s.log.type"))

		// scope logs
		sl := newLog.ScopeLogs().At(0)
		assert.Equal(t, 1, newLog.ScopeLogs().Len())
		assert.Equal(t, 1, sl.Scope().Attributes().Len())
		assert.Equal(t, true, getBoolValue(t, sl.Scope().Attributes(), "otel.entity.event_as_log"))

		expectedNumberOfEvents := len(expectedContainers) + len(expectedImages) + len(expectedContainers) // containers + images + relationships (one for each container)
		// log records
		assert.Equal(t, expectedNumberOfEvents, sl.LogRecords().Len())
		for _, lr := range sl.LogRecords().All() {
			assert.Equal(t, timestamp, lr.ObservedTimestamp())

			attrs := lr.Attributes()
			assert.GreaterOrEqual(t, attrs.Len(), 1)
			assert.Equal(t, "", lr.Body().Str())

			eventType := getStringValue(t, attrs, "otel.entity.event.type")
			switch eventType {
			case "entity_state":
				entityType := getStringValue(t, attrs, "otel.entity.type")
				switch entityType {
				case "KubernetesContainer":
					verifyContainer(t, attrs, expectedContainers)
				case "KubernetesContainerImage":
					verifyImage(t, attrs, expectedImages)
				default:
					require.Fail(t, "Unexpected entity type", "entity_type: %s", entityType)
				}
			case "entity_relationship_state":
				//TODO: verify created relationship events
				//verifyRelationship(t, attrs, expectedContainers, expectedImages)
			default:
				require.Fail(t, "Unexpected event type", "event_type: %s", eventType)
			}
		}
	}

	t.Run("TestInvalidManifest", func(t *testing.T) {
		consumer, err := startAndConsumeLogs(t, generateManifestLogs("Pod", ""))
		assert.Error(t, err)
		assert.Len(t, consumer.AllLogs(), 0)
	})

	t.Run("TestPodLogBody", func(t *testing.T) {
		l := generateManifestLogs("Pod", podManifest)

		consumer, err := startAndConsumeLogs(t, l)
		assert.NoError(t, err)

		result := consumer.AllLogs()
		assert.Len(t, result, 1)
		assert.Equal(t, 2, result[0].ResourceLogs().Len())

		origLog := result[0].ResourceLogs().At(0)
		verifyOriginalLog(t, origLog, podManifest)

		newLog := result[0].ResourceLogs().At(1)
		verifyNewLog(t, newLog,
			map[string]manifests.Container{
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
			},
			map[string]manifests.Image{
				"test-container-image-id": {
					ImageID: "test-container-image-id",
					Name:    "test-container-image",
					Tag:     "v2",
				},
				"test-init-container-image-id": {
					ImageID: "test-init-container-image-id",
					Name:    "test-init-container-image",
					Tag:     "latest",
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
	rl.Resource().Attributes().PutStr("sw.k8s.cluster.uid", "test-cluster-uid")
	sl := rl.ScopeLogs().At(0)
	lr := sl.LogRecords().AppendEmpty()
	lr.Attributes().PutStr("k8s.object.kind", objKind)
	lr.Body().SetStr(manifest)
	lr.SetObservedTimestamp(timestamp)
	return l
}

func verifyOriginalLog(t *testing.T, origLog plog.ResourceLogs, expectedBody string) {
	assert.Equal(t, 2, origLog.Resource().Attributes().Len())
	assert.Equal(t, true, getBoolValue(t, origLog.Resource().Attributes(), "ORIGINAL_LOG"))
	assert.Equal(t, "test-cluster-uid", getStringValue(t, origLog.Resource().Attributes(), "sw.k8s.cluster.uid"))
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
