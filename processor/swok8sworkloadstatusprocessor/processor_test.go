// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package swok8sworkloadstatusprocessor // import "github.com/solarwinds/solarwinds-otel-collector-contrib/processor/swok8sworkloadstatusprocessor"

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	timestamp                     = pcommon.NewTimestampFromTime(time.Now())
	availableDeploymentManifest   = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"available-deployment","namespace":"default"},"spec":{"replicas":3,"selector":{"matchLabels":{"app":"available"}},"template":{"metadata":{"labels":{"app":"available"}},"spec":{"containers":[{"name":"app","image":"nginx:1.21"}]}}},"status":{"replicas":3,"updatedReplicas":3,"readyReplicas":3,"availableReplicas":3,"conditions":[{"type":"Available","status":"True"},{"type":"Progressing","status":"True"}]}}`
	outdatedDeploymentManifest    = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"outdated-deployment","namespace":"default"},"spec":{"replicas":3,"selector":{"matchLabels":{"app":"outdated"}},"template":{"metadata":{"labels":{"app":"outdated"}},"spec":{"containers":[{"name":"app","image":"nginx:1.20"}]}}},"status":{"replicas":3,"updatedReplicas":2,"readyReplicas":3,"availableReplicas":3,"conditions":[{"type":"Available","status":"True"},{"type":"Progressing","status":"False"}]}}`
	unavailableDeploymentManifest = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"unavailable-deployment","namespace":"default"},"spec":{"replicas":3,"selector":{"matchLabels":{"app":"unavailable"}},"template":{"metadata":{"labels":{"app":"unavailable"}},"spec":{"containers":[{"name":"app","image":"nginx:broken"}]}}},"status":{"replicas":3,"updatedReplicas":3,"readyReplicas":0,"availableReplicas":0,"conditions":[{"type":"Available","status":"False"},{"type":"Progressing","status":"True"}]}}`
	otherDeploymentManifest       = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"other-deployment","namespace":"default"},"spec":{"replicas":2,"selector":{"matchLabels":{"app":"other"}},"template":{"metadata":{"labels":{"app":"other"}},"spec":{"containers":[{"name":"app","image":"nginx:1.21"}]}}},"status":{"replicas":2,"updatedReplicas":2,"readyReplicas":1,"conditions":[{"type":"Progressing","status":"True"}]}}`
	emptyStatusDeploymentManifest = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"zero-pods-deployment","namespace":"default"},"spec":{"replicas":0,"selector":{"matchLabels":{"app":"zero-pods"}},"template":{"metadata":{"labels":{"app":"zero-pods"}},"spec":{"containers":[{"name":"app","image":"nginx:1.21"}]}}},"status":{"replicas":0,"conditions":[]}}`
	statefulSetManifest           = `{"apiVersion":"apps/v1","kind":"StatefulSet","metadata":{"name":"app","namespace":"test-namespace"},"spec":{"serviceName":"","replicas":2,"selector":{"matchLabels":{"app":"app"}},"template":{"metadata":{"labels":{"app":"app"}},"spec":{"containers":[{"name":"app","image":"nginx:1.25-alpine","ports":[{"containerPort":80}]}]}}}}`
	daemonSetManifest             = `{"apiVersion":"apps/v1","kind":"DaemonSet","metadata":{"name":"app","namespace":"test-namespace"},"spec":{"selector":{"matchLabels":{"app":"app"}},"template":{"metadata":{"labels":{"app":"app"}},"spec":{"containers":[{"name":"app","image":"nginx:1.21"}]}}}}`
)

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

func workloadAttrKey(kind string) string {
	return "sw.k8s." + strings.ToLower(kind) + ".status"
}

func runWorkloadStatusTest(t *testing.T, kind, manifest string, pods []*corev1.Pod, expectedStatus string) {
	mock, reset := MockKubeClient()
	defer reset()

	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	consumer := new(consumertest.LogsSink)

	c, err := factory.CreateLogs(context.Background(), processortest.NewNopSettings(factory.Type()), cfg, consumer)
	require.NoError(t, err)

	for _, pod := range pods {
		_, err := mock.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	combined := plog.NewLogs()
	generateManifestLogs(kind, manifest).ResourceLogs().MoveAndAppendTo(combined.ResourceLogs())

	require.NoError(t, c.Start(context.Background(), componenttest.NewNopHost()))
	defer func() { require.NoError(t, c.Shutdown(context.Background())) }()

	require.NoError(t, c.ConsumeLogs(context.Background(), combined))

	allLogs := consumer.AllLogs()
	require.Len(t, allLogs, 1)
	attrKey := workloadAttrKey(kind)
	status, exists := allLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().Get(attrKey)
	require.True(t, exists, "expected attribute %s to exist", attrKey)
	require.Equal(t, expectedStatus, status.Str())
}

func createPod(name, namespace, ownerKind, ownerName string, phase corev1.PodPhase) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1",
				Kind:       ownerKind,
				Name:       ownerName,
			}},
		},
		Status: corev1.PodStatus{Phase: phase},
	}
}

func TestStatusProcessorForDeployment(t *testing.T) {
	cases := []struct {
		name     string
		manifest string
		expected string
	}{
		{"Deployment with available status", availableDeploymentManifest, "AVAILABLE"},
		{"Deployment with outdated status", outdatedDeploymentManifest, "OUTDATED"},
		{"Deployment with unavailable status", unavailableDeploymentManifest, "UNAVAILABLE"},
		{"Deployment with other status", otherDeploymentManifest, "UNAVAILABLE"},
		{"Deployment with zero pods", emptyStatusDeploymentManifest, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { runWorkloadStatusTest(t, "Deployment", c.manifest, nil, c.expected) })
	}
}

func TestStatusProcessorForStatefulSet(t *testing.T) {
	ownerKind, ownerName, ns := "StatefulSet", "app", "test-namespace"

	runningPods := []*corev1.Pod{
		createPod("app-0", ns, ownerKind, ownerName, corev1.PodRunning),
		createPod("app-1", ns, ownerKind, ownerName, corev1.PodRunning),
	}
	failedPod := createPod("app-2", ns, ownerKind, ownerName, corev1.PodFailed)
	unknownPod := createPod("app-3", ns, ownerKind, ownerName, corev1.PodUnknown)

	cases := []struct {
		name     string
		pods     []*corev1.Pod
		expected string
	}{
		{"StatefulSet with running status", runningPods, "RUNNING"},
		{"StatefulSet with failed status", []*corev1.Pod{failedPod, unknownPod}, "FAILED"},
		{"StatefulSet with pending status", append(append([]*corev1.Pod{}, runningPods...), failedPod, unknownPod), "PENDING"},
		{"StatefulSet with no status", []*corev1.Pod{}, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { runWorkloadStatusTest(t, ownerKind, statefulSetManifest, c.pods, c.expected) })
	}
}

func TestStatusProcessorForDaemonSet(t *testing.T) {
	ownerKind, ownerName, ns := "DaemonSet", "app", "test-namespace"

	runningPods := []*corev1.Pod{
		createPod("app-0", ns, ownerKind, ownerName, corev1.PodRunning),
		createPod("app-1", ns, ownerKind, ownerName, corev1.PodRunning),
	}
	failedPod := createPod("app-2", ns, ownerKind, ownerName, corev1.PodFailed)
	unknownPod := createPod("app-3", ns, ownerKind, ownerName, corev1.PodUnknown)

	cases := []struct {
		name     string
		pods     []*corev1.Pod
		expected string
	}{
		{"DaemonSet with running status", runningPods, "RUNNING"},
		{"DaemonSet with failed status", []*corev1.Pod{failedPod, unknownPod}, "FAILED"},
		{"DaemonSet with pending status", append(append([]*corev1.Pod{}, runningPods...), failedPod, unknownPod), "PENDING"},
		{"DaemonSet with no status", []*corev1.Pod{}, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { runWorkloadStatusTest(t, ownerKind, daemonSetManifest, c.pods, c.expected) })
	}
}
