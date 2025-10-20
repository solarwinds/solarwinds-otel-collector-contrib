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

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// TestLookupWorkloadKindByNameAndNamespace tests the basic lookup functionality
func TestLookupWorkloadKindByNameAndNamespace(t *testing.T) {
	logger := zap.NewNop()

	// Create a test service
	testService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
	}

	serviceInformer := createInformerWithObjects(t, testService)

	informers := map[string]cache.SharedIndexInformer{
		ServicesWorkloadType: serviceInformer,
	}

	// Test successful lookup with both name and namespace
	result := LookupWorkloadKindByNameAndNamespace(
		"test-service",
		"test-namespace",
		[]string{ServicesWorkloadType},
		logger,
		informers,
		false,
	)

	assert.Equal(t, "test-service", result.Name)
	assert.Equal(t, "test-namespace", result.Namespace)
	assert.Equal(t, ServiceKind, result.Kind)
}

// TestLookupWorkloadKindByNameAndNamespaceNotFound tests that lookup returns nil when not found
func TestLookupWorkloadKindByNameAndNamespaceNotFound(t *testing.T) {
	logger := zap.NewNop()

	serviceInformer := createInformerWithObjects[*corev1.Service](t)

	informers := map[string]cache.SharedIndexInformer{
		ServicesWorkloadType: serviceInformer,
	}

	// Test lookup with non-existent service
	result := LookupWorkloadKindByNameAndNamespace(
		"non-existent-service",
		"test-namespace",
		[]string{ServicesWorkloadType},
		logger,
		informers,
		false,
	)

	assert.Equal(t, "", result.Name, "Should not find non-existent service")
	assert.Equal(t, "", result.Namespace, "Should not find non-existent service")
	assert.Equal(t, "", result.Kind, "Should not find non-existent service")
}

// TestLookupWorkloadByNameAndNamespace tests the internal lookup function
func TestLookupWorkloadByNameAndNamespace(t *testing.T) {
	logger := zap.NewNop()

	// Create a test service
	testService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
	}

	serviceInformer := createInformerWithObjects(t, testService)

	informers := map[string]cache.SharedIndexInformer{
		ServicesWorkloadType: serviceInformer,
	}

	// Test successful lookup
	result := lookupWorkloadByNameAndNamespace(
		"test-service",
		"test-namespace",
		[]string{ServicesWorkloadType},
		logger,
		informers,
	)

	assert.NotNil(t, result, "Should find the service")

	// Cast result to proper type to check fields
	service, ok := result.(*corev1.Service)
	assert.True(t, ok, "Result should be a Service")
	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "test-namespace", service.Namespace)

	// Test lookup with empty namespace (should fail without cross-namespace fallback)
	result2 := lookupWorkloadByNameAndNamespace(
		"test-service",
		"",
		[]string{ServicesWorkloadType},
		logger,
		informers,
	)

	assert.Nil(t, result2, "Should not find service with empty namespace")
}

func TestLookupWorkloadKindByHostname(t *testing.T) {
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "ReplicaSet",
					Name: "test-rs",
				},
			},
		},
	}

	testPodWithoutParent := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-without-parent",
			Namespace: "test-namespace",
		},
	}

	testRS := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rs",
			Namespace: "test-namespace",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					Name: "test-deployment",
				},
			},
		},
	}

	testService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
	}
	testServiceWithDots := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my.test.service.with.dots",
			Namespace: "test-namespace",
		},
	}

	tests := []struct {
		name              string
		hostname          string
		namespaceFromAttr string
		preferPodOwner    bool
		expectedResult    LookupResult
	}{
		{
			name:              "full Pod DNS name",
			hostname:          "test-pod.test-namespace.pod.cluster.local",
			namespaceFromAttr: "",
			expectedResult: LookupResult{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Kind:      PodKind,
			},
		},
		{
			name:              "full Pod DNS name, prefer pod owner",
			hostname:          "test-pod.test-namespace.pod.cluster.local",
			namespaceFromAttr: "",
			preferPodOwner:    true,
			expectedResult: LookupResult{
				Name:      "test-deployment",
				Namespace: "test-namespace",
				Kind:      "Deployment",
			},
		},
		{
			name:              "full Pod DNS name ending with dot",
			hostname:          "test-pod.test-namespace.pod.cluster.local.",
			namespaceFromAttr: "",
			expectedResult: LookupResult{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Kind:      PodKind,
			},
		},
		{
			name:              "full Pod DNS name ending with dot, prefer pod owner",
			hostname:          "test-pod.test-namespace.pod.cluster.local.",
			namespaceFromAttr: "",
			preferPodOwner:    true,
			expectedResult: LookupResult{
				Name:      "test-deployment",
				Namespace: "test-namespace",
				Kind:      "Deployment",
			},
		},
		{
			name:              "short Pod DNS name",
			hostname:          "test-pod.test-namespace.pod",
			namespaceFromAttr: "",
			expectedResult: LookupResult{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Kind:      PodKind,
			},
		},
		{
			name:              "short Pod DNS name, prefer pod owner",
			hostname:          "test-pod.test-namespace.pod",
			namespaceFromAttr: "",
			preferPodOwner:    true,
			expectedResult: LookupResult{
				Name:      "test-deployment",
				Namespace: "test-namespace",
				Kind:      "Deployment",
			},
		},
		{
			name:              "full Service DNS name",
			hostname:          "test-service.test-namespace.svc.cluster.local",
			namespaceFromAttr: "",
			expectedResult: LookupResult{
				Name:      "test-service",
				Namespace: "test-namespace",
				Kind:      ServiceKind,
			},
		},
		{
			name:              "full Service DNS name ending with dot",
			hostname:          "test-service.test-namespace.svc.cluster.local.",
			namespaceFromAttr: "",
			expectedResult: LookupResult{
				Name:      "test-service",
				Namespace: "test-namespace",
				Kind:      ServiceKind,
			},
		},
		{
			name:              "short Service DNS name",
			hostname:          "test-service.test-namespace.svc",
			namespaceFromAttr: "",
			expectedResult: LookupResult{
				Name:      "test-service",
				Namespace: "test-namespace",
				Kind:      ServiceKind,
			},
		},
		{
			name:              "empty host name",
			hostname:          "",
			namespaceFromAttr: "",
			expectedResult:    EmptyLookupResult,
		},
		{
			name:              "Pod name and namespace",
			hostname:          "test-pod.test-namespace",
			namespaceFromAttr: "",
			expectedResult: LookupResult{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Kind:      PodKind,
			},
		},
		{
			name:              "Pod name and namespace, prefer pod owner",
			hostname:          "test-pod.test-namespace",
			namespaceFromAttr: "",
			preferPodOwner:    true,
			expectedResult: LookupResult{
				Name:      "test-deployment",
				Namespace: "test-namespace",
				Kind:      "Deployment",
			},
		},
		{
			name:              "Pod name without namespace, fallback to namespace attribute",
			hostname:          "test-pod",
			namespaceFromAttr: "test-namespace",
			expectedResult: LookupResult{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Kind:      PodKind,
			},
		},
		{
			name:              "Pod name without namespace, fallback to namespace attribute, prefer pod owner",
			hostname:          "test-pod",
			namespaceFromAttr: "test-namespace",
			preferPodOwner:    true,
			expectedResult: LookupResult{
				Name:      "test-deployment",
				Namespace: "test-namespace",
				Kind:      "Deployment",
			},
		},
		{
			name:              "Pod name and namespace, ignore namespace attribute",
			hostname:          "test-pod.test-namespace",
			namespaceFromAttr: "other-namespace",
			expectedResult: LookupResult{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Kind:      PodKind,
			},
		},
		{
			name:              "Pod name and namespace, ignore namespace attribute, prefer pod owner",
			hostname:          "test-pod.test-namespace",
			namespaceFromAttr: "other-namespace",
			preferPodOwner:    true,
			expectedResult: LookupResult{
				Name:      "test-deployment",
				Namespace: "test-namespace",
				Kind:      "Deployment",
			},
		},
		{
			name:              "Service name and namespace",
			hostname:          "test-service.test-namespace",
			namespaceFromAttr: "",
			expectedResult: LookupResult{
				Name:      "test-service",
				Namespace: "test-namespace",
				Kind:      ServiceKind,
			},
		},
		{
			name:              "Service name without namespace, fallback to namespace attribute",
			hostname:          "test-service",
			namespaceFromAttr: "test-namespace",
			expectedResult: LookupResult{
				Name:      "test-service",
				Namespace: "test-namespace",
				Kind:      ServiceKind,
			},
		},
		{
			name:              "Service name and namespace, ignore namespace attribute",
			hostname:          "test-service.test-namespace",
			namespaceFromAttr: "other-namespace",
			expectedResult: LookupResult{
				Name:      "test-service",
				Namespace: "test-namespace",
				Kind:      ServiceKind,
			},
		},
		{
			name:              "invalid Pod address - missing namespace",
			hostname:          "test-pod.pod.cluster.local",
			namespaceFromAttr: "",
			expectedResult:    EmptyLookupResult,
		},
		{
			name:              "invalid Pod address - empty namespace",
			hostname:          "test-pod..pod.cluster.local",
			namespaceFromAttr: "",
			expectedResult:    EmptyLookupResult,
		},
		{
			name:              "invalid Pod address - empty name",
			hostname:          ".test-namespace.pod.cluster.local",
			namespaceFromAttr: "",
			expectedResult:    EmptyLookupResult,
		},
		{
			name:              "invalid Service address - missing namespace",
			hostname:          "test-service.svc.cluster.local",
			namespaceFromAttr: "",
			expectedResult:    EmptyLookupResult,
		},
		{
			name:              "invalid Service address - empty namespace",
			hostname:          "test-service..svc.cluster.local",
			namespaceFromAttr: "",
			expectedResult:    EmptyLookupResult,
		},
		{
			name:              "invalid Service address - empty name",
			hostname:          ".test-namespace.svc.cluster.local",
			namespaceFromAttr: "",
			expectedResult:    EmptyLookupResult,
		},
		{
			name:              "Service name with dots",
			hostname:          "my.test.service.with.dots",
			namespaceFromAttr: "test-namespace",
			expectedResult: LookupResult{
				Name:      "my.test.service.with.dots",
				Namespace: "test-namespace",
				Kind:      ServiceKind,
			},
		},
		{
			name:     "non-existent Pod DNS name",
			hostname: "unknown-pod.test-namespace.pod.cluster.local",
			expectedResult: LookupResult{
				Name:      "unknown-pod",
				Namespace: "test-namespace",
				Kind:      PodKind,
			},
		},
		{
			name:     "non-existent Service DNS name",
			hostname: "unknown-service.test-namespace.svc.cluster.local",
			expectedResult: LookupResult{
				Name:      "unknown-service",
				Namespace: "test-namespace",
				Kind:      ServiceKind,
			},
		},
		{
			name:           "non-existent Pod name and namespace",
			hostname:       "unknown-pod.test-namespace",
			expectedResult: EmptyLookupResult,
		},
		{
			name:           "non-existent Service name and namespace",
			hostname:       "unknown-service.test-namespace",
			expectedResult: EmptyLookupResult,
		},
		{
			name:           "Pod name and namespace, prefer pod owner, non-existent parent",
			hostname:       "test-pod-without-parent.test-namespace",
			preferPodOwner: true,
			expectedResult: LookupResult{
				Name:      "test-pod-without-parent",
				Namespace: "test-namespace",
				Kind:      PodKind,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podInformer := createInformerWithObjects(t, testPod, testPodWithoutParent)
			rsInformer := createInformerWithObjects(t, testRS)
			serviceInformer := createInformerWithObjects(t, testService, testServiceWithDots)

			informers := map[string]cache.SharedIndexInformer{
				PodsWorkloadType:        podInformer,
				ReplicaSetsWorkloadType: rsInformer,
				ServicesWorkloadType:    serviceInformer,
			}

			result := LookupWorkloadKindByHostname(
				tt.hostname,
				tt.namespaceFromAttr,
				[]string{PodsWorkloadType, ReplicaSetsWorkloadType, ServicesWorkloadType},
				zap.NewNop(),
				informers,
				tt.preferPodOwner,
			)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func createInformerWithObjects[T runtime.Object](t *testing.T, objects ...T) cache.SharedIndexInformer {
	t.Helper()
	var exampleObject T

	informer := cache.NewSharedIndexInformer(
		nil, // ListerWatcher is nil for testing
		exampleObject,
		0, // resyncPeriod
		cache.Indexers{},
	)
	for _, obj := range objects {
		err := informer.GetStore().Add(obj)
		assert.NoError(t, err)
	}
	return informer
}
