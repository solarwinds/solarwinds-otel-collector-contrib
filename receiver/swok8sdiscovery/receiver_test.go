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

package swok8sdiscovery

import (
	"context"
	"testing"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver/receivertest"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testClusterUid = "test-cluster-uid"
)

func TestNewReceiver(t *testing.T) {
	t.Parallel()

	rCfg := createImplicitConfig().(*Config)

	err := rCfg.Validate()
	require.NoError(t, err)

	rCfg.makeClient = newMockClient().getMockClient
	r, err := newReceiver(
		receivertest.NewNopSettings(receivertest.NopType),
		rCfg,
		consumertest.NewNop(),
	)

	require.NoError(t, err)
	require.NotNil(t, r)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	assert.NoError(t, r.Shutdown(context.Background()))
}

// captureLogsImages is a local sink for image discovery tests.
type captureLogsImages struct{ logs []plog.Logs }

func (c *captureLogsImages) ConsumeLogs(_ context.Context, l plog.Logs) error {
	c.logs = append(c.logs, l)
	return nil
}
func (c *captureLogsImages) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// runImageCycle executes a single discovery cycle with provided objects.
func runImageCycle(t *testing.T, cfg *Config, objects ...runtime.Object) []plog.Logs {
	fakeClient := fake.NewSimpleClientset(objects...)
	cfg.makeClient = func() (k8s.Interface, error) { return fakeClient, nil }
	cap := &captureLogsImages{}
	rec, err := newReceiver(receivertest.NewNopSettings(receivertest.NopType), cfg, cap)

	done := make(chan bool, 1)

	if rec, ok := rec.(*swok8sdiscoveryReceiver); ok {
		rec.clusterUid = testClusterUid
		rec.cycleCallback = func() { done <- true }
	}
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, rec.Start(ctx, componenttest.NewNopHost()))

	<-done
	rec.Shutdown(ctx)

	return cap.logs
}

func collectAttrs(all []plog.Logs, key, value string) []pcommon.Map {
	var out []pcommon.Map
	for _, l := range all {
		rl := l.ResourceLogs()
		for i := 0; i < rl.Len(); i++ {
			sl := rl.At(i).ScopeLogs()
			for j := 0; j < sl.Len(); j++ {
				recs := sl.At(j).LogRecords()
				for k := 0; k < recs.Len(); k++ {
					attrs := recs.At(k).Attributes()
					if v, ok := attrs.Get(key); ok && v.Str() == value {
						out = append(out, attrs)
					}
				}
			}
		}
	}

	return out
}

// collectDBAttrs gathers attribute maps for a specific db.type from a slice of logs.
func collectDBAttrs(all []plog.Logs, dbType string) []pcommon.Map {
	var out []pcommon.Map
	for _, attrs := range collectAttrs(all, otelEntityEventType, entityState) {
		if v, ok := attrs.Get(otelEntityId); ok {
			if vv, ok := v.Map().Get(swDiscoveryDbType); ok && vv.Str() == dbType {
				out = append(out, attrs)
			}
		}
	}
	return out
}

func checkAttr(t *testing.T, attrs pcommon.Map, key string, expectedValue string) {
	if v, ok := attrs.Get(key); ok {
		assert.Equal(t, expectedValue, v.Str())
	} else {
		t.Fatalf("missing attribute: %s", key)
	}
}

func checkAttrInt(t *testing.T, attrs pcommon.Map, key string, expectedValue int64) {
	if v, ok := attrs.Get(key); ok {
		assert.Equal(t, expectedValue, v.Int())
	} else {
		t.Fatalf("missing attribute: %s", key)
	}
}

func checkAttrInts(t *testing.T, attrs pcommon.Map, key string, expected []int) {
	if v, ok := attrs.Get(key); ok {
		arr := v.Slice()

		assert.Equal(t, len(expected), arr.Len())
		for i := 0; i < arr.Len(); i++ {
			assert.Equal(t, expected[i], int(arr.At(i).Int()))
		}
	} else {
		t.Fatalf("missing attribute: %s", key)
	}
}

func getAttrMap(t *testing.T, attrs pcommon.Map, key string) pcommon.Map {
	if v, ok := attrs.Get(key); ok {
		return v.Map()
	} else {
		t.Fatalf("missing attribute: %s", key)
		return pcommon.Map{}
	}
}

func TestDomainDiscovery_SingleMatch(t *testing.T) {
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
		Interval:  time.Second,
		Database: &DatabaseDiscoveryConfig{
			DomainRules: []*DomainRule{{
				DatabaseType: "redis",
				Patterns:     []string{".*redis.example.com"},
			}},
		},
	}
	require.NoError(t, cfg.Validate())

	trueVal := true

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: "db", Name: "redis-deploy", UID: types.UID("deploy-uid"), Labels: map[string]string{"app": "redis"}},
	}
	rs := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "db", Name: "redis-deploy-rs", UID: types.UID("rs-uid"), Labels: map[string]string{"app": "redis"}, OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Deployment", Name: deploy.Name, UID: deploy.UID, Controller: &trueVal}}},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "db", Name: "redis-pod", UID: types.UID("pod-uid"), Labels: map[string]string{"app": "redis"}, OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: rs.Name, UID: rs.UID, Controller: &trueVal}}},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "redis", Image: "docker.io/library/redis:7"}}},
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "db", Name: "redis-ext"},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: "cache-01.redis.example.com",
			Selector:     map[string]string{"app": "redis"},
		},
	}

	logs := runImageCycle(t, cfg, deploy, rs, pod, svc)
	attrs := collectDBAttrs(logs, "redis")
	require.Len(t, attrs, 1)
	idAttrs := getAttrMap(t, attrs[0], otelEntityId)
	require.Equal(t, idAttrs.Len(), 3)
	checkAttr(t, idAttrs, swDiscoveryDbAddress, "cache-01.redis.example.com")
	checkAttr(t, idAttrs, swDiscoveryDbType, "redis")
	checkAttr(t, idAttrs, swDiscoveryId, "external")
	entityAttrs := getAttrMap(t, attrs[0], otelEntityAttributes)
	checkAttr(t, entityAttrs, swDiscoveryDbName, "cache-01.redis.example.com#db#redis-deploy")

	relAttrs := collectAttrs(logs, otelEntityEventType, relationshipState)
	require.Len(t, relAttrs, 1, "expected relationship when workload resolved")
	checkAttr(t, relAttrs[0], otelEntityRelationSourceType, "KubernetesDeployment")
	srcIDs := getAttrMap(t, relAttrs[0], otelEntityRelationSourceID)
	checkAttr(t, srcIDs, "k8s.deployment.name", "redis-deploy")
	checkAttr(t, srcIDs, k8sNamespace, "db")
}

func TestDomainDiscovery_DedupByDomainHint(t *testing.T) {
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
		Interval:  time.Second,
		Database: &DatabaseDiscoveryConfig{
			DomainRules: []*DomainRule{
				{DatabaseType: "pg", Patterns: []string{".*db.company.internal"}, DomainHints: []string{"primary"}},
				{DatabaseType: "mysql", Patterns: []string{".*db.company.internal"}, DomainHints: []string{"mysql"}}, // higher potential hint score
			}},
	}
	require.NoError(t, cfg.Validate())

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "mysql-svc"},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: "some-db.company.internal",
		},
	}

	logs := runImageCycle(t, cfg, svc)
	require.Len(t, logs, 1)
	attrs := collectDBAttrs(logs, "mysql") // expect mysql chosen due to more hint matches ("company" substring)
	require.Len(t, attrs, 1)
	// ensure pg not emitted
	pgAttrs := collectDBAttrs(logs, "pg")
	require.Len(t, pgAttrs, 0)

	idMap := getAttrMap(t, attrs[0], otelEntityId)
	checkAttr(t, idMap, swDiscoveryDbAddress, "some-db.company.internal")
	checkAttr(t, idMap, swDiscoveryDbType, "mysql")
}

func TestDomainDiscovery_MultipleMatchNoDomainHint(t *testing.T) {
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
		Interval:  time.Second,
		Database: &DatabaseDiscoveryConfig{
			DomainRules: []*DomainRule{
				{DatabaseType: "pg", Patterns: []string{".*db.company.internal"}, DomainHints: []string{"primary"}},
				{DatabaseType: "mysql", Patterns: []string{".*db.company.internal"}, DomainHints: []string{"mysql"}}, // higher potential hint score
			}},
	}
	require.NoError(t, cfg.Validate())

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "unknown-svc"},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: "unknown-db.company.internal",
		},
	}

	logs := runImageCycle(t, cfg, svc)
	require.Len(t, logs, 0)
}

func TestImageDiscovery_SingleMatch(t *testing.T) {
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		Interval: time.Second,
		Database: &DatabaseDiscoveryConfig{
			ImageRules: []*ImageRule{{DatabaseType: "mongo", Patterns: []string{".*/mongo:.*"}, DefaultPort: 27017}}}}
	require.NoError(t, cfg.Validate())

	trueVal := true
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongo-ds",
			Namespace: "db",
			UID:       types.UID("ds-uid"),
			Labels:    map[string]string{"app": "mongo"},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongo-b",
			Namespace: "db",
			Labels:    map[string]string{"app": "mongo"},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
				Name:       ds.Name,
				UID:        ds.UID,
				Controller: &trueVal,
			}},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "mongo", Image: "docker.io/library/mongo:7", Ports: []corev1.ContainerPort{{ContainerPort: 27017}, {ContainerPort: 27018}}}}}}

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "mongo-svc", Namespace: "db"}, Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "mongo"}, Ports: []corev1.ServicePort{{Port: 27017}}}}

	logs := runImageCycle(t, cfg, ds, pod, svc)
	require.Len(t, logs, 2, "expected two log records (entity + relationship)")

	attrs := collectDBAttrs(logs, "mongo")
	require.Len(t, attrs, 1)
	id_attrs := getAttrMap(t, attrs[0], otelEntityId)
	require.Equal(t, id_attrs.Len(), 3)
	checkAttr(t, id_attrs, swDiscoveryDbAddress, "mongo-svc")
	checkAttr(t, id_attrs, swDiscoveryDbType, "mongo")
	checkAttr(t, id_attrs, swDiscoveryId, testClusterUid)
	entity_attrs := getAttrMap(t, attrs[0], otelEntityAttributes)
	checkAttr(t, entity_attrs, swDiscoveryDbName, "mongo-svc#db#mongo-ds")
	checkAttrInt(t, entity_attrs, swDiscoveryDbPort, 27017)

	relationAttrs := collectAttrs(logs, otelEntityEventType, relationshipState)
	require.Len(t, relationAttrs, 1)
	checkAttr(t, relationAttrs[0], otelEntityRelationType, "DiscoveredBy")
	checkAttr(t, relationAttrs[0], otelEntityRelationSourceType, "KubernetesDaemonSet")
	source_ids := getAttrMap(t, relationAttrs[0], otelEntityRelationSourceID)
	checkAttr(t, source_ids, "k8s.daemonset.name", "mongo-ds")
	checkAttr(t, source_ids, k8sNamespace, "db")
	checkAttr(t, source_ids, swK8sClusterUid, testClusterUid)

	checkAttr(t, relationAttrs[0], otelEntityRelationDestinationType, discoveredDatabaseEntityType)
	dst_ids := getAttrMap(t, relationAttrs[0], otelEntityRelationDestinationID)
	checkAttr(t, dst_ids, swDiscoveryDbAddress, "mongo-svc")
	checkAttr(t, dst_ids, swDiscoveryDbType, "mongo")
}

func TestImageDiscovery_NonDefaultPortMatch(t *testing.T) {
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		Interval: time.Second,
		Database: &DatabaseDiscoveryConfig{
			ImageRules: []*ImageRule{{DatabaseType: "mongo", Patterns: []string{".*/mongo:.*"}, DefaultPort: 8080}}}}
	require.NoError(t, cfg.Validate())

	pod1 := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "mongo-b", Namespace: "db", Labels: map[string]string{"app": "mongo"}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "mongo", Image: "docker.io/library/mongo:7", Ports: []corev1.ContainerPort{{ContainerPort: 27017}, {ContainerPort: 27018}}}}}}

	logs := runImageCycle(t, cfg, pod1)
	require.Len(t, logs, 1)
	attrs := collectDBAttrs(logs, "mongo")
	require.Len(t, attrs, 1)
	id_attrs := getAttrMap(t, attrs[0], otelEntityId)
	checkAttr(t, id_attrs, swDiscoveryDbAddress, "mongo-b")
	entity_attrs := getAttrMap(t, attrs[0], otelEntityAttributes)
	checkAttr(t, entity_attrs, swDiscoveryDbName, "mongo-b#db")
	checkAttrInts(t, entity_attrs, swDiscoveryDbPossiblePorts, []int{27017, 27018})

}

func TestImageDiscovery_DedupPerPortSignature(t *testing.T) {
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		Interval: time.Second,
		Database: &DatabaseDiscoveryConfig{
			ImageRules: []*ImageRule{{DatabaseType: "mongo", Patterns: []string{".*/mongo:.*"}, DefaultPort: 27017}}}}
	require.NoError(t, cfg.Validate())

	pod1 := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "mongo-a", Namespace: "db", Labels: map[string]string{"app": "mongo"}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "mongo", Image: "docker.io/library/mongo:7", Ports: []corev1.ContainerPort{{ContainerPort: 27017}}}}}}
	pod2 := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "mongo-b", Namespace: "db", Labels: map[string]string{"app": "mongo"}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "mongo", Image: "docker.io/library/mongo:7", Ports: []corev1.ContainerPort{{ContainerPort: 27017}, {ContainerPort: 27018}}}}}}

	logs := runImageCycle(t, cfg, pod1, pod2)
	attrs := collectDBAttrs(logs, "mongo")
	require.Len(t, attrs, 2, "expected two distinct mongo events due to different port signatures")

	var seenSingle, seenMulti bool
	for _, a := range attrs {
		idMap := getAttrMap(t, a, otelEntityId)
		if v, ok := idMap.Get(swDiscoveryDbAddress); ok {
			if v.Str() == "mongo-a" {
				seenSingle = true
			}
			if v.Str() == "mongo-b" {
				seenMulti = true
			}
		}
	}
	assert.True(t, seenSingle, "single port event missing")
	assert.True(t, seenMulti, "multi port event missing")
}

func TestImageDiscovery_MultipleServicesOverlapHeuristic(t *testing.T) {
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		Interval: time.Second,
		Database: &DatabaseDiscoveryConfig{
			ImageRules: []*ImageRule{{DatabaseType: "pg", Patterns: []string{".*/postgres:.*"}, DefaultPort: 5432}}}}
	require.NoError(t, cfg.Validate())

	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pg-0", Namespace: "db", Labels: map[string]string{"app": "pg"}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "postgres", Image: "docker.io/library/postgres:16", Ports: []corev1.ContainerPort{{ContainerPort: 5432}, {ContainerPort: 6432}}}}}}
	serviceA := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "pg-primary", Namespace: "db"}, Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "pg"}, Ports: []corev1.ServicePort{{Port: 5432}}}}
	serviceB := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "pg-broad", Namespace: "db"}, Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "pg"}, Ports: []corev1.ServicePort{{Port: 5433}, {Port: 6432}}}}

	logs := runImageCycle(t, cfg, pod, serviceA, serviceB)
	attrs := collectDBAttrs(logs, "pg")
	require.Len(t, attrs, 1)
	idMap := getAttrMap(t, attrs[0], otelEntityId)
	checkAttr(t, idMap, swDiscoveryDbAddress, "pg-primary")
	entityMap := getAttrMap(t, attrs[0], otelEntityAttributes)
	checkAttrInt(t, entityMap, swDiscoveryDbPort, 5432)
}

func TestImageDiscovery_TargetPortNamedResolution(t *testing.T) {
	cfg := &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		Interval: time.Second,
		Database: &DatabaseDiscoveryConfig{
			ImageRules: []*ImageRule{{DatabaseType: "mysql", Patterns: []string{".*/mysql:.*"}, DefaultPort: 3306}}}}
	require.NoError(t, cfg.Validate())

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "mysql-0", Namespace: "db", Labels: map[string]string{"app": "mysql"}},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "mysql", Image: "docker.io/library/mysql:8",
				Ports: []corev1.ContainerPort{{ContainerPort: 3306, Name: "db"}}}}},
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "mysql-svc", Namespace: "db"},
		Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "mysql"},
			Ports: []corev1.ServicePort{{Port: 13306, TargetPort: intstr.FromString("db")}}},
	}

	logs := runImageCycle(t, cfg, pod, service)
	attrs := collectDBAttrs(logs, "mysql")
	require.Len(t, attrs, 1)

	idMap := getAttrMap(t, attrs[0], otelEntityId)
	checkAttr(t, idMap, swDiscoveryDbAddress, "mysql-svc")
	entityMap := getAttrMap(t, attrs[0], otelEntityAttributes)
	checkAttrInt(t, entityMap, swDiscoveryDbPort, 13306)
}
